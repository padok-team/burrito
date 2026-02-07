package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	DefaultRunRetention int = 10
)

type ExtraArgs []string

type OverrideRunnerSpec struct {
	ImagePullSecrets   []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Image              string                        `json:"image,omitempty"`
	ImagePullPolicy    corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	Tolerations        []corev1.Toleration           `json:"tolerations,omitempty"`
	NodeSelector       map[string]string             `json:"nodeSelector,omitempty"`
	Affinity           *corev1.Affinity              `json:"affinity,omitempty"`
	ServiceAccountName string                        `json:"serviceAccountName,omitempty"`
	Resources          corev1.ResourceRequirements   `json:"resources,omitempty"`
	Env                []corev1.EnvVar               `json:"env,omitempty"`
	EnvFrom            []corev1.EnvFromSource        `json:"envFrom,omitempty"`
	Volumes            []corev1.Volume               `json:"volumes,omitempty"`
	VolumeMounts       []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
	Metadata           MetadataOverride              `json:"metadata,omitempty"`
	InitContainers     []corev1.Container            `json:"initContainers,omitempty"`
	Command            []string                      `json:"command,omitempty"`
	Args               []string                      `json:"args,omitempty"`
	ExtraInitArgs      ExtraArgs                     `json:"extraInitArgs,omitempty"`
	ExtraPlanArgs      ExtraArgs                     `json:"extraPlanArgs,omitempty"`
	ExtraApplyArgs     ExtraArgs                     `json:"extraApplyArgs,omitempty"`
}

type MetadataOverride struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type RunHistoryPolicy struct {
	KeepLastRuns *int `json:"runs,omitempty"`
}

type RemediationStrategy struct {
	AutoApply                *bool                      `json:"autoApply,omitempty"`
	ApplyWithoutPlanArtifact *bool                      `json:"applyWithoutPlanArtifact,omitempty"`
	OnError                  OnErrorRemediationStrategy `json:"onError,omitempty"`
}

type OnErrorRemediationStrategy struct {
	MaxRetries *int `json:"maxRetries,omitempty"`
}

