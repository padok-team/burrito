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
	return pod
}

func defaultPodSpec(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:       "runner",
				Image:      fmt.Sprintf("eu.gcr.io/padok-playground/burrito:%s", "alpha"),
				WorkingDir: "/repository",
				Args:       []string{"runner", "start"},
				Env: []corev1.EnvVar{
					{
						Name:  "BURRITO_REDIS_URL",
						Value: "burrito-redis-headless:6379",
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
						Name:  "BURRITO_RUNNER_REPOSITORY_SSH",
						Value: "",
					},
					{
						Name:  "BURRITO_RUNNER_REPOSITORY_USERNAME",
						Value: "",
					},
					{
						Name:  "BURRITO_RUNNER_REPOSITORY_PASSWORD",
						Value: "",
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
						Name:  "BURRITO_RUNNER_LAYER_LOCK",
						Value: fmt.Sprintf("%s%s", CachePrefixLock, computeHash(layer.Spec.Repository.Name, layer.Spec.Repository.Namespace, layer.Spec.Path)),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_PLANSUM",
						Value: fmt.Sprintf("%s%s", CachePrefixLastPlannedArtifact, computeHash(layer.Spec.Repository.Name, layer.Spec.Repository.Namespace, layer.Spec.Path, layer.Spec.Branch)),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_PLANBIN",
						Value: fmt.Sprintf("%s%s", CachePrefixLastPlannedArtifactBin, computeHash(layer.Spec.Repository.Name, layer.Spec.Repository.Namespace, layer.Spec.Path, layer.Spec.Branch)),
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_APPLYSUM",
						Value: fmt.Sprintf("%s%s", CachePrefixLastAppliedArtifact, computeHash(layer.Spec.Repository.Name, layer.Spec.Repository.Namespace, layer.Spec.Path, layer.Spec.Branch)),
					},
				},
			},
		},
	}
}
