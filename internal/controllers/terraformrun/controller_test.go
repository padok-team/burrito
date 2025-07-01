package terraformrun_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/lock"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	controller "github.com/padok-team/burrito/internal/controllers/terraformrun"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	utils "github.com/padok-team/burrito/internal/testing"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	logClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
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
var reconcilerMaxConcurrentPods *controller.Reconciler

const testTime = "Sun May  8 11:21:53 UTC 2023"

func TestLayer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "TerraformRun Controller Suite")
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
	logClient, err := logClient.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	utils.LoadResources(k8sClient, "testdata")
	reconciler = &controller.Reconciler{
		Client:       k8sClient,
		Scheme:       scheme.Scheme,
		Config:       config.TestConfig(),
		Clock:        &MockClock{},
		Datastore:    datastore.NewMockClient(),
		K8SLogClient: logClient,
		Recorder: record.NewBroadcasterForTests(1*time.Second).NewRecorder(scheme.Scheme, corev1.EventSource{
			Component: "burrito",
		}),
	}

	// Create the controller with a max parallelism of 2
	configMaxConcurrent := config.TestConfig()
	configMaxConcurrent.Controller.MaxConcurrentRunnerPods = 2
	reconcilerMaxConcurrentPods = &controller.Reconciler{
		Client:       k8sClient,
		Scheme:       scheme.Scheme,
		Config:       configMaxConcurrent,
		Clock:        &MockClock{},
		Datastore:    datastore.NewMockClient(),
		K8SLogClient: logClient,
		Recorder: record.NewBroadcasterForTests(1*time.Second).NewRecorder(scheme.Scheme, corev1.EventSource{
			Component: "burrito",
		}),
	}

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

func getResult(name types.NamespacedName) (reconcile.Result, *configv1alpha1.TerraformRun, error, error) {
	result, reconcileError := reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: name,
	})
	run := &configv1alpha1.TerraformRun{}
	err := k8sClient.Get(context.TODO(), name, run)
	return result, run, reconcileError, err
}

func getResultCustomConfig(name types.NamespacedName, r *controller.Reconciler) (reconcile.Result, *configv1alpha1.TerraformRun, error, error) {
	result, reconcileError := r.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: name,
	})
	run := &configv1alpha1.TerraformRun{}
	err := k8sClient.Get(context.TODO(), name, run)
	return result, run, reconcileError, err
}

func updatePodPhase(name types.NamespacedName, phase corev1.PodPhase) error {
	run := &configv1alpha1.TerraformRun{}
	err := k8sClient.Get(context.TODO(), name, run)
	if err != nil {
		return err
	}
	if run.Status.RunnerPod == "" {
		return errors.New("no pod associated to the run")
	}
	pod := &corev1.Pod{}
	err = k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      run.Status.RunnerPod,
		Namespace: run.Namespace,
	}, pod)
	if err != nil {
		return err
	}
	pod.Status.Phase = phase
	return k8sClient.Status().Update(context.TODO(), pod)
}

func updateLastRunDate(name types.NamespacedName, unixDate string) error {
	run := &configv1alpha1.TerraformRun{}
	err := k8sClient.Get(context.TODO(), name, run)
	if err != nil {
		return err
	}
	run.Status.LastRun = unixDate
	return k8sClient.Status().Update(context.TODO(), run)
}

