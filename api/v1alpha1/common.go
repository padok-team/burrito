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

// +kubebuilder:validation:Enum=dry;autoApply
type RemediationStrategy string

const (
	DryRemediationStrategy       RemediationStrategy = "dry"
	AutoApplyRemediationStrategy RemediationStrategy = "autoApply"
)

type TerraformConfig struct {
	Version    string           `json:"version,omitempty"`
	Terragrunt TerragruntConfig `json:"terragrunt,omitempty"`
}

type TerragruntConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Version string `json:"version,omitempty"`
}
