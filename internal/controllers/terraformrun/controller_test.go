package terraformrun_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	controller "github.com/padok-team/burrito/internal/controllers/terraformrun"
	utils "github.com/padok-team/burrito/internal/testing"
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

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	utils.LoadResources(k8sClient, "testdata")
	reconciler = &controller.Reconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
		Config: config.TestConfig(),
		Clock:  &MockClock{},
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

var _ = Describe("Run", func() {
	var run *configv1alpha1.TerraformRun
	var reconcileError error
	var err error
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
			It("should end in Running state", func() {
				Expect(run.Status.State).To(Equal("Running"))
			})
			It("should have an associated pod", func() {
				Expect(run.Status.RunnerPod).To(Not(BeEmpty()))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
			It("should have created a plan pod", func() {
				pods, err := reconciler.GetLinkedPods(run)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(pods.Items)).To(Equal(1))
			})
		})
		// Describe("When a TerraformRun is running", Ordered, func() {
		// 	BeforeAll(func() {
		// 		name = types.NamespacedName{
		// 			Name:      "nominal-case-2",
		// 			Namespace: "default",
		// 		}
		// 		result, run, reconcileError, err = getResult(name)
		// 	})
		// 	It("should still exists", func() {
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})
		// 	It("should not return an error", func() {
		// 		Expect(reconcileError).NotTo(HaveOccurred())
		// 	})
		// 	It("should still be in Running state", func() {
		// 		Expect(run.Status.State).To(Equal("Running"))
		// 	})
		// 	It("should set RequeueAfter to WaitAction", func() {
		// 		Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
		// 	})
		// 	It("should not create any new pod", func() {
		// 		pods, err := getLinkedPods(k8sClient, run, controller.PlanAction, name.Namespace)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(len(pods.Items)).To(Equal(1))
		// 	})
		// })
		// Describe("When a TerraformRun has a completed pod", Ordered, func() {
		// 	BeforeAll(func() {
		// 		name = types.NamespacedName{
		// 			Name:      "nominal-case-3",
		// 			Namespace: "default",
		// 		}
		// 		result, run, reconcileError, err = getResult(name)
		// 	})
		// 	It("should still exists", func() {
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})
		// 	It("should not return an error", func() {
		// 		Expect(reconcileError).NotTo(HaveOccurred())
		// 	})
		// 	It("should end in Succeeded state", func() {
		// 		Expect(run.Status.State).To(Equal("Succeeded"))
		// 	})
		// 	It("should set RequeueAfter to DriftDetection", func() {
		// 		Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.DriftDetection))
		// 	})
		// 	It("should not have created an apply pod", func() {
		// 		pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(len(pods.Items)).To(Equal(0))
		// 	})
		// })
		// 	Describe("When a TerraformLayer just got applied", Ordered, func() {
		// 		BeforeAll(func() {
		// 			name = types.NamespacedName{
		// 				Name:      "nominal-case-4",
		// 				Namespace: "default",
		// 			}
		// 			result, layer, reconcileError, err = getResult(name)
		// 		})
		// 		It("should still exists", func() {
		// 			Expect(err).NotTo(HaveOccurred())
		// 		})
		// 		It("should not return an error", func() {
		// 			Expect(reconcileError).NotTo(HaveOccurred())
		// 		})
		// 		It("should end in Idle state", func() {
		// 			Expect(layer.Status.State).To(Equal("Idle"))
		// 		})
		// 		It("should not be locked", func() {
		// 			Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeFalse())
		// 		})
		// 		It("should set RequeueAfter to DriftDetection", func() {
		// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.DriftDetection))
		// 		})
		// 		It("should not have created a plan pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(0))
		// 		})
		// 		It("should not have created an apply pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(0))
		// 		})
		// 	})
		// 	Describe("When a TerraformLayer shares a path with another TerraformLayer and an action is already running", Ordered, func() {
		// 		BeforeAll(func() {
		// 			name = types.NamespacedName{
		// 				Name:      "nominal-case-5",
		// 				Namespace: "default",
		// 			}
		// 			result, layer, reconcileError, err = getResult(name)
		// 		})
		// 		It("should still exists", func() {
		// 			Expect(err).NotTo(HaveOccurred())
		// 		})
		// 		It("should not return an error", func() {
		// 			Expect(reconcileError).NotTo(HaveOccurred())
		// 		})
		// 		It("should not update status", func() {
		// 			Expect(layer.Status.State).To(Equal(""))
		// 		})
		// 		It("should be locked", func() {
		// 			Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
		// 		})
		// 		It("should set RequeueAfter to WaitAction", func() {
		// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
		// 		})
		// 		It("should not have created a plan pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(0))
		// 		})
		// 		It("should not have created an apply pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(0))
		// 		})
		// 	})
		// 	Describe("When a TerraformLayer hasn't been planned since more time than the DriftDetection period", Ordered, func() {
		// 		BeforeAll(func() {
		// 			name = types.NamespacedName{
		// 				Name:      "nominal-case-6",
		// 				Namespace: "default",
		// 			}
		// 			result, layer, reconcileError, err = getResult(name)
		// 		})
		// 		It("should still exists", func() {
		// 			Expect(err).NotTo(HaveOccurred())
		// 		})
		// 		It("should not return an error", func() {
		// 			Expect(reconcileError).NotTo(HaveOccurred())
		// 		})
		// 		It("should be in PlanNeeded state", func() {
		// 			Expect(layer.Status.State).To(Equal("PlanNeeded"))
		// 		})
		// 		It("should be locked", func() {
		// 			Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
		// 		})
		// 		It("should set RequeueAfter to WaitAction", func() {
		// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
		// 		})
		// 		It("should have created a plan pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(1))
		// 		})
		// 		It("should not have created an apply pod", func() {
		// 			pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(len(pods.Items)).To(Equal(0))
		// 		})
		// 	})
	})
	// Describe("Error Case", func() {
	// 	Describe("When a TerraformLayer has errored once on plan and still in grace period", Ordered, func() {
	// 		BeforeAll(func() {
	// 			name = types.NamespacedName{
	// 				Name:      "error-case-1",
	// 				Namespace: "default",
	// 			}
	// 			result, layer, reconcileError, err = getResult(name)
	// 		})
	// 		It("should still exists", func() {
	// 			Expect(err).NotTo(HaveOccurred())
	// 		})
	// 		It("should not return an error", func() {
	// 			Expect(reconcileError).NotTo(HaveOccurred())
	// 		})
	// 		It("should end in FailureGracePeriod state", func() {
	// 			Expect(layer.Status.State).To(Equal("FailureGracePeriod"))
	// 		})
	// 		It("should set RequeueAfter to WaitAction", func() {
	// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
	// 		})
	// 	})
	// 	Describe("When a TerraformLayer has errored once on apply and still in grace period", Ordered, func() {
	// 		BeforeAll(func() {
	// 			name = types.NamespacedName{
	// 				Name:      "error-case-2",
	// 				Namespace: "default",
	// 			}
	// 			result, layer, reconcileError, err = getResult(name)
	// 		})
	// 		It("should still exists", func() {
	// 			Expect(err).NotTo(HaveOccurred())
	// 		})
	// 		It("should not return an error", func() {
	// 			Expect(reconcileError).NotTo(HaveOccurred())
	// 		})
	// 		It("should end in FailureGracePeriod state", func() {
	// 			Expect(layer.Status.State).To(Equal("FailureGracePeriod"))
	// 		})
	// 		It("should set RequeueAfter to WaitAction", func() {
	// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
	// 		})
	// 	})
	// })
	// Describe("When a TerraformLayer has errored once on plan and not in grace period anymore", Ordered, func() {
	// 	BeforeAll(func() {
	// 		name = types.NamespacedName{
	// 			Name:      "error-case-3",
	// 			Namespace: "default",
	// 		}
	// 		result, layer, reconcileError, err = getResult(name)
	// 	})
	// 	It("should still exists", func() {
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	// 	It("should not return an error", func() {
	// 		Expect(reconcileError).NotTo(HaveOccurred())
	// 	})
	// 	It("should end in PlanNeeded state", func() {
	// 		Expect(layer.Status.State).To(Equal("PlanNeeded"))
	// 	})
	// 	It("should have the condition IsPlanArtifactUpToDate set to false", func() {
	// 		Expect(layer.Status.Conditions[0].Reason).To(Equal("LastPlanFailed"))
	// 		Expect(string(layer.Status.Conditions[0].Status)).To(Equal("False"))
	// 	})
	// 	It("should set RequeueAfter to WaitAction", func() {
	// 		Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
	// 	})
	// 	It("should be locked", func() {
	// 		Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
	// 	})
	// 	It("should have created a plan pod", func() {
	// 		pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(len(pods.Items)).To(Equal(1))
	// 	})
	// })
	// Describe("When a TerraformLayer has errored once on apply and not in grace period anymore", Ordered, func() {
	// 	BeforeAll(func() {
	// 		name = types.NamespacedName{
	// 			Name:      "error-case-4",
	// 			Namespace: "default",
	// 		}
	// 		result, layer, reconcileError, err = getResult(name)
	// 	})
	// 	It("should still exists", func() {
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	// 	It("should not return an error", func() {
	// 		Expect(reconcileError).NotTo(HaveOccurred())
	// 	})
	// 	It("should end in ApplyNeeded state", func() {
	// 		Expect(layer.Status.State).To(Equal("ApplyNeeded"))
	// 	})
	// 	It("should set RequeueAfter to WaitAction", func() {
	// 		Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
	// 	})
	// 	It("should be locked", func() {
	// 		Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
	// 	})
	// 	It("should have created an apply pod", func() {
	// 		pods, err := getLinkedPods(k8sClient, layer, controller.ApplyAction, name.Namespace)
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(len(pods.Items)).To(Equal(1))
	// 	})
	// })
	// Describe("Merge case", func() {
	// 	Describe("When a TerraformLayer is created", Ordered, func() {
	// 		BeforeAll(func() {
	// 			name = types.NamespacedName{
	// 				Name:      "merge-case-1",
	// 				Namespace: "default",
	// 			}
	// 			result, layer, reconcileError, err = getResult(name)
	// 		})
	// 		It("should still exists", func() {
	// 			Expect(err).NotTo(HaveOccurred())
	// 		})
	// 		It("should not return an error", func() {
	// 			Expect(reconcileError).NotTo(HaveOccurred())
	// 		})
	// 		It("should end in PlanNeeded state", func() {
	// 			Expect(layer.Status.State).To(Equal("PlanNeeded"))
	// 		})
	// 		It("should be locked", func() {
	// 			Expect(lock.IsLocked(context.TODO(), k8sClient, layer)).To(BeTrue())
	// 		})
	// 		It("should set RequeueAfter to WaitAction", func() {
	// 			Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
	// 		})
	// 		It("should have created a plan pod", func() {
	// 			pods, err := getLinkedPods(k8sClient, layer, controller.PlanAction, name.Namespace)
	// 			Expect(err).NotTo(HaveOccurred())
	// 			Expect(len(pods.Items)).To(Equal(1))
	// 		})
	// 	})
	// })
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// func TestGetLayerExponentialBackOffTime(t *testing.T) {
// 	tt := []struct {
// 		name         string
// 		defaultTime  time.Duration
// 		layer        *configv1alpha1.TerraformLayer
// 		expectedTime time.Duration
// 	}{
// 		{
// 			"Exponential backoff : No retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			time.Minute,
// 		},
// 		{
// 			"Exponential backoff : Success",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "0"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 1 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "1"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			2 * time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 2 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "2"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			7 * time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 3 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "3"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			20 * time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 5 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "5"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			148 * time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 10 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "10"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			22026 * time.Minute,
// 		},
// 		{
// 			"Exponential backoff : 17 retry",
// 			time.Minute,
// 			&configv1alpha1.TerraformLayer{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Annotations: map[string]string{"runner.terraform.padok.cloud/failure": "17"},
// 				},
// 				Spec: configv1alpha1.TerraformLayerSpec{
// 					TerraformConfig: configv1alpha1.TerraformConfig{
// 						Version: "1.0.1",
// 					},
// 				},
// 			},
// 			24154952 * time.Minute,
// 		},
// 	}

// 	for _, tc := range tt {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result := controller.GetLayerExponentialBackOffTime(tc.defaultTime, tc.layer)
// 			if tc.expectedTime != result {
// 				t.Errorf("different version computed: expected %s go %s", tc.expectedTime, result)
// 			}
// 		})
// 	}
// }
