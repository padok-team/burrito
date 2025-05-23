package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func GetDefaultLabels(layer *configv1alpha1.TerraformLayer) map[string]string {
	return map[string]string{
		"burrito/managed-by": layer.Name,
	}
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
	labelSelector := labels.NewSelector()
	for key, value := range GetDefaultLabels(layer) {
		requirement, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return []*configv1alpha1.TerraformRun{}, err
		}
		labelSelector = labelSelector.Add(*requirement)
	}
	err := r.Client.List(
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
