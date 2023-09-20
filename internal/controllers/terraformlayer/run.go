package terraformlayer

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *Reconciler) getRun(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action Action) configv1alpha1.TerraformRun {
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
			},
		},
	}
}