var _ = Describe("Run", func() {
	var run *configv1alpha1.TerraformRun
	var reconcileError error
	var err error
	var podErr error
	var dateErr error
	var result reconcile.Result
	var name types.NamespacedName
	Describe("Nominal Case", func() {
		Describe("When a TerraformRun is created", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-1",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
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
			It("should set RequeueAfter to 1s", func() {
				Expect(result.RequeueAfter).To(Equal(time.Duration(1 * time.Second)))
			})
			It("should have created a pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		Describe("When a TerraformRun is running its initial pod", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-1",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should be in Running state", func() {
				Expect(run.Status.State).To(Equal("Running"))
			})
			It("should still have a lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should not create a new runner pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		Describe("When a running TerraformRun has a completed pod", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-1",
					Namespace: "default",
				}
				podErr = updatePodPhase(name, corev1.PodSucceeded)
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(podErr).NotTo(HaveOccurred())
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should be in Succeeded state", func() {
				Expect(run.Status.State).To(Equal("Succeeded"))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have released the lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeFalse())
			})
			It("should not create any new pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
				Expect(pods.Items[0].Status.Phase).To(Equal(corev1.PodSucceeded))
			})
		})
		Describe("When TerraformRun is already suceeded", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "nominal-case-1",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should be in Succeeded state", func() {
				Expect(run.Status.State).To(Equal("Succeeded"))
			})
			It("should not set RequeueAfter", func() {
				Expect(result.RequeueAfter).To(Equal(time.Duration(0)))
			})
			It("should not have any pod running", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				var isFinished = func(p corev1.Pod) bool {
					return p.Status.Phase == corev1.PodSucceeded || p.Status.Phase == corev1.PodFailed
				}
				Expect(pods.Items).To(HaveEach(Satisfy(isFinished)))
			})
		})
	})
	Describe("Error Case", func() {
		Describe("When a TerraformRun is associated to an unknown layer", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-1",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should return an error", func() {
				Expect(reconcileError).To(HaveOccurred())
			})
			It("should set RequeueAfter to OnError", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.OnError))
			})
		})
		Describe("When a TerraformRun is associated to a layer with an unknown repo", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-2",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should return an error", func() {
				Expect(reconcileError).To(HaveOccurred())
			})
			It("should set RequeueAfter to OnError", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.OnError))
			})
		})
		Describe("When a TerraformRun has errored once and still in grace period", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-3",
					Namespace: "default",
				}
				// first reconciliation to create the run in initial state
				// not testing because not the purpose of this test
				result, run, reconcileError, err = getResult(name)

				podErr = updatePodPhase(name, corev1.PodFailed)
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(podErr).NotTo(HaveOccurred())
				Expect(dateErr).NotTo(HaveOccurred())
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should still have a lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should end in FailureGracePeriod state", func() {
				Expect(run.Status.State).To(Equal("FailureGracePeriod"))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRun has errored once and not in grace period anymore", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-3",
					Namespace: "default",
				}
				// error-case-3 has already been reconciled in previous test
				t, _ := time.Parse(time.UnixDate, testTime)
				// substract 1h to be sure to be out of grace period
				t = t.Add(-time.Duration(1 * time.Hour))
				dateErr = updateLastRunDate(name, t.Format(time.UnixDate))
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(dateErr).NotTo(HaveOccurred())
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in Retrying state", func() {
				Expect(run.Status.State).To(Equal("Retrying"))
			})
			It("should set RequeueAfter to 1s", func() {
				Expect(result.RequeueAfter).To(Equal(time.Duration(1 * time.Second)))
			})
			It("should still have a lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should have created a second pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(2))
			})
		})
		Describe("When a TerraformRun has reached its max retry limit", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "error-case-3",
					Namespace: "default",
				}
				// error-case-3 has already been reconciled in previous test
				t, _ := time.Parse(time.UnixDate, testTime)
				// substract 1h to be sure to be out of grace period
				t = t.Add(-time.Duration(1 * time.Hour))
				dateErr = updateLastRunDate(name, t.Format(time.UnixDate))
				// make sure the pod has failed again
				podErr = updatePodPhase(name, corev1.PodFailed)
				result, run, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(dateErr).NotTo(HaveOccurred())
				Expect(podErr).NotTo(HaveOccurred())
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in Failed state", func() {
				Expect(run.Status.State).To(Equal("Failed"))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should release the lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeFalse())
			})
			It("should not have created any new pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(2))
			})
		})
	})
	Describe("Concurrent case", func() {
		Describe("When an another TerraformRun is already running on the same layer", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "concurrent-case-1",
					Namespace: "default",
				}
				// not testing for first run because not the purpose of this test
				result, run, reconcileError, err = getResult(name)

				name = types.NamespacedName{
					Name:      "concurrent-case-2",
					Namespace: "default",
				}
				result, run, reconcileError, err = getResult(name)
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
			It("should set RequeueAfter to OnError", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.OnError))
			})
			It("should not have released the lock on the layer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      run.Spec.Layer.Name,
					Namespace: run.Namespace,
				}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLayerLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
			})
			It("should not have created any pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(0))
			})
		})
	})
	Describe("Parallel case", func() {
		var run1, run2, run3 *configv1alpha1.TerraformRun
		var reconcileError1, reconcileError2, reconcileError3 error
		var err1, err2, err3 error
		Describe("When 3 layers are running with a maxConcurrentRunnerPods configuration set to 2", Ordered, func() {
			BeforeAll(func() {
				removeRunnerPods()

				name1 := types.NamespacedName{
					Name:      "parallel-case-1",
					Namespace: "default",
				}
				_, run1, reconcileError1, err1 = getResultCustomConfig(name1, reconcilerMaxConcurrentPods)

				name2 := types.NamespacedName{
					Name:      "parallel-case-2",
					Namespace: "default",
				}
				_, run2, reconcileError2, err2 = getResultCustomConfig(name2, reconcilerMaxConcurrentPods)

				name3 := types.NamespacedName{
					Name:      "parallel-case-3",
					Namespace: "default",
				}
				_, run3, reconcileError3, err3 = getResultCustomConfig(name3, reconcilerMaxConcurrentPods)
			})
			It("should still exists", func() {
				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())
				Expect(err3).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError1).NotTo(HaveOccurred())
				Expect(reconcileError2).NotTo(HaveOccurred())
				Expect(reconcileError3).NotTo(HaveOccurred())
			})
			It("should not have created only 2 runner pods", func() {
				pods1, err1 := reconciler.GetLinkedPods(run1)
				Expect(err1).NotTo(HaveOccurred())
				pods2, err2 := reconciler.GetLinkedPods(run2)
				Expect(err2).NotTo(HaveOccurred())
				pods3, err3 := reconciler.GetLinkedPods(run3)
				Expect(err3).NotTo(HaveOccurred())
				Expect(len(pods1.Items) + len(pods2.Items) + len(pods3.Items)).To(Equal(2))
			})
		})
		Describe("When 3 layers are running without a maxConcurrentRunnerPods configuration", Ordered, func() {
			BeforeAll(func() {
				removeRunnerPods()

				name1 := types.NamespacedName{
					Name:      "parallel-case-1",
					Namespace: "default",
				}
				_, run1, reconcileError1, err1 = getResult(name1)

				name2 := types.NamespacedName{
					Name:      "parallel-case-2",
					Namespace: "default",
				}
				_, run2, reconcileError2, err2 = getResult(name2)

				name3 := types.NamespacedName{
					Name:      "parallel-case-3",
					Namespace: "default",
				}
				_, run3, reconcileError3, err3 = getResult(name3)
			})
			It("should still exists", func() {
				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())
				Expect(err3).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError1).NotTo(HaveOccurred())
				Expect(reconcileError2).NotTo(HaveOccurred())
				Expect(reconcileError3).NotTo(HaveOccurred())
			})
			It("should not have created all 3 runner pods", func() {
				pods1, err1 := reconciler.GetLinkedPods(run1)
				Expect(err1).NotTo(HaveOccurred())
				pods2, err2 := reconciler.GetLinkedPods(run2)
				Expect(err2).NotTo(HaveOccurred())
				pods3, err3 := reconciler.GetLinkedPods(run3)
				Expect(err3).NotTo(HaveOccurred())
				Expect(len(pods1.Items) + len(pods2.Items) + len(pods3.Items)).To(Equal(3))
			})
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func TestGetMaxRetries(t *testing.T) {
	// Config
	defaultMaxRetries := 42
	// Test case 1: Both repo and layer max retries are nil
	r1 := &configv1alpha1.TerraformRepository{Spec: configv1alpha1.TerraformRepositorySpec{}}
	l1 := &configv1alpha1.TerraformLayer{Spec: configv1alpha1.TerraformLayerSpec{}}
	expectedResult1 := defaultMaxRetries
	result1 := controller.GetMaxRetries(defaultMaxRetries, r1, l1)
	if result1 != expectedResult1 {
		t.Errorf("Test case 1 failed: expected %d, got %d", expectedResult1, result1)
	}

	// Test case 2: Repo max retries is nil, layer max retries is not nil
	r2 := &configv1alpha1.TerraformRepository{Spec: configv1alpha1.TerraformRepositorySpec{}}
	l2 := &configv1alpha1.TerraformLayer{Spec: configv1alpha1.TerraformLayerSpec{RemediationStrategy: configv1alpha1.RemediationStrategy{OnError: configv1alpha1.OnErrorRemediationStrategy{MaxRetries: intPtr(3)}}}}
	expectedResult2 := 3
	result2 := controller.GetMaxRetries(defaultMaxRetries, r2, l2)
	if result2 != expectedResult2 {
		t.Errorf("Test case 2 failed: expected %d, got %d", expectedResult2, result2)
	}

	// Test case 3: Repo max retries is not nil, layer max retries is nil
	r3 := &configv1alpha1.TerraformRepository{Spec: configv1alpha1.TerraformRepositorySpec{RemediationStrategy: configv1alpha1.RemediationStrategy{OnError: configv1alpha1.OnErrorRemediationStrategy{MaxRetries: intPtr(7)}}}}
	l3 := &configv1alpha1.TerraformLayer{Spec: configv1alpha1.TerraformLayerSpec{}}
	expectedResult3 := 7
	result3 := controller.GetMaxRetries(defaultMaxRetries, r3, l3)
	if result3 != expectedResult3 {
		t.Errorf("Test case 3 failed: expected %d, got %d", expectedResult3, result3)
	}

	// Test case 4: Both repo and layer max retries are not nil, layer takes precedence
	r4 := &configv1alpha1.TerraformRepository{Spec: configv1alpha1.TerraformRepositorySpec{RemediationStrategy: configv1alpha1.RemediationStrategy{OnError: configv1alpha1.OnErrorRemediationStrategy{MaxRetries: intPtr(4)}}}}
	l4 := &configv1alpha1.TerraformLayer{Spec: configv1alpha1.TerraformLayerSpec{RemediationStrategy: configv1alpha1.RemediationStrategy{OnError: configv1alpha1.OnErrorRemediationStrategy{MaxRetries: intPtr(8)}}}}
	expectedResult4 := 8
	result4 := controller.GetMaxRetries(defaultMaxRetries, r4, l4)
	if result4 != expectedResult4 {
		t.Errorf("Test case 4 failed: expected %d, got %d", expectedResult4, result4)
	}
}

