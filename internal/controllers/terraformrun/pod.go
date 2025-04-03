package terraformrun

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func getDefaultLabels(run *configv1alpha1.TerraformRun) map[string]string {
	return map[string]string{
		"burrito/component":  "runner",
		"burrito/managed-by": run.Name,
		"burrito/action":     string(run.Spec.Action),
	}
}

func getLabelSelector(run *configv1alpha1.TerraformRun) labels.Selector {
	selector := labels.NewSelector()
	for key, value := range getDefaultLabels(run) {
		requirement, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return selector
		}
		selector = selector.Add(*requirement)
	}
	return selector
}

func (r *Reconciler) GetLinkedPods(run *configv1alpha1.TerraformRun) (*corev1.PodList, error) {
	list := &corev1.PodList{}
	selector := getLabelSelector(run)
	err := r.Client.List(context.Background(), list, client.MatchingLabelsSelector{Selector: selector}, &client.ListOptions{
		Namespace: run.Namespace,
	})
	if err != nil {
		return list, err
	}
	return list, nil
}

func (r *Reconciler) ensureCertificateAuthoritySecret(tenantNamespace, caSecretName string) error {
	secret := &corev1.Secret{}
	err := r.Client.Get(context.Background(), client.ObjectKey{
		Namespace: r.Config.Controller.MainNamespace,
		Name:      caSecretName,
	}, secret)
	if err != nil {
		return err
	}
	if _, ok := secret.Data["ca.crt"]; !ok {
		return fmt.Errorf("ca.crt not found in secret %s", caSecretName)
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caSecretName,
			Namespace: tenantNamespace,
		},
		Data: map[string][]byte{
			"ca.crt": secret.Data["ca.crt"],
		},
	}
	err = r.Client.Create(context.Background(), secret)
	if err != nil && apierrors.IsAlreadyExists(err) {
		err = r.Client.Update(context.Background(), secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	log.Infof("CA certificate is available in namespace %s", tenantNamespace)
	return nil
}

func mountCA(podSpec *corev1.PodSpec, caSecretName, caName string) {
	volumeName := fmt.Sprintf("%s-cert", caName)
	mountPath := fmt.Sprintf("/etc/ssl/certs/%s.crt", caName)
	caFilename := fmt.Sprintf("%s.crt", caName)

	podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: caSecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  "ca.crt",
						Path: caFilename,
					},
				},
			},
		},
	})
	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
		MountPath: mountPath,
		Name:      volumeName,
		SubPath:   caFilename,
	})
}