type TerraformConfig struct {
	Version string `json:"version,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

type OpenTofuConfig struct {
	Version string `json:"version,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

type TerragruntConfig struct {
	Version string `json:"version,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

func GetTerraformEnabled(repository *TerraformRepository, layer *TerraformLayer) bool {
	if isEnabled(layer.Spec.OpenTofuConfig.Enabled) {
		return false
	}
	return chooseBool(repository.Spec.TerraformConfig.Enabled, layer.Spec.TerraformConfig.Enabled, false)
}

func GetOpenTofuEnabled(repository *TerraformRepository, layer *TerraformLayer) bool {
	if isEnabled(layer.Spec.TerraformConfig.Enabled) {
		return false
	}
	return chooseBool(repository.Spec.OpenTofuConfig.Enabled, layer.Spec.OpenTofuConfig.Enabled, false)
}

func GetTerraformVersion(repository *TerraformRepository, layer *TerraformLayer) string {
	return chooseString(repository.Spec.TerraformConfig.Version, layer.Spec.TerraformConfig.Version)
}

func GetOpenTofuVersion(repository *TerraformRepository, layer *TerraformLayer) string {
	return chooseString(repository.Spec.OpenTofuConfig.Version, layer.Spec.OpenTofuConfig.Version)
}

func GetTerragruntEnabled(repository *TerraformRepository, layer *TerraformLayer) bool {
	return chooseBool(repository.Spec.TerragruntConfig.Enabled, layer.Spec.TerragruntConfig.Enabled, false)
}

func GetTerragruntVersion(repository *TerraformRepository, layer *TerraformLayer) string {
	return chooseString(repository.Spec.TerragruntConfig.Version, layer.Spec.TerragruntConfig.Version)
}

func GetOverrideRunnerSpec(repository *TerraformRepository, layer *TerraformLayer) OverrideRunnerSpec {
	return OverrideRunnerSpec{
		Tolerations:  overrideTolerations(repository.Spec.OverrideRunnerSpec.Tolerations, layer.Spec.OverrideRunnerSpec.Tolerations),
		Affinity:     overrideAffinity(repository.Spec.OverrideRunnerSpec.Affinity, layer.Spec.OverrideRunnerSpec.Affinity),
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
		InitContainers:     MergeInitContainers(repository.Spec.OverrideRunnerSpec.InitContainers, layer.Spec.OverrideRunnerSpec.InitContainers),
		Command:            ChooseSlice(repository.Spec.OverrideRunnerSpec.Command, layer.Spec.OverrideRunnerSpec.Command),
		Args:               ChooseSlice(repository.Spec.OverrideRunnerSpec.Args, layer.Spec.OverrideRunnerSpec.Args),
		ExtraInitArgs:      overrideExtraArgs(repository.Spec.OverrideRunnerSpec.ExtraInitArgs, layer.Spec.OverrideRunnerSpec.ExtraInitArgs),
		ExtraPlanArgs:      overrideExtraArgs(repository.Spec.OverrideRunnerSpec.ExtraPlanArgs, layer.Spec.OverrideRunnerSpec.ExtraPlanArgs),
		ExtraApplyArgs:     overrideExtraArgs(repository.Spec.OverrideRunnerSpec.ExtraApplyArgs, layer.Spec.OverrideRunnerSpec.ExtraApplyArgs),
	}
}

func GetRunHistoryPolicy(repository *TerraformRepository, layer *TerraformLayer) RunHistoryPolicy {
	return RunHistoryPolicy{
		KeepLastRuns: chooseInt(repository.Spec.RunHistoryPolicy.KeepLastRuns, layer.Spec.RunHistoryPolicy.KeepLastRuns, 5),
	}
}

func GetApplyWithoutPlanArtifactEnabled(repository *TerraformRepository, layer *TerraformLayer) bool {
	return chooseBool(repository.Spec.RemediationStrategy.ApplyWithoutPlanArtifact, layer.Spec.RemediationStrategy.ApplyWithoutPlanArtifact, false)
}

func GetAutoApplyEnabled(repo *TerraformRepository, layer *TerraformLayer) bool {
	return chooseBool(repo.Spec.RemediationStrategy.AutoApply, layer.Spec.RemediationStrategy.AutoApply, false)
}

func isEnabled(enabled *bool) bool {
	return enabled != nil && *enabled
}

func chooseBool(a, b *bool, defaultVal bool) bool {
	if b != nil {
		return *b
	}
	if a != nil {
		return *a
	}
	return defaultVal
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

func ChooseSlice(a, b []string) []string {
	if len(b) > 0 {
		return b
	}
	return a
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
	// Track which volume names are overridden by layer (b)
	overriddenNames := make(map[string]bool)

	// Collect all volume names from layer (b)
	for _, elt := range b {
		overriddenNames[elt.Name] = true
	}

	// Add volume mounts from repo (a) that are not overridden
	for _, elt := range a {
		if !overriddenNames[elt.Name] {
			result = append(result, elt)
		}
	}

	// Add all volume mounts from layer (b)
	result = append(result, b...)
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

func overrideTolerations(a, b []corev1.Toleration) []corev1.Toleration {
	result := b

	if len(result) == 0 {
		result = a
	}

	return result
}

func overrideAffinity(repoAffinity, layerAffinity *corev1.Affinity) *corev1.Affinity {
	if layerAffinity != nil {
		return layerAffinity
	}
	return repoAffinity
}

func mergeEnvVars(a, b []corev1.EnvVar) []corev1.EnvVar {
	result := b

	if len(result) == 0 {
		result = a
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

func overrideExtraArgs(a, b ExtraArgs) ExtraArgs {
	result := b

	if len(result) == 0 {
		result = a
	}

	return result
}

func MergeInitContainers(a, b []corev1.Container) []corev1.Container {
	result := []corev1.Container{}
	tempMap := map[string]corev1.Container{}

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
