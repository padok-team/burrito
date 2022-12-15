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

func defaultPodSpec(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image:      fmt.Sprintf("terraform:%s", layer.Spec.TerraformVersion),
				WorkingDir: "/repository",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "gitRepository",
						MountPath: "/repository",
					},
					{
						Name:      "redis-cli",
						MountPath: "/redis",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name: "REDIS_URL",
						Value: "redis",
					},
					{
						Name: "REDIS_USER",
						Value: "redis",
					},
					{
						Name: "REDIS_PASSWORD",
						Value: "",
					},
					{
						Name: "REDIS_PORT",
						Value: "6379",
					},
				},
			},
		},
		InitContainers: []corev1.Container{
			{
				Name:  "0-git-repository-cloning",
				Image: "alpine/git",
				Command: []string{
					"sh",
					"-c",
					fmt.Sprintf("git clone --branch --single-branch %s %s .", layer.Spec.Branch, repository.Spec.Repository.Url),
				},
				WorkingDir: "/repository",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "gitRepository",
						MountPath: "/repository",
					},
				},
			},
			{
				Name:  "1-install-redis-cli",
				Image: "alpine",
				Command: []string{
					"sh",
					"-c",
					"apk install redis;cp /usr/bin/redis-cli /redis/cli",
				},
				WorkingDir: "/repository",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "redis-cli",
						MountPath: "/redis",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "gitRepository",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "redis-cli",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}
}

func getPod(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action Action) corev1.Pod {
	pod := corev1.Pod{
		Spec: defaultPodSpec(layer, repository),
	}
	switch action {
	case PlanAction:
		pod.Spec.Containers[0].Command = []string{
			"sh",
			"-c",
			fmt.Sprintf("%s;%s;%s;%s;%s;%s",
			"cd /repository",
			"terraform init",
			"terraform plan -out plan.out",
			"/redis/cli -u redis://${REDIS_USER}:${REDIS_PASSWORD}@${REDIS_HOST}:${REDIS_PORT} -x HSET set ${KEY} plan_binary <plan.out", //TODO: Replace with the correct key
			"/redis/cli -u redis://${REDIS_USER}:${REDIS_PASSWORD}@${REDIS_HOST}:${REDIS_PORT} delete ${KEY}", //TODO: Replace with the correct key
		)
			
		}
	case ApplyAction:
		pod.Spec.Containers[0].Command = []string{
			"",
		}
	}
	return pod
}
