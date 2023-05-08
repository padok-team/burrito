package terraformlayer_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/lock"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	controller "github.com/padok-team/burrito/internal/controllers/terraformlayer"
	utils "github.com/padok-team/burrito/internal/controllers/testing"
	storage "github.com/padok-team/burrito/internal/storage/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var reconciler *controller.Reconciler

const testTime = "Sun May  8 11:21:53 UTC 2023"

func TestLayer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "TerraformLayer Controller Suite")
}

type MockClock struct{}

func (m *MockClock) Now() time.Time {
	t, _ := time.Parse(time.UnixDate, testTime)
	return t
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../../..", "manifests", "crds")},
		ErrorIfCRDPathMissing: true,
	}
	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = configv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	utils.LoadResources(k8sClient, "testdata")
	reconciler = &controller.Reconciler{
		Client:  k8sClient,
		Scheme:  scheme.Scheme,
		Config:  config.TestConfig(),
		Storage: storage.New(),
		Clock:   &MockClock{},
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

func getResult(name types.NamespacedName) (reconcile.Result, error, error, *configv1alpha1.TerraformLayer) {
	result, reconcileError := reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: name,
	})
	layer := &configv1alpha1.TerraformLayer{}
	err := k8sClient.Get(context.TODO(), name, layer)
	return result, reconcileError, err, layer
}

func getLinkedPods(cl client.Client, layer *configv1alpha1.TerraformLayer, action controller.Action, namespace string) (*corev1.PodList, error) {
	list := &corev1.PodList{}
	selector := labels.NewSelector()
	for key, value := range controller.GetDefaultLabels(layer, action) {
		requirement, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return list, err
		}
		selector = selector.Add(*requirement)
	}
	err := cl.List(context.TODO(), list, client.MatchingLabelsSelector{Selector: selector}, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return list, err
	}
	return list, nil
}

var _ = Describe("Layer", func() {
	var layer *configv1alpha1.TerraformLayer
	var reconcileError error
	var err error
	var result reconcile.Result
	var name types.NamespacedName
	Describe("Nominal Case", func() {
		Describe("When a TerraformLayer is created", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-1",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in PlanNeeded state", func() {
				Expect(layer.Status.State).To(Equal("PlanNeeded"))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to DriftCheckInterval", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		Describe("When a TerraformLayer just got planned in autoApply mode", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-2",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in ApplyNeeded state", func() {
				Expect(layer.Status.State).To(Equal("ApplyNeeded"))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		Describe("When a TerraformLayer just got planned in dryRun mode", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-3",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in ApplyNeeded state", func() {
				Expect(layer.Status.State).To(Equal("ApplyNeeded"))
			})
			It("should not be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeFalse())
			})
			It("should set RequeueAfter to DriftDetection", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.DriftDetection))
			})
			It("should not have created a pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
		})
		Describe("When a TerraformLayer just got applied", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-4",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in Idle state", func() {
				Expect(layer.Status.State).To(Equal("Idle"))
			})
			It("should not be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeFalse())
			})
			It("should set RequeueAfter to DriftDetection", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.DriftDetection))
			})
			It("should not have created a pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
		})
		Describe("When a TerraformLayer shares a path with another TerraformLayer and an action is already running", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-5",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should not update status", func() {
				Expect(layer.Status.State).To(Equal(""))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should not have created a plan pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
			It("should not have created an apply pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
		})
		Describe("When a TerraformLayer hasn't been planned since more time than the DriftDetection period", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-6",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should be in PlanNeeded state", func() {
				Expect(layer.Status.State).To(Equal("PlanNeeded"))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a plan pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
			It("should not have created an apply pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
		})
	})
	Describe("Error Case", func() {
		Describe("When a TerraformLayer has errored once on plan", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-1",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in PlanNeeded state", func() {
				Expect(layer.Status.State).To(Equal("PlanNeeded"))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a plan pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		Describe("When a TerraformLayer has errored once on apply", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-2",
					Namespace: "default",
				}
				result, reconcileError, err, layer = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in ApplyNeeded state", func() {
				Expect(layer.Status.State).To(Equal("ApplyNeeded"))
			})
			It("should be locked", func() {
				Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a plan pod", func() {
				pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
