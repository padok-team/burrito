package terraformlayer_test

import (
	"fmt"
	"reflect"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformlayer"
	"github.com/padok-team/burrito/internal/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// func TestGetPod(t *testing.T) {
// 	tt := []struct {
// 		name        string
// 		config      *config.Config
// 		layer       *configv1alpha1.TerraformLayer
// 		repository  *configv1alpha1.TerraformRepository
// 		action      terraformlayer.Action
// 		expectedPod *corev1.Pod
// 	}{}
// }

func TestDefaultSpec(t *testing.T) {
	tt := []struct {
		name            string
		config          *config.Config
		layer           *configv1alpha1.TerraformLayer
		repository      *configv1alpha1.TerraformRepository
		expectedPodSpec *corev1.PodSpec
	}{
		{"default", &config.Config{
			Runner: config.RunnerConfig{
				SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
			},
		}, &configv1alpha1.TerraformLayer{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TerraformLayer",
				APIVersion: "config.terraform.padok.cloud/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "random-pets",
				Namespace: "burrito",
			},
			Spec: configv1alpha1.TerraformLayerSpec{
				Path:             "internal/e2e/random-pets",
				Branch:           "main",
				TerraformVersion: "1.3.1",
			},
		}, &configv1alpha1.TerraformRepository{
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{
					Url: "github.com/padok-team/burrito",
				},
			},
		}, &corev1.PodSpec{
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
								Name: "burrito-ssh-known-hosts",
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
							Value: "github.com/padok-team/burrito",
						},
						{
							Name:  "BURRITO_RUNNER_PATH",
							Value: "internal/e2e/random-pets",
						},
						{
							Name:  "BURRITO_RUNNER_BRANCH",
							Value: "main",
						},
						{
							Name:  "BURRITO_RUNNER_VERSION",
							Value: "1.3.1",
						},
						{
							Name:  "BURRITO_RUNNER_LAYER_NAME",
							Value: "random-pets",
						},
						{
							Name:  "BURRITO_RUNNER_LAYER_NAMESPACE",
							Value: "burrito",
						},
						{
							Name:  "SSH_KNOWN_HOSTS",
							Value: "/go/.ssh/known_hosts",
						},
					},
				},
			},
		}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			pod := terraformlayer.DefaultPodSpec(tc.config, tc.layer, tc.repository)
			if !reflect.DeepEqual(pod, tc.expectedPodSpec) {
				t.Errorf("different generated default pod spec and expected spec")
			}
		})
	}
}

// func TestMergeSpecs(t *testing.T) {
// 	tt := []struct {
// 		name
// 		DefaultPodSpec corve1.PodSpec
// 		repoOverride   configv1alpha1.OverrideRunnerSpec
// 		layerOverride  configv1alpha1.OverrideRunnerSpec
// 	}{}
// }
