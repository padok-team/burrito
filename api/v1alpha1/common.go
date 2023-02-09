package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type OverrideRunnerSpec struct {
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations        []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector       map[string]string             `json:"nodeSelector,omitempty"`
	ServiceAccountName string                        `json:"serviceAccountName,omitempty"`
}

type RemediationStrategy struct {
	PlanOnDrift  bool `json:"planOnDrift,omitempty"`
	ApplyOnDrift bool `json:"applyOnDrift,omitempty"`
	PlanOnPush   bool `json:"planOnPush,omitempty"`
	ApplyOnPush  bool `json:"applyOnPush,omitempty"`
}
