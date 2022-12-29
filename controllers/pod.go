package controllers

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func getPod(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action Action) corev1.Pod {
	pod := corev1.Pod{
		Spec: defaultPodSpec(layer, repository),
	}
	pod.SetNamespace(layer.Namespace)
	pod.SetGenerateName(fmt.Sprintf("%s-%s-", layer.Name, action))
	switch action {
	case PlanAction:
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "BURRITO_RUNNER_ACTION",
			Value: "plan",
		})
	case ApplyAction:
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "BURRITO_RUNNER_ACTION",
			Value: "apply",
		})
	}
	if repository.Spec.Repository.SecretRef.Name != "" {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretRef.Name,
					},
					Key:      "username",
					Optional: &[]bool{true}[0],
				},
			},
		})
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretRef.Name,
					},
					Key:      "password",
					Optional: &[]bool{true}[0],
				},
			},
		})
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretRef.Name,
					},
					Key:      "password",
					Optional: &[]bool{true}[0],
				},
			},
		})
	}
	return pod
}

func defaultPodSpec(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name:         "repository",
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
		},
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: "burrito-runner",
		Containers: []corev1.Container{
			{
				Name:       "runner",
				Image:      fmt.Sprintf("ghcr.io/padok-team/burrito:%s", "latest"),
				WorkingDir: "/repository",
				Args:       []string{"runner", "start"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "repository",
						MountPath: "/repository",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "BURRITO_REDIS_URL",
						Value: "burrito-redis-master:6379",
					},
					{
						Name:  "BURRITO_REDIS_PASSWORD",
						Value: "",
					},
					{
						Name:  "BURRITO_REDIS_DATABASE",
						Value: "0",
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
						Value: layer.Spec.TerraformVersion,
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_NAME",
						Value: layer.GetObjectMeta().GetName(),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_NAMESPACE",
						Value: layer.GetObjectMeta().GetNamespace(),
					},
				},
			},
		},
	}
}
