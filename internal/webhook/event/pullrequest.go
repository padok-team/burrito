package event

import (
	"context"
	"fmt"

	utils "github.com/padok-team/burrito/internal/utils/url"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hashicorp/go-multierror"
)

type PullRequestEvent struct {
	URL       string
	Reference string
	Base      string
	Action    string
	ID        string
	Commit    string
}

func (e *PullRequestEvent) Handle(c client.Client) error {
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := c.List(context.Background(), repositories)
	if err != nil {
		log.Errorf("could not list TerraformRepositories: %s", err)
		return err
	}
	affectedRepositories := e.getAffectedRepositories(repositories.Items)
	if len(affectedRepositories) == 0 {
		log.Infof("no affected repositories found for pull request event")
		return nil
	}

	prs := e.generateTerraformPullRequests(affectedRepositories)
	switch e.Action {
	case PullRequestOpened:
		return batchCreatePullRequests(context.TODO(), c, prs)
	case PullRequestClosed:
		// remove annotation from affected repositories
		for _, repo := range affectedRepositories {
			key := annotations.ComputeKeyForSyncBranchNow(e.Reference)
			err := annotations.Remove(context.TODO(), c, &repo, key)
			if err != nil {
				log.Errorf("could not remove annotation to TerraformRepository %s", err)
			}
		}
		return batchDeletePullRequests(context.TODO(), c, prs)
	default:
		log.Infof("action %s not supported", e.Action)
	}
	return nil
}

func batchCreatePullRequests(ctx context.Context, c client.Client, prs []configv1alpha1.TerraformPullRequest) error {
	var errResult error
	for _, pr := range prs {
		log.Infof("creating TerraformPullRequest %s/%s", pr.Namespace, pr.Name)
		err := c.Create(ctx, &pr)
		if err != nil {
			errResult = multierror.Append(errResult, err)
		}
	}
	return errResult
}

func batchDeletePullRequests(ctx context.Context, c client.Client, prs []configv1alpha1.TerraformPullRequest) error {
	var errResult error
	for _, pr := range prs {
		log.Infof("deleting TerraformPullRequest %s/%s", pr.Namespace, pr.Name)
		err := c.Delete(ctx, &pr)
		if err != nil {
			errResult = multierror.Append(errResult, err)
		}
	}
	return errResult
}

func (e *PullRequestEvent) generateTerraformPullRequests(repositories []configv1alpha1.TerraformRepository) []configv1alpha1.TerraformPullRequest {
	prs := []configv1alpha1.TerraformPullRequest{}
	for _, repository := range repositories {
		pr := configv1alpha1.TerraformPullRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", repository.Name, e.ID),
				Namespace: repository.Namespace,
				Annotations: map[string]string{
					annotations.LastBranchCommit: e.Commit,
				},
			},
			Spec: configv1alpha1.TerraformPullRequestSpec{
				Branch: e.Reference,
				ID:     e.ID,
				Base:   e.Base,
				Repository: configv1alpha1.TerraformLayerRepository{
					Name:      repository.Name,
					Namespace: repository.Namespace,
				},
			},
		}
		prs = append(prs, pr)
	}
	return prs
}

// Function that checks if any TerraformRepository is linked to a PullRequestEvent
func (e *PullRequestEvent) getAffectedRepositories(repositories []configv1alpha1.TerraformRepository) []configv1alpha1.TerraformRepository {
	affectedRepositories := []configv1alpha1.TerraformRepository{}
	for _, repo := range repositories {
		log.Infof("evaluating TerraformRepository %s for url %s", repo.Name, repo.Spec.Repository.Url)
		log.Infof("comparing normalized url %s with received URL from payload %s", utils.NormalizeUrl(repo.Spec.Repository.Url), e.URL)
		if e.URL == utils.NormalizeUrl(repo.Spec.Repository.Url) {
			affectedRepositories = append(affectedRepositories, repo)
		}
	}
	return affectedRepositories
}
