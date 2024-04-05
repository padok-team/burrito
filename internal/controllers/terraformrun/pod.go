package terraformrun

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/version"
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

func (r *Reconciler) ensureHermitcrabSecret(tenantNamespace string) error {
	secret := &corev1.Secret{}
	err := r.Client.Get(context.Background(), client.ObjectKey{Namespace: r.Config.Controller.MainNamespace,
		Name: r.Config.Hermitcrab.CertificateSecretName}, secret)
	if err != nil {
		return err
	}
	if _, ok := secret.Data["ca.crt"]; !ok {
		return fmt.Errorf("ca.crt not found in secret %s", r.Config.Hermitcrab.CertificateSecretName)
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Config.Hermitcrab.CertificateSecretName,
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
	log.Infof("hermitcrab certificate is available in namespace %s", tenantNamespace)
	return nil
}

func (r *Reconciler) getPod(run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.Pod {
	defaultSpec := defaultPodSpec(r.Config, layer, repository)

	if r.Config.Hermitcrab.Enabled {
		err := r.ensureHermitcrabSecret(layer.Namespace)
		if err != nil {
			log.Errorf("failed to ensure HermitCrab secret in namespace %s: %s", layer.Namespace, err)
		} else {
			defaultSpec.Volumes = append(defaultSpec.Volumes, corev1.Volume{
				Name: "hermitcrab-ca-cert",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: r.Config.Hermitcrab.CertificateSecretName,
						Items: []corev1.KeyToPath{
							{
								Key:  "ca.crt",
								Path: "hermitcrab-ca.crt",
							},
						},
					},
				},
			})
			defaultSpec.Containers[0].VolumeMounts = append(defaultSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
				MountPath: "/etc/ssl/certs/hermitcrab-ca.crt",
				Name:      "hermitcrab-ca-cert",
				SubPath:   "hermitcrab-ca.crt",
			})

			defaultSpec.Containers[0].Env = append(defaultSpec.Containers[0].Env,
				corev1.EnvVar{
					Name:  "HERMITCRAB_ENABLED",
					Value: "true",
				},
				corev1.EnvVar{
					Name:  "HERMITCRAB_URL",
					Value: fmt.Sprintf("https://burrito-hermitcrab.%s.svc.cluster.local/v1/providers/", r.Config.Controller.MainNamespace),
				},
			)
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

func defaultPodSpec(config *config.Config, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) corev1.PodSpec {
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
		},
		RestartPolicy:      corev1.RestartPolicyNever,
		ServiceAccountName: "burrito-runner",
		Containers: []corev1.Container{
			{
				Name:  "runner",
				Image: fmt.Sprintf("ghcr.io/padok-team/burrito:%s", version.Version),
				Args:  []string{"runner", "start"},
				VolumeMounts: []corev1.VolumeMount{
					{
						MountPath: "/home/burrito/.ssh/known_hosts",
						Name:      "ssh-known-hosts",
						SubPath:   "known_hosts",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "BURRITO_REDIS_HOSTNAME",
						Value: config.Redis.Hostname,
					},
					{
						Name:  "BURRITO_REDIS_SERVERPORT",
						Value: fmt.Sprintf("%d", config.Redis.ServerPort),
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
						Value: "/home/burrito/.ssh/known_hosts",
					},
				},
			},
		},
	}
}
