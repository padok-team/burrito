package event

import (
	"context"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	controller "github.com/padok-team/burrito/internal/controllers/terraformlayer"
	utils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PushEvent struct {
	URL       string
	Reference string
	ChangeInfo
	Changes []string
}

func (e *PushEvent) Handle(c client.Client) error {
	date := time.Now().Format(time.UnixDate)
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := c.List(context.Background(), repositories)
	if err != nil {
		log.Errorf("could not list TerraformRepositories: %s", err)
		return err
	}
	layers := &configv1alpha1.TerraformLayerList{}
	err = c.List(context.Background(), layers)
	if err != nil {
		log.Errorf("could not list TerraformLayers: %s", err)
		return err
	}
	prs := &configv1alpha1.TerraformPullRequestList{}
	err = c.List(context.Background(), prs)
	if err != nil {
		log.Errorf("could not list TerraformPullRequests: %s", err)
		return err
	}
	affectedRepositories := e.getAffectedRepositories(repositories.Items)
	for _, repo := range affectedRepositories {
		ann := map[string]string{}
		ann[annotations.ComputeKeyForSyncBranchNow(e.Reference)] = time.Now().Format(time.UnixDate)
		err := annotations.Add(context.TODO(), c, &repo, ann)
		if err != nil {
			log.Errorf("could not add annotation to TerraformRepository %s", err)
			return err
		}
	}

	// TODO: Remove this loop once the repo controller implements the same behavior
	for _, layer := range e.getAffectedLayers(layers.Items, affectedRepositories) {
		ann := map[string]string{}
		log.Printf("evaluating TerraformLayer %s for revision %s", layer.Name, e.Reference)
		if layer.Spec.Branch != e.Reference {
			log.Infof("branch %s for TerraformLayer %s not matching revision %s", layer.Spec.Branch, layer.Name, e.Reference)
			continue
		}
		ann[annotations.LastBranchCommit] = e.ChangeInfo.ShaAfter
		ann[annotations.LastBranchCommitDate] = date

		if controller.LayerFilesHaveChanged(layer, e.Changes) {
			log.Infof("layer %s is affected by push event", layer.Name)
			ann[annotations.LastRelevantCommit] = e.ChangeInfo.ShaAfter
			ann[annotations.LastRelevantCommitDate] = date
		}

		err := annotations.Add(context.TODO(), c, &layer, ann)
		if err != nil {
			log.Errorf("could not add annotation to TerraformLayer %s", err)
			return err
		}
	}

	// TODO: Remove this loop once the repo controller implements the same behavior
	for _, pr := range e.getAffectedPullRequests(prs.Items, affectedRepositories) {
		ann := map[string]string{}
		ann[annotations.LastBranchCommit] = e.ChangeInfo.ShaAfter
		ann[annotations.LastBranchCommitDate] = date
		err := annotations.Add(context.TODO(), c, &pr, ann)
		if err != nil {
			log.Errorf("could not add annotation to TerraformPullRequest %s", err)
			return err
		}
	}
	return nil
}

func (e *PushEvent) getAffectedRepositories(repositories []configv1alpha1.TerraformRepository) []configv1alpha1.TerraformRepository {
	affectedRepositories := []configv1alpha1.TerraformRepository{}
	log.Infof("looking for affected repositories, event url: %s", e.URL)
	for _, repo := range repositories {
		if e.URL == utils.NormalizeUrl(repo.Spec.Repository.Url) {
			affectedRepositories = append(affectedRepositories, repo)
			continue
		}
	}
	return affectedRepositories
}

func (e *PushEvent) getAffectedLayers(allLayers []configv1alpha1.TerraformLayer, affectedRepositories []configv1alpha1.TerraformRepository) []configv1alpha1.TerraformLayer {
	layers := []configv1alpha1.TerraformLayer{}
	for _, layer := range allLayers {
		if isLayerLinkedToAnyRepositories(affectedRepositories, layer) {
			layers = append(layers, layer)
		}
	}
	return layers
}

func (e *PushEvent) getAffectedPullRequests(prs []configv1alpha1.TerraformPullRequest, affectedRepositories []configv1alpha1.TerraformRepository) []configv1alpha1.TerraformPullRequest {
	affectedPRs := []configv1alpha1.TerraformPullRequest{}
	for _, pr := range prs {
		if isPRLinkedToAnyRepositories(pr, affectedRepositories) && pr.Spec.Branch == e.Reference {
			affectedPRs = append(affectedPRs, pr)
		}
	}
	return affectedPRs
}
