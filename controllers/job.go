package controllers

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func defaultPodSpec(terraformVersion string, path string, branch string) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			corev1.Container{
				Image:   "",
				Command: []string{"terraform init"},
			},
		},
	}
}

func getPod(layer *configv1alpha1.TerraformLayer, action Action) corev1.Pod {
	pod := corev1.Pod{
		Spec: defaultPodSpec(layer.Spec.TerraformVersion, layer.Spec.Path, layer.Spec.Branch),
	}
	switch action {
	case PlanAction:
	case ApplyAction:
	}
	return pod

}
