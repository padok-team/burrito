package terraformpullrequest

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
				GenerateName: fmt.Sprintf("%s-pr-%s-", layer.Name, pr.Spec.ID),
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

func GetLinkedLayers(cl client.Client, pr *configv1alpha1.TerraformPullRequest) ([]configv1alpha1.TerraformLayer, error) {
	layers := configv1alpha1.TerraformLayerList{}
	requirement, err := labels.NewRequirement("burrito/managed-by", selection.Equals, []string{pr.Name})
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*requirement)
	err = cl.List(context.TODO(), &layers, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	return layers.Items, nil
}

func (r *Reconciler) deleteTempLayers(ctx context.Context, pr *configv1alpha1.TerraformPullRequest) error {
	return r.Client.DeleteAllOf(
		ctx, &configv1alpha1.TerraformLayer{}, client.InNamespace(pr.Namespace), client.MatchingLabels{"burrito/managed-by": pr.Name},
	)
}
