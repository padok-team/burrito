package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v84/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIProvider struct {
	client *github.Client
}

func (api *APIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return []string{}, err
	}
	// Per page is 30 by default, max is 100
	opts := &github.ListOptions{
		PerPage: 100,
	}
	// Get all pull request files from Github
	var allChangedFiles []string
	for {
		changedFiles, resp, err := api.client.PullRequests.ListFiles(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return []string{}, err
		}
		for _, file := range changedFiles {
			if *file.Status != "unchanged" {
				allChangedFiles = append(allChangedFiles, *file.Filename)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allChangedFiles, nil
}

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	body, err := prComment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	body = comment.WithManagedMarker(body)
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return err
	}
	managedCommentID, err := api.findManagedCommentID(owner, repoName, id)
	if err != nil {
		return err
	}
	if managedCommentID != 0 {
		_, _, err = api.client.Issues.EditComment(context.TODO(), owner, repoName, managedCommentID, &github.IssueComment{
			Body: &body,
		})
		return err
	}
	_, _, err = api.client.Issues.CreateComment(context.TODO(), owner, repoName, id, &github.IssueComment{
		Body: &body,
	})
	return err
}

func (api *APIProvider) findManagedCommentID(owner string, repoName string, id int) (int64, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		comments, resp, err := api.client.Issues.ListComments(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return 0, err
		}
		for _, existingComment := range comments {
			if existingComment != nil && comment.HasManagedMarker(existingComment.GetBody()) {
				return existingComment.GetID(), nil
			}
		}
		if resp.NextPage == 0 {
			return 0, nil
		}
		opts.Page = resp.NextPage
	}
}

func (api *APIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var pullRequests []configv1alpha1.TerraformPullRequest
	for {
		openPullRequests, resp, err := api.client.PullRequests.List(context.Background(), owner, repoName, opts)
		if err != nil {
			return nil, err
		}
		for _, pullRequest := range openPullRequests {
			if pullRequest == nil {
				continue
			}
			pullRequests = append(pullRequests, configv1alpha1.TerraformPullRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", repository.Name, pullRequest.GetNumber()),
					Namespace: repository.Namespace,
					Annotations: map[string]string{
						annotations.LastBranchCommit: pullRequest.GetHead().GetSHA(),
					},
				},
				Spec: configv1alpha1.TerraformPullRequestSpec{
					Branch: pullRequest.GetHead().GetRef(),
					Base:   pullRequest.GetBase().GetRef(),
					ID:     strconv.Itoa(pullRequest.GetNumber()),
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
