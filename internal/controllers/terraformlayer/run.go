package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const managedByLabel = "burrito/managed-by"

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func GetDefaultLabels(layer *configv1alpha1.TerraformLayer) map[string]string {
	return map[string]string{
		managedByLabel: managedByLabelValue(layer),
	}
}

func managedByLabelValue(layer *configv1alpha1.TerraformLayer) string {
	// Kubernetes label values max out at 63 bytes; a stable hash keeps long
	// layer names valid while still linking created runs back to the layer.
	return fmt.Sprintf("layer-%s", utils.ShortHash(fmt.Sprintf("%s/%s", layer.Namespace, layer.Name)))
}

// managedByLabelSelectorValues returns every "burrito/managed-by" value a run
// belonging to this layer may carry: the current hash-based value, plus the
// legacy raw layer name for runs created before that hashing was introduced.
// The legacy value is only included when it's a valid label value, since a
// layer name that isn't (the case this hashing fixes) could never have been
// used as a label before either.
func managedByLabelSelectorValues(layer *configv1alpha1.TerraformLayer) []string {
	values := []string{managedByLabelValue(layer)}
	if len(validation.IsValidLabelValue(layer.Name)) == 0 {
		values = append(values, layer.Name)
	}
	return values
}

func (r *Reconciler) getRun(layer *configv1alpha1.TerraformLayer, revision string, action Action) configv1alpha1.TerraformRun {
	artifact := configv1alpha1.Artifact{}
	if action == ApplyAction {
		run := strings.Split(layer.Annotations[annotations.LastPlanRun], "/")
		artifact.Attempt = run[1]
		artifact.Run = run[0]
	}
	return configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", layer.Name, action),
			Namespace:    layer.Namespace,
			Labels:       GetDefaultLabels(layer),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: layer.GetAPIVersion(),
					Kind:       layer.GetKind(),
					Name:       layer.Name,
					UID:        layer.UID,
				},
			},
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: string(action),
			Layer: configv1alpha1.TerraformRunLayer{
				Name:      layer.Name,
				Namespace: layer.Namespace,
				Revision:  revision,
			},
			Artifact: artifact,
		},
	}
}

func (r *Reconciler) getAllRuns(ctx context.Context, layer *configv1alpha1.TerraformLayer) ([]*configv1alpha1.TerraformRun, error) {
	list := &configv1alpha1.TerraformRunList{}
	requirement, err := labels.NewRequirement(managedByLabel, selection.In, managedByLabelSelectorValues(layer))
	if err != nil {
		return []*configv1alpha1.TerraformRun{}, err
	}
	labelSelector := labels.NewSelector().Add(*requirement)
	err = r.Client.List(
		ctx,
		list,
		client.MatchingLabelsSelector{Selector: labelSelector},
		&client.ListOptions{Namespace: layer.Namespace},
	)
	if err != nil {
		return []*configv1alpha1.TerraformRun{}, err
	}

	// Keep only runs with state Succeeded or Failed
	var runs []*configv1alpha1.TerraformRun
	for _, run := range list.Items {
		runs = append(runs, &run)
	}
	return runs, nil
}

func deleteAll(ctx context.Context, c client.Client, objs []*configv1alpha1.TerraformRun) error {
	for _, obj := range objs {
		if err := c.Delete(ctx, obj); err != nil {
			return err
		}
	}
	return nil
}
