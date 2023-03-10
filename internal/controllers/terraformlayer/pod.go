package terraformlayer

import (
	"fmt"
	"strconv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/version"
	corev1 "k8s.io/api/core/v1"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func (r *Reconciler) getPod(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action Action) corev1.Pod {
	defaultSpec := defaultPodSpec(r.Config, layer, repository)

	switch action {
	case PlanAction:
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "BURRITO_RUNNER_ACTION",
			Value: "plan",
		})
	case ApplyAction:
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "BURRITO_RUNNER_ACTION",
			Value: "apply",
		})
	}
	if repository.Spec.Repository.SecretName != "" {
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "username",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "password",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_SSHPRIVATEKEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "sshPrivateKey",
					Optional: &[]bool{true}[0],
				},
			},
		})
	}

	pod := corev1.Pod{
		Spec: mergeSpecs(defaultSpec, repository.Spec.OverrideRunnerSpec, layer.Spec.OverrideRunnerSpec),
	}
	pod.SetNamespace(layer.Namespace)
	pod.SetGenerateName(fmt.Sprintf("%s-%s-", layer.Name, action))

	return pod
}

func defaultPodSpec(config *config.Config, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name:         "repository",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
			{
				Name: "ssh-known-hosts",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: config.Runner.SSHKnownHostsConfigMapName,
						},
						Optional: &[]bool{true}[0],
					},
				},
			},
		},
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: "burrito-runner",
		Containers: []corev1.Container{
			{
				Name:       "runner",
				Image:      fmt.Sprintf("ghcr.io/padok-team/burrito:%s", version.Version),
				WorkingDir: "/repository",
				Args:       []string{"runner", "start"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "repository",
						MountPath: "/repository",
					},
					{
						MountPath: "/go/.ssh/",
						Name:      "ssh-known-hosts",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "BURRITO_REDIS_URL",
						Value: config.Redis.URL,
					},
					{
						Name:  "BURRITO_REDIS_PASSWORD",
						Value: config.Redis.Password,
					},
					{
						Name:  "BURRITO_REDIS_DATABASE",
						Value: fmt.Sprintf("%d", config.Redis.Database),
					},
					{
						Name:  "BURRITO_RUNNER_REPOSITORY_URL",
						Value: repository.Spec.Repository.Url,
					},
					{
						Name:  "BURRITO_RUNNER_PATH",
						Value: layer.Spec.Path,
					},
					{
						Name:  "BURRITO_RUNNER_BRANCH",
						Value: layer.Spec.Branch,
					},
					{
						Name:  "BURRITO_RUNNER_VERSION",
						Value: getTerraformVersion(layer, repository),
					},
					{
						Name:  "BURRITO_RUNNER_TERRAGRUNT_ENABLED",
						Value: strconv.FormatBool(getTerragruntEnabled(layer, repository)),
					},
					{
						Name:  "BURRITO_RUNNER_TERRAGRUNT_VERSION",
						Value: getTerragruntVersion(layer, repository),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_NAME",
						Value: layer.GetObjectMeta().GetName(),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_NAMESPACE",
						Value: layer.GetObjectMeta().GetNamespace(),
					},
					{
						Name:  "SSH_KNOWN_HOSTS",
						Value: "/go/.ssh/known_hosts",
					},
				},
			},
		},
	}
}

func mergeSpecs(defaultSpec corev1.PodSpec, repositorySpec configv1alpha1.OverrideRunnerSpec, layerSpec configv1alpha1.OverrideRunnerSpec) corev1.PodSpec {
	if len(repositorySpec.ImagePullSecrets) > 0 {
		defaultSpec.ImagePullSecrets = repositorySpec.ImagePullSecrets
	}
	if len(layerSpec.ImagePullSecrets) > 0 {
		defaultSpec.ImagePullSecrets = layerSpec.ImagePullSecrets
	}

	if len(repositorySpec.Tolerations) > 0 {
		defaultSpec.Tolerations = repositorySpec.Tolerations
	}
	if len(layerSpec.Tolerations) > 0 {
		defaultSpec.Tolerations = layerSpec.Tolerations
	}

	if len(repositorySpec.NodeSelector) > 0 {
		defaultSpec.NodeSelector = repositorySpec.NodeSelector
	}
	if len(layerSpec.NodeSelector) > 0 {
		defaultSpec.NodeSelector = layerSpec.NodeSelector
	}

	if len(repositorySpec.ServiceAccountName) > 0 {
		defaultSpec.ServiceAccountName = repositorySpec.ServiceAccountName
	}
	if len(layerSpec.ServiceAccountName) > 0 {
		defaultSpec.ServiceAccountName = layerSpec.ServiceAccountName
	}
	return defaultSpec
}

func getTerraformVersion(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) string {
	version := repository.Spec.TerraformConfig.Version
	if len(layer.Spec.TerraformConfig.Version) > 0 {
		version = layer.Spec.TerraformConfig.Version
	}
	return version
}

func getTerragruntVersion(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) string {
	version := repository.Spec.TerraformConfig.TerragruntConfig.Version
	if len(layer.Spec.TerraformConfig.TerragruntConfig.Version) > 0 {
		version = layer.Spec.TerraformConfig.TerragruntConfig.Version
	}
	return version
}

func getTerragruntEnabled(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) bool {
	enabled := false
	if repository.Spec.TerraformConfig.TerragruntConfig.Enabled != nil {
		enabled = *repository.Spec.TerraformConfig.TerragruntConfig.Enabled
	}
	if layer.Spec.TerraformConfig.TerragruntConfig.Enabled != nil {
		enabled = *layer.Spec.TerraformConfig.TerragruntConfig.Enabled
	}
	return enabled
}
