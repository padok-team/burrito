package gitea

import (
	"context"
	"fmt"
	"strconv"

	gitea "gitea.dev/sdk"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/repository/types"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIProvider struct {
	client *gitea.Client
}

func (api *APIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	owner, repoName := parseGiteaUrl(repository.Spec.Repository.Url)
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing Gitea pull request ID: %s", err)
		return []string{}, err
	}
	opts := gitea.ListPullRequestFilesOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 50,
		},
	}
	// Get all pull request files from Gitea
	var allChangedFiles []string
	for {
		changedFiles, resp, err := api.client.PullRequests.ListPullRequestFiles(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return []string{}, err
		}
		for _, file := range changedFiles {
			if file.Status != "unchanged" {
				allChangedFiles = append(allChangedFiles, file.Filename)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allChangedFiles, nil
}

func (api *APIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, s types.CommitStatus) error {
	owner, repoName := parseGiteaUrl(repository.Spec.Repository.Url)
	commit := s.Commit
	if commit == "" && pr != nil {
		commit = pr.Annotations[annotations.LastBranchCommit]
	}
	name := "burrito/" + string(s.Phase)
	if s.Context != "" {
		name = s.Context
	}
	opts := gitea.CreateStatusOption{
		State:       toGiteaStatusState(s.State),
		Context:     name,
		Description: s.Description,
	}
	if s.TargetURL != "" {
		opts.TargetURL = s.TargetURL
	}
	// Errors are logged by the caller (commitstatus.Post), which has the context to tell a
	// permanent failure from a transient one worth retrying.
	_, _, err := api.client.Repositories.CreateStatus(context.TODO(), owner, repoName, commit, opts)
	return err
}

// toGiteaStatusState maps our internal types.State to a Gitea-accepted
// value. Gitea only accepts error/failure/pending/success, with no notion
// of "running": fold it into pending, like the GitHub provider.
func toGiteaStatusState(s types.State) gitea.StatusState {
	switch s {
	case types.StateSuccess:
		return gitea.StatusSuccess
	case types.StateFailure:
		return gitea.StatusFailure
	default:
		return gitea.StatusPending
	}
}

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	body, err := prComment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	body = comment.WithManagedMarker(body)
	owner, repoName := parseGiteaUrl(repository.Spec.Repository.Url)
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing Gitea pull request ID: %s", err)
		return err
	}
	managedCommentID, err := api.findManagedCommentID(owner, repoName, id)
	if err != nil {
		return err
	}
	if managedCommentID != 0 {
		_, _, err = api.client.Issues.EditIssueComment(context.TODO(), owner, repoName, managedCommentID, gitea.EditIssueCommentOption{
			Body: body,
		})
		return err
	}
	_, _, err = api.client.Issues.CreateIssueComment(context.TODO(), owner, repoName, id, gitea.CreateIssueCommentOption{
		Body: body,
	})
	return err
}

func (api *APIProvider) findManagedCommentID(owner string, repoName string, id int64) (int64, error) {
	opts := gitea.ListIssueCommentOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 50,
		},
	}
	for {
		comments, resp, err := api.client.Issues.ListIssueComments(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return 0, err
		}
		for _, existingComment := range comments {
			if existingComment != nil && comment.HasManagedMarker(existingComment.Body) {
				return existingComment.ID, nil
			}
		}
		if resp.NextPage == 0 {
			return 0, nil
		}
		opts.Page = resp.NextPage
	}
}

func (api *APIProvider) GetPullRequestTitle(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) (string, error) {
	owner, repoName := parseGiteaUrl(repository.Spec.Repository.Url)
	id, err := strconv.ParseInt(pr.Spec.ID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("could not parse Gitea pull request ID: %w", err)
	}
	pullRequest, _, err := api.client.PullRequests.GetPullRequest(context.TODO(), owner, repoName, id)
	if err != nil {
		return "", err
	}
	return pullRequest.Title, nil
}

func (api *APIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	owner, repoName := parseGiteaUrl(repository.Spec.Repository.Url)
	opts := gitea.ListPullRequestsOptions{
		State: gitea.StateOpen,
		ListOptions: gitea.ListOptions{
			PageSize: 50,
		},
	}

	var pullRequests []configv1alpha1.TerraformPullRequest
	for {
		openPullRequests, resp, err := api.client.PullRequests.ListRepoPullRequests(context.TODO(), owner, repoName, opts)
		if err != nil {
			return nil, err
		}
		for _, pullRequest := range openPullRequests {
			if pullRequest == nil || pullRequest.Head == nil || pullRequest.Base == nil {
				continue
			}
			pullRequests = append(pullRequests, configv1alpha1.TerraformPullRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", repository.Name, pullRequest.Index),
					Namespace: repository.Namespace,
					Annotations: map[string]string{
						annotations.LastBranchCommit: pullRequest.Head.Sha,
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
					Branch: pullRequest.Head.Ref,
					Base:   pullRequest.Base.Ref,
					ID:     strconv.FormatInt(pullRequest.Index, 10),
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
		opts.Page = resp.NextPage
	}

	return pullRequests, nil
}
