package terraformlayer

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	overrideSpec := configv1alpha1.GetOverrideRunnerSpec(repository, layer)

	defaultSpec.Tolerations = overrideSpec.Tolerations
	defaultSpec.NodeSelector = overrideSpec.NodeSelector
	defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, overrideSpec.Env...)
	defaultSpec.Volumes = append(defaultSpec.Volumes, overrideSpec.Volumes...)
	defaultSpec.Containers[0].VolumeMounts = append(defaultSpec.Containers[0].VolumeMounts, overrideSpec.VolumeMounts...)
	defaultSpec.Containers[0].Resources = overrideSpec.Resources
	defaultSpec.Containers[0].EnvFrom = append(defaultSpec.Containers[0].EnvFrom, overrideSpec.EnvFrom...)
	defaultSpec.ImagePullSecrets = append(defaultSpec.ImagePullSecrets, overrideSpec.ImagePullSecrets...)

	if len(overrideSpec.ServiceAccountName) > 0 {
		defaultSpec.ServiceAccountName = overrideSpec.ServiceAccountName
	}
	if len(overrideSpec.Image) > 0 {
		defaultSpec.Containers[0].Image = overrideSpec.Image
	}

	pod := corev1.Pod{
		Spec: defaultSpec,
		ObjectMeta: metav1.ObjectMeta{
			Labels:      overrideSpec.Metadata.Labels,
			Annotations: overrideSpec.Metadata.Annotations,
		},
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
