package gitlab

import (
	"fmt"
	"strconv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	log "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIProvider struct {
	client *gitlab.Client
}

func (api *APIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return []string{}, err
	}
	listOpts := gitlab.ListMergeRequestDiffsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
		},
	}
	var changes []string
	for {
		diffs, resp, err := api.client.MergeRequests.ListMergeRequestDiffs(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &listOpts)
		if err != nil {
			log.Errorf("Error while getting merge request changes: %s", err)
			return []string{}, err
		}
		for _, change := range diffs {
			changes = append(changes, change.NewPath)
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return changes, nil
}

func (api *APIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	commit := s.Commit
	if commit == "" {
		commit = pr.Annotations[annotations.LastBranchCommit]
	}
	name := "burrito/" + string(s.Phase)
	if s.Context != "" {
		name = s.Context
	}
	description := s.Description
	state := toGitlabBuildState(s.State)
	opts := &gitlab.SetCommitStatusOptions{
		State:       state,
		Name:        &name,
		Description: &description,
	}
	if s.TargetURL != "" {
		opts.TargetURL = &s.TargetURL
	}
	// Errors are logged by the caller (commitstatus.Post), which has the context to tell a
	// permanent failure from a transient one worth retrying.
	_, _, err := api.client.Commits.SetCommitStatus(getGitlabNamespacedName(repository.Spec.Repository.Url), commit, opts)
	return err
}

func toGitlabBuildState(s status.State) gitlab.BuildStateValue {
	switch s {
	case status.StateRunning:
		return gitlab.Running
	case status.StateSuccess:
		return gitlab.Success
	case status.StateFailure:
		return gitlab.Failed
	default:
		return gitlab.Pending
	}
}

func (api *APIProvider) GetMergeCommit(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) (string, error) {
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return "", err
	}
	mergeRequest, _, err := api.client.MergeRequests.GetMergeRequest(getGitlabNamespacedName(repository.Spec.Repository.Url), id, nil)
	if err != nil {
		return "", err
	}
	if mergeRequest.MergeCommitSHA != "" {
		return mergeRequest.MergeCommitSHA, nil
	}
	// Squashed merge requests only get a squash_commit_sha, not a merge_commit_sha.
	return mergeRequest.SquashCommitSHA, nil
}

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	body, err := prComment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	body = comment.WithManagedMarker(body)
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return err
	}
	projectName := getGitlabNamespacedName(repository.Spec.Repository.Url)
	managedNoteID, err := api.findManagedNoteID(projectName, id)
	if err != nil {
		return err
	}
	if managedNoteID != 0 {
		_, _, err = api.client.Notes.UpdateMergeRequestNote(projectName, id, managedNoteID, &gitlab.UpdateMergeRequestNoteOptions{
			Body: gitlab.Ptr(body),
		})
		return err
	}
	_, _, err = api.client.Notes.CreateMergeRequestNote(projectName, id, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr(body),
	})
	if err != nil {
		log.Errorf("Error while creating merge request note: %s", err)
		return err
	}
	return nil
}

func (api *APIProvider) findManagedNoteID(projectName string, id int64) (int64, error) {
	opts := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}
	for {
		notes, resp, err := api.client.Notes.ListMergeRequestNotes(projectName, id, opts)
		if err != nil {
			return 0, err
		}
		for _, existingNote := range notes {
			if existingNote != nil && comment.HasManagedMarker(existingNote.Body) {
				return existingNote.ID, nil
			}
		}
		if resp.NextPage == 0 {
			return 0, nil
		}
		opts.Page = resp.NextPage
	}
}

func (api *APIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	state := "opened"
	listOpts := &gitlab.ListProjectMergeRequestsOptions{
		State: &state,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	var pullRequests []configv1alpha1.TerraformPullRequest
	for {
		mergeRequests, resp, err := api.client.MergeRequests.ListProjectMergeRequests(getGitlabNamespacedName(repository.Spec.Repository.Url), listOpts)
		if err != nil {
			return nil, err
		}
		for _, mergeRequest := range mergeRequests {
			if mergeRequest == nil {
				continue
			}
			pullRequests = append(pullRequests, configv1alpha1.TerraformPullRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", repository.Name, mergeRequest.IID),
					Namespace: repository.Namespace,
					Annotations: map[string]string{
						annotations.LastBranchCommit: mergeRequest.SHA,
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: repository.GetAPIVersion(),
							Kind:       repository.GetKind(),
							Name:       repository.Name,
							UID:        repository.UID,
						},
					},
				},
				Spec: configv1alpha1.TerraformPullRequestSpec{
					Branch: mergeRequest.SourceBranch,
					Base:   mergeRequest.TargetBranch,
					ID:     strconv.FormatInt(mergeRequest.IID, 10),
					Repository: configv1alpha1.TerraformLayerRepository{
						Name:      repository.Name,
						Namespace: repository.Namespace,
					},
				},
			})
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return pullRequests, nil
}
