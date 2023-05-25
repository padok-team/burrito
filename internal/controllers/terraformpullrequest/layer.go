package terraformpullrequest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
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
	if layerFilesHaveChanged(layer, changes) {
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
						APIVersion: pr.APIVersion,
						Kind:       pr.Kind,
						Name:       pr.Name,
						UID:        pr.UID,
					},
				},
				Labels: map[string]string{
					"burrito/managed-by": pr.Name,
				},
			},
			Spec: configv1alpha1.TerraformLayerSpec{
				Path:                layer.Spec.Path,
				Branch:              pr.Spec.Branch,
				TerraformConfig:     layer.Spec.TerraformConfig,
				Repository:          layer.Spec.Repository,
				RemediationStrategy: "dry",
				OverrideRunnerSpec:  layer.Spec.OverrideRunnerSpec,
			},
		}
		list = append(list, new)
	}
	return list
}

func layerFilesHaveChanged(layer configv1alpha1.TerraformLayer, changedFiles []string) bool {
	if len(changedFiles) == 0 {
		return true
	}

	// At last one changed file must be under refresh path
	for _, f := range changedFiles {
		f = ensureAbsPath(f)
		if strings.Contains(f, layer.Spec.Path) {
			return true
		}
		// Check if the file is under an additionnal trigger path
		if val, ok := layer.Annotations[annotations.AdditionnalTriggerPaths]; ok {
			for _, p := range strings.Split(val, ",") {
				p = ensureAbsPath(p)
				if strings.Contains(f, p) {
					return true
				}
			}
		}
	}

	return false
}

func ensureAbsPath(input string) string {
	if !filepath.IsAbs(input) {
		return string(filepath.Separator) + input
	}
	return input
}
