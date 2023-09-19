package terraformpullrequest

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	controller "github.com/padok-team/burrito/internal/controllers/terraformlayer"
)

func (r *Reconciler) getAffectedLayers(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]configv1alpha1.TerraformLayer, error) {
	var layers configv1alpha1.TerraformLayerList
	err := r.Client.List(context.Background(), &layers)
	if err != nil {
		return nil, err
	}
	var provider Provider
	for _, p := range r.Providers {
		if p.IsFromProvider(pr) {
			provider = p
			break
		}
	}
	if provider == nil {
		return nil, fmt.Errorf("could not find provider for pull request %s", pr.Name)
	}
	changes, err := provider.GetChanges(repository, pr)
	if err != nil {
		return nil, err
	}
	affectedLayers := []configv1alpha1.TerraformLayer{}
	for _, layer := range layers.Items {
		if isLayerAffected(layer, *pr, changes) {
			affectedLayers = append(affectedLayers, layer)
		}
	}

	return affectedLayers, nil
}

func isLayerAffected(layer configv1alpha1.TerraformLayer, pr configv1alpha1.TerraformPullRequest, changes []string) bool {
	if layer.Spec.Repository.Name != pr.Spec.Repository.Name {
		return false
	}
	if layer.Spec.Repository.Namespace != pr.Spec.Repository.Namespace {
		return false
	}
	if layer.Spec.Branch != pr.Spec.Base {
		return false
	}
	if controller.LayerFilesHaveChanged(layer, changes) {
		return true
	}
	return false
}

func generateTempLayers(pr *configv1alpha1.TerraformPullRequest, layers []configv1alpha1.TerraformLayer) []configv1alpha1.TerraformLayer {
	list := []configv1alpha1.TerraformLayer{}
	for _, layer := range layers {
		new := configv1alpha1.TerraformLayer{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    layer.ObjectMeta.Namespace,
				GenerateName: fmt.Sprintf("%s-%s-", layer.Name, pr.Spec.ID),
				Annotations: map[string]string{
					annotations.LastBranchCommit:   pr.Annotations[annotations.LastBranchCommit],
					annotations.LastRelevantCommit: pr.Annotations[annotations.LastBranchCommit],
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: pr.GetAPIVersion(),
						Kind:       pr.GetKind(),
						Name:       pr.Name,
						UID:        pr.UID,
					},
				},
				Labels: map[string]string{
					"burrito/managed-by": pr.Name,
				},
			},
			Spec: configv1alpha1.TerraformLayerSpec{
				Path:            layer.Spec.Path,
				Branch:          pr.Spec.Branch,
				TerraformConfig: layer.Spec.TerraformConfig,
				Repository:      layer.Spec.Repository,
				RemediationStrategy: configv1alpha1.RemediationStrategy{
					AutoApply: false,
				},
				OverrideRunnerSpec: layer.Spec.OverrideRunnerSpec,
			},
		}
		list = append(list, new)
	}
	return list
}
