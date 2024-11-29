package terraformrun_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/lock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Pod", func() {
	var run *configv1alpha1.TerraformRun
	var reconcileError error
	var err error
	var name types.NamespacedName
	Describe("Nominal Case", func() {
		Describe("When a TerraformRun is created with overridden TF args", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-extra-args-plan",
					Namespace: "default",
				}
				_, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in Initial state", func() {
				Expect(run.Status.State).To(Equal("Initial"))
			})
			It("should have an associated pod", func() {
				Expect(run.Status.RunnerPod).To(Not(BeEmpty()))
			})
			It("should have created a lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should have created a pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
			It("should have passed the extra args env variables to the pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(pods.Items[0].Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{
					Name:  "TF_CLI_ARGS_plan",
					Value: "--target 'module.this.random_pet.this[\"first\"]'",
				}))
				Expect(pods.Items[0].Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{
					Name:  "TF_CLI_ARGS_apply",
					Value: "--target 'module.this.random_pet.this[\"first\"]'",
				}))
				Expect(pods.Items[0].Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{
					Name:  "TF_CLI_ARGS_init",
					Value: "--upgrade",
				}))

			})
		})
	})
})
