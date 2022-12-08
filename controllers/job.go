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

func defaultPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers: [
			corev1.Container{
				Image: "",
			},
		],

	}
}

func getPod(layer *configv1alpha1.TerraformLayer, action Action) corev1.Pod {
	pod := corev1.Pod{
		Spec: defaultPodSpec(),
	}
	switch action {
	case PlanAction:
	case ApplyAction:
	}
	return pod

}