func intPtr(i int) *int {
	return &i
}

func removeRunnerPods() {
	// make sure that there are no runner pods
	pods := &corev1.PodList{}
	err := k8sClient.List(context.TODO(), pods, client.MatchingLabels{
		"burrito/component": "runner",
	})
	Expect(err).NotTo(HaveOccurred())
	for _, pod := range pods.Items {
		err := k8sClient.Delete(context.TODO(), &pod)
		Expect(err).NotTo(HaveOccurred())
	}
	// wait for the pods to be deleted
	Eventually(func() int {
		pods := &corev1.PodList{}
		err := k8sClient.List(context.TODO(), pods, client.MatchingLabels{
			"burrito/component": "runner",
		})
		Expect(err).NotTo(HaveOccurred())
		return len(pods.Items)
	}, 10*time.Second, 1*time.Second).Should(Equal(0))
}

func TestGetRunExponentialBackOffTime(t *testing.T) {
	tt := []struct {
		name         string
		defaultTime  time.Duration
		run          *configv1alpha1.TerraformRun
		expectedTime time.Duration
	}{
		{
			"Exponential backoff: No retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 0,
				},
			},
			time.Minute,
		},
		{
			"Exponential backoff: 1 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 1,
				},
			},
			2 * time.Minute,
		},
		{
			"Exponential backoff : 2 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 2,
				},
			},
			7 * time.Minute,
		},
		{
			"Exponential backoff : 3 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 3,
				},
			},
			20 * time.Minute,
		},
		{
			"Exponential backoff : 5 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 5,
				},
			},
			148 * time.Minute,
		},
		{
			"Exponential backoff : 10 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 10,
				},
			},
			22026 * time.Minute,
		},
		{
			"Exponential backoff : 17 retry",
			time.Minute,
			&configv1alpha1.TerraformRun{
				Status: configv1alpha1.TerraformRunStatus{
					Retries: 17,
				},
			},
			24154952 * time.Minute,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := controller.GetRunExponentialBackOffTime(tc.defaultTime, tc.run)
			if tc.expectedTime != result {
				t.Errorf("different version computed: expected %s go %s", tc.expectedTime, result)
			}
		})
	}
}
