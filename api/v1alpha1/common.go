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
	Version          string           `json:"version,omitempty"`
	TerragruntConfig TerragruntConfig `json:"terragrunt,omitempty"`
}

type TerragruntConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Version string `json:"version,omitempty"`
}

func GetTerraformVersion(repository *TerraformRepository, layer *TerraformLayer) string {
	version := repository.Spec.TerraformConfig.Version
	if len(layer.Spec.TerraformConfig.Version) > 0 {
		version = layer.Spec.TerraformConfig.Version
	}
	return version
}

func GetTerragruntVersion(repository *TerraformRepository, layer *TerraformLayer) string {
	version := repository.Spec.TerraformConfig.TerragruntConfig.Version
	if len(layer.Spec.TerraformConfig.TerragruntConfig.Version) > 0 {
		version = layer.Spec.TerraformConfig.TerragruntConfig.Version
	}
	return version
}

func GetTerragruntEnabled(repository *TerraformRepository, layer *TerraformLayer) bool {
	enabled := false
	if repository.Spec.TerraformConfig.TerragruntConfig.Enabled != nil {
		enabled = *repository.Spec.TerraformConfig.TerragruntConfig.Enabled
	}
	if layer.Spec.TerraformConfig.TerragruntConfig.Enabled != nil {
		enabled = *layer.Spec.TerraformConfig.TerragruntConfig.Enabled
	}
	return enabled
}
