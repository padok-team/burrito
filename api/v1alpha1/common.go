package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	PlanRunRetention  int = 6
	ApplyRunRetention int = 6
)

type OverrideRunnerSpec struct {
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Image              string                        `json:"image,omitempty"`
	ImagePullPolicy    corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
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
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type RunHistoryPolicy struct {
	KeepLastPlanRuns  *int `json:"plan,omitempty"`
	KeepLastApplyRuns *int `json:"apply,omitempty"`
}

type RemediationStrategy struct {
	AutoApply bool                       `json:"autoApply,omitempty"`
	OnError   OnErrorRemediationStrategy `json:"onError,omitempty"`
}

type OnErrorRemediationStrategy struct {
	MaxRetries *int `json:"maxRetries,omitempty"`
}

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

func GetOverrideRunnerSpec(repository *TerraformRepository, layer *TerraformLayer) OverrideRunnerSpec {
	return OverrideRunnerSpec{
		Tolerations:  mergeTolerations(repository.Spec.OverrideRunnerSpec.Tolerations, layer.Spec.OverrideRunnerSpec.Tolerations),
		NodeSelector: mergeMaps(repository.Spec.OverrideRunnerSpec.NodeSelector, layer.Spec.OverrideRunnerSpec.NodeSelector),
		Metadata: MetadataOverride{
			Annotations: mergeMaps(repository.Spec.OverrideRunnerSpec.Metadata.Annotations, layer.Spec.OverrideRunnerSpec.Metadata.Annotations),
			Labels:      mergeMaps(repository.Spec.OverrideRunnerSpec.Metadata.Labels, layer.Spec.OverrideRunnerSpec.Metadata.Labels),
		},
		Env:                mergeEnvVars(repository.Spec.OverrideRunnerSpec.Env, layer.Spec.OverrideRunnerSpec.Env),
		Volumes:            mergeVolumes(repository.Spec.OverrideRunnerSpec.Volumes, layer.Spec.OverrideRunnerSpec.Volumes),
		VolumeMounts:       mergeVolumeMounts(repository.Spec.OverrideRunnerSpec.VolumeMounts, layer.Spec.OverrideRunnerSpec.VolumeMounts),
		Resources:          mergeResources(repository.Spec.OverrideRunnerSpec.Resources, layer.Spec.OverrideRunnerSpec.Resources),
		EnvFrom:            mergeEnvFrom(repository.Spec.OverrideRunnerSpec.EnvFrom, layer.Spec.OverrideRunnerSpec.EnvFrom),
		Image:              chooseString(repository.Spec.OverrideRunnerSpec.Image, layer.Spec.OverrideRunnerSpec.Image),
		ImagePullPolicy:    chooseImagePullPolicy(repository.Spec.OverrideRunnerSpec.ImagePullPolicy, layer.Spec.OverrideRunnerSpec.ImagePullPolicy),
		ServiceAccountName: chooseString(repository.Spec.OverrideRunnerSpec.ServiceAccountName, layer.Spec.OverrideRunnerSpec.ServiceAccountName),
		ImagePullSecrets:   mergeImagePullSecrets(repository.Spec.OverrideRunnerSpec.ImagePullSecrets, layer.Spec.OverrideRunnerSpec.ImagePullSecrets),
	}
}

func GetRunHistoryPolicy(repository *TerraformRepository, layer *TerraformLayer) RunHistoryPolicy {
	return RunHistoryPolicy{
		KeepLastPlanRuns:  chooseInt(repository.Spec.RunHistoryPolicy.KeepLastPlanRuns, layer.Spec.RunHistoryPolicy.KeepLastPlanRuns, PlanRunRetention),
		KeepLastApplyRuns: chooseInt(repository.Spec.RunHistoryPolicy.KeepLastApplyRuns, layer.Spec.RunHistoryPolicy.KeepLastApplyRuns, ApplyRunRetention),
	}
}

func GetRemediationStrategy(repo *TerraformRepository, layer *TerraformLayer) RemediationStrategy {
	result := RemediationStrategy{
		AutoApply: false,
	}
	if repo.Spec.RemediationStrategy.AutoApply {
		result.AutoApply = true
	}
	if layer.Spec.RemediationStrategy.AutoApply {
		result.AutoApply = true
	}
	return result
}

func mergeImagePullSecrets(a, b []corev1.LocalObjectReference) []corev1.LocalObjectReference {
	result := []corev1.LocalObjectReference{}
	temp := map[string]string{}

	for _, elt := range a {
		temp[elt.Name] = ""
	}
	for _, elt := range b {
		temp[elt.Name] = ""
	}

	for k := range temp {
		result = append(result, corev1.LocalObjectReference{Name: k})
	}
	return result
}

func chooseString(a, b string) string {
	if len(b) > 0 {
		return b
	}
	return a
}

func chooseImagePullPolicy(a, b corev1.PullPolicy) corev1.PullPolicy {
	if b != "" {
		return b
	}
	return a
}

func chooseInt(a, b *int, d int) *int {
	if b != nil {
		return b
	}
	if a != nil {
		return a
	}
	return &d
}

func mergeEnvFrom(a, b []corev1.EnvFromSource) []corev1.EnvFromSource {
	result := []corev1.EnvFromSource{}
	tempSecret := map[string]string{}
	tempConfigMap := map[string]string{}

	for _, elt := range a {
		if elt.ConfigMapRef != nil {
			tempConfigMap[elt.ConfigMapRef.LocalObjectReference.Name] = elt.Prefix
		} else {
			tempSecret[elt.SecretRef.LocalObjectReference.Name] = elt.Prefix
		}
	}
	for _, elt := range b {
		if elt.ConfigMapRef != nil {
			tempConfigMap[elt.ConfigMapRef.LocalObjectReference.Name] = elt.Prefix
		} else {
			tempSecret[elt.SecretRef.LocalObjectReference.Name] = elt.Prefix
		}
	}

	for k, v := range tempConfigMap {
		result = append(result, corev1.EnvFromSource{
			Prefix: v,
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: k,
				},
			},
		})
	}

	for k, v := range tempSecret {
		result = append(result, corev1.EnvFromSource{
			Prefix: v,
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: k,
				},
			},
		})
	}
	return result
}

func mergeResources(a, b corev1.ResourceRequirements) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits:   mergeResourceList(a.Limits, b.Limits),
		Requests: mergeResourceList(a.Requests, b.Requests),
	}
}

func mergeResourceList(a, b corev1.ResourceList) map[corev1.ResourceName]resource.Quantity {
	result := map[corev1.ResourceName]resource.Quantity{}

	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

func mergeVolumeMounts(a, b []corev1.VolumeMount) []corev1.VolumeMount {
	result := []corev1.VolumeMount{}
	tempMap := map[string]corev1.VolumeMount{}

	for _, elt := range a {
		tempMap[elt.Name] = elt
	}
	for _, elt := range b {
		tempMap[elt.Name] = elt
	}

	for _, v := range tempMap {
		result = append(result, v)
	}
	return result
}

func mergeVolumes(a, b []corev1.Volume) []corev1.Volume {
	result := []corev1.Volume{}
	tempMap := map[string]corev1.Volume{}

	for _, elt := range a {
		tempMap[elt.Name] = elt
	}
	for _, elt := range b {
		tempMap[elt.Name] = elt
	}

	for _, v := range tempMap {
		result = append(result, v)
	}
	return result
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
