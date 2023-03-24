package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type OverrideRunnerSpec struct {
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Tolerations        []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector       map[string]string             `json:"nodeSelector,omitempty"`
	ServiceAccountName string                        `json:"serviceAccountName,omitempty"`
	Resources          corev1.ResourceRequirements   `json:"resources,omitempty"`
	Env                []corev1.EnvVar               `json:"env,omitempty"`
	EnvFrom            []corev1.EnvFromSource        `json:"envFrom,omitempty"`
	Volumes            []corev1.Volume               `json:"volumes,omitempty"`
	VolumeMounts       []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	Metadata           MetadataOverride              `json:"metadata,omitempty"`
}

type MetadataOverride struct {
	Annotations  map[string]string `json:"annotations,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	GenerateName string            `json:"generateName,omitempty"`
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

func getOverrideRunenrSpec(repository *TerraformRepository, layer *TerraformLayer) OverrideRunnerSpec {
	return OverrideRunnerSpec{
		Tolerations:  mergeTolerations(repository.Spec.OverrideRunnerSpec.Tolerations, layer.Spec.OverrideRunnerSpec.Tolerations),
		NodeSelector: mergeMaps(repository.Spec.OverrideRunnerSpec.NodeSelector, layer.Spec.OverrideRunnerSpec.NodeSelector),
		Metadata: MetadataOverride{
			Annotations: mergeMaps(repository.Spec.OverrideRunnerSpec.Metadata.Annotations, layer.Spec.OverrideRunnerSpec.Metadata.Annotations),
			Labels:      mergeMaps(repository.Spec.OverrideRunnerSpec.Metadata.Labels, layer.Spec.OverrideRunnerSpec.Metadata.Labels),
		},
		Env: mergeEnvVars(repository.Spec.OverrideRunnerSpec.Env, layer.Spec.OverrideRunnerSpec.Env),
	}

}

func mergeTolerations(a, b []corev1.Toleration) []corev1.Toleration {
	result := []corev1.Toleration{}
	tempMap := map[string]corev1.Toleration{}

	for _, elt := range a {
		tempMap[elt.Key] = elt
	}
	for _, elt := range b {
		tempMap[elt.Key] = elt
	}

	for _, v := range tempMap {
		result = append(result, v)
	}
	return result
}

func mergeEnvVars(a, b []corev1.EnvVar) []corev1.EnvVar {
	result := []corev1.EnvVar{}
	tempMap := map[string]string{}

	for _, elt := range a {
		tempMap[elt.Name] = elt.Value
	}
	for _, elt := range b {
		tempMap[elt.Name] = elt.Value
	}

	for k, v := range tempMap {
		result = append(result, corev1.EnvVar{Name: k, Value: v})
	}

	return result

}

func mergeMaps(a, b map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
