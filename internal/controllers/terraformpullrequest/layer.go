package terraformpullrequest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	controller "github.com/padok-team/burrito/internal/controllers/terraformlayer"
	logrus "github.com/sirupsen/logrus"
)

const (
	managedByLabel              = "burrito/managed-by"
	maxGenerateNamePrefixLength = 240
)

func (r *Reconciler) getAffectedLayers(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]configv1alpha1.TerraformLayer, error) {
	var layers configv1alpha1.TerraformLayerList
	err := r.Client.List(context.Background(), &layers)
	if err != nil {
		return nil, err
	}

	provider, err := r.getAPIProvider(repository)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Provider error", "Failed to get API provider for get changes from pull request")
		return nil, err
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
	// Check if branch matches OR if PR base is in additionalTargetRefs
	branchMatches := layer.Spec.Branch == pr.Spec.Base
	additionalTargetMatches := slices.Contains(layer.Spec.AdditionalTargetRefs, pr.Spec.Base)

	// If neither branch matches nor is in additional targets, skip this layer
	if !branchMatches && !additionalTargetMatches {
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
				GenerateName: tempLayerGenerateName(layer.Name, pr),
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
					managedByLabel: managedByLabelValue(pr),
				},
			},
			Spec: configv1alpha1.TerraformLayerSpec{
				Path:             layer.Spec.Path,
				Branch:           pr.Spec.Branch,
				TerraformConfig:  layer.Spec.TerraformConfig,
				TerragruntConfig: layer.Spec.TerragruntConfig,
				OpenTofuConfig:   layer.Spec.OpenTofuConfig,
				Repository:       layer.Spec.Repository,
				RemediationStrategy: configv1alpha1.RemediationStrategy{
					AutoApply: &[]bool{false}[0],
				},
				RunHistoryPolicy:   layer.Spec.RunHistoryPolicy,
				OverrideRunnerSpec: layer.Spec.OverrideRunnerSpec,
			},
		}
		list = append(list, new)
	}
	return list
}

// GetLinkedLayers looks up temp layers by both the current hashed label value and the
// legacy raw PR name, so layers created before the hash-based labeling was introduced
// are still found.
func GetLinkedLayers(cl client.Client, pr *configv1alpha1.TerraformPullRequest) ([]configv1alpha1.TerraformLayer, error) {
	return getLinkedLayersForLabels(cl, pr, managedByLabelValue(pr), pr.Name)
}

func getLinkedLayersForLabels(cl client.Client, pr *configv1alpha1.TerraformPullRequest, labelValues ...string) ([]configv1alpha1.TerraformLayer, error) {
	linkedLayers := []configv1alpha1.TerraformLayer{}
	seen := map[string]struct{}{}
	for _, labelValue := range labelValues {
		layers, err := getLinkedLayersForLabel(cl, labelValue)
		if err != nil {
			return nil, err
		}
		for _, layer := range layers {
			key := fmt.Sprintf("%s/%s", layer.Namespace, layer.Name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			linkedLayers = append(linkedLayers, layer)
		}
	}
	return linkedLayers, nil
}

func getLinkedLayersForLabel(cl client.Client, labelValue string) ([]configv1alpha1.TerraformLayer, error) {
	layers := configv1alpha1.TerraformLayerList{}
	requirement, err := labels.NewRequirement(managedByLabel, selection.Equals, []string{labelValue})
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
	logrus.Infof("deleting temporary layers for pull request %s/%s", pr.Namespace, pr.Name)
	if err := r.deleteTempLayersByLabel(ctx, pr, managedByLabelValue(pr)); err != nil {
		return err
	}
	// Keep deleting the legacy label value so upgrades clean up old temp layers.
	return r.deleteTempLayersByLabel(ctx, pr, pr.Name)
}

func (r *Reconciler) deleteTempLayersByLabel(ctx context.Context, pr *configv1alpha1.TerraformPullRequest, labelValue string) error {
	return r.Client.DeleteAllOf(
		ctx, &configv1alpha1.TerraformLayer{}, client.InNamespace(pr.Namespace), client.MatchingLabels{managedByLabel: labelValue},
	)
}

func tempLayerGenerateName(layerName string, pr *configv1alpha1.TerraformPullRequest) string {
	hash := shortHash(fmt.Sprintf("%s/%s/%s", pr.Namespace, pr.Name, pr.Spec.ID))
	prefix := fmt.Sprintf("%s-pr-%s-", layerName, hash)
	if len(prefix) <= maxGenerateNamePrefixLength {
		return prefix
	}
	return prefix[:maxGenerateNamePrefixLength]
}

func managedByLabelValue(pr *configv1alpha1.TerraformPullRequest) string {
	// Kubernetes label values max out at 63 chars; a stable hash keeps long PR
	// names valid while still linking generated layers back to the same PR.
	return fmt.Sprintf("pr-%s", shortHash(fmt.Sprintf("%s/%s", pr.Namespace, pr.Name)))
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:16]
}