func (r *Reconciler) getPod(run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.Pod {
	defaultSpec := defaultPodSpec(r.Config, layer, repository, run)

	if r.Config.Hermitcrab.Enabled {
		err := r.ensureCertificateAuthoritySecret(layer.Namespace, r.Config.Hermitcrab.CertificateSecretName)
		if err != nil {
			log.Errorf("failed to ensure HermitCrab secret in namespace %s: %s", layer.Namespace, err)
		} else {
			mountCA(&defaultSpec, r.Config.Hermitcrab.CertificateSecretName, "burrito-hermitcrab-ca")
			defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env,
				corev1.EnvVar{
					Name:  "BURRITO_HERMITCRAB_ENABLED",
					Value: "true",
				},
				corev1.EnvVar{
					Name:  "BURRITO_HERMITCRAB_URL",
					Value: fmt.Sprintf("https://burrito-hermitcrab.%s.svc.cluster.local/v1/providers/", r.Config.Controller.MainNamespace),
				},
			)
		}
	}
	if r.Config.Datastore.TLS {
		err := r.ensureCertificateAuthoritySecret(layer.Namespace, r.Config.Datastore.CertificateSecretName)
		if err != nil {
			log.Errorf("failed to ensure Datastore secret in namespace %s: %s", layer.Namespace, err)
		} else {
			mountCA(&defaultSpec, r.Config.Datastore.CertificateSecretName, "burrito-datastore-ca")
			defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
				Name:  "BURRITO_DATASTORE_TLS",
				Value: "true",
			})
		}
	}
	switch Action(run.Spec.Action) {
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
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_GITHUBAPPID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "githubAppId",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_GITHUBAPPINSTALLATIONID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "githubAppInstallationId",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_GITHUBAPPPRIVATEKEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "githubAppPrivateKey",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_GITHUBTOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "githubToken",
					Optional: &[]bool{true}[0],
				},
			},
		})
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name: "BURRITO_RUNNER_REPOSITORY_GITLABTOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: repository.Spec.Repository.SecretName,
					},
					Key:      "gitlabToken",
					Optional: &[]bool{true}[0],
				},
			},
		})
	}

	overrideSpec := configv1alpha1.GetOverrideRunnerSpec(repository, layer)

	defaultSpec.Tolerations = overrideSpec.Tolerations
	defaultSpec.Affinity = overrideSpec.Affinity
	defaultSpec.NodeSelector = overrideSpec.NodeSelector
	defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, overrideSpec.Env...)
	defaultSpec.Volumes = append(defaultSpec.Volumes, overrideSpec.Volumes...)
	defaultSpec.Containers[0].VolumeMounts = append(defaultSpec.Containers[0].VolumeMounts, overrideSpec.VolumeMounts...)
	defaultSpec.Containers[0].Resources = overrideSpec.Resources
	defaultSpec.Containers[0].EnvFrom = append(defaultSpec.Containers[0].EnvFrom, overrideSpec.EnvFrom...)
	defaultSpec.ImagePullSecrets = append(defaultSpec.ImagePullSecrets, overrideSpec.ImagePullSecrets...)
	defaultSpec.Containers[0].ImagePullPolicy = overrideSpec.ImagePullPolicy

	if len(overrideSpec.ServiceAccountName) > 0 {
		defaultSpec.ServiceAccountName = overrideSpec.ServiceAccountName
	}
	if len(overrideSpec.Image) > 0 {
		defaultSpec.Containers[0].Image = overrideSpec.Image
	}

	if len(overrideSpec.ExtraInitArgs) > 0 {
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "TF_CLI_ARGS_init",
			Value: strings.Join(overrideSpec.ExtraInitArgs, " "),
		})
	}
	if len(overrideSpec.ExtraPlanArgs) > 0 {
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "TF_CLI_ARGS_plan",
			Value: strings.Join(overrideSpec.ExtraPlanArgs, " "),
		})
	}
	if len(overrideSpec.ExtraApplyArgs) > 0 {
		defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "TF_CLI_ARGS_apply",
			Value: strings.Join(overrideSpec.ExtraApplyArgs, " "),
		})
	}

	pod := corev1.Pod{
		Spec: defaultSpec,
		ObjectMeta: metav1.ObjectMeta{
			Labels:      mergeMaps(overrideSpec.Metadata.Labels, getDefaultLabels(run)),
			Annotations: overrideSpec.Metadata.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: run.GetAPIVersion(),
					Kind:       run.GetKind(),
					Name:       run.Name,
					UID:        run.UID,
				},
			},
		},
	}
	pod.SetNamespace(layer.Namespace)
	pod.SetGenerateName(fmt.Sprintf("%s-%s-", layer.Name, run.Spec.Action))

	return pod
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

func defaultPodSpec(config *config.Config, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, run *configv1alpha1.TerraformRun) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: []corev1.Volume{
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
			{
				Name: "burrito-token",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
									Audience:          "burrito",
									ExpirationSeconds: &[]int64{3600}[0],
									Path:              "burrito",
								},
							},
						},
					},
				},
			},
		},
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: "burrito-runner",
		Containers: []corev1.Container{
			{
				Name:            "runner",
				Image:           fmt.Sprintf("%s:%s", config.Runner.Image.Repository, config.Runner.Image.Tag),
				ImagePullPolicy: corev1.PullPolicy(config.Runner.Image.PullPolicy),
				Args:            []string{"runner", "start"},
				VolumeMounts: []corev1.VolumeMount{
					{
						MountPath: "/home/burrito/.ssh/known_hosts",
						Name:      "ssh-known-hosts",
						SubPath:   "known_hosts",
					},
					{
						MountPath: "/var/run/secrets/token",
						Name:      "burrito-token",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "BURRITO_RUNNER_LAYER_NAME",
						Value: layer.Name,
					},
					{
						Name:  "BURRITO_RUNNER_LAYER_NAMESPACE",
						Value: layer.Namespace,
					},
					{
						Name:  "BURRITO_RUNNER_RUN",
						Value: run.Name,
					},
					{
						Name:  "SSH_KNOWN_HOSTS",
						Value: "/home/burrito/.ssh/known_hosts",
					},
					{
						Name:  "BURRITO_DATASTORE_HOSTNAME",
						Value: config.Datastore.Hostname,
					},
				},
			},
		},
	}
}
