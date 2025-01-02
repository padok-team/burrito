/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package terraformpullrequest_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	controller "github.com/padok-team/burrito/internal/controllers/terraformpullrequest"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	utils "github.com/padok-team/burrito/internal/testing"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var reconciler *controller.Reconciler

func TestPullRequest(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "TerraformPullRequest Controller Suite")
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
	Expect(err).NotTo(HaveOccurred())
	utils.LoadResources(k8sClient, "testdata")
	statuses := []StatusUpdate{
		{
			Name:      "pr-nominal-case-3",
			Namespace: "default",
			Status: configv1alpha1.TerraformPullRequestStatus{
				LastDiscoveredCommit: "04410b5b7d90b82ad658b86564a9aa4bce411ac9",
				LastCommentedCommit:  "04410b5b7d90b82ad658b86564a9aa4bce411ac9",
			},
		},
		{
			Name:      "pr-nominal-case-2",
			Namespace: "default",
			Status: configv1alpha1.TerraformPullRequestStatus{
				LastDiscoveredCommit: "04410b5b7d90b82ad658b86564a9aa4bce411ac9",
			},
		},
	}
	err = initStatus(k8sClient, statuses)
	Expect(err).NotTo(HaveOccurred())
	reconciler = &controller.Reconciler{
		Client:    k8sClient,
		Config:    config.TestConfig(),
		Scheme:    scheme.Scheme,
		Datastore: datastore.NewMockClient(),
		Providers: map[string]gitprovider.Provider{
			"mock": func() gitprovider.Provider {
				provider, err := gitprovider.NewWithName(gitprovider.Config{EnableMock: true}, "mock")
				Expect(err).NotTo(HaveOccurred())
				return provider
			}(),
		},
		Recorder: record.NewBroadcasterForTests(1*time.Second).NewRecorder(scheme.Scheme, corev1.EventSource{
			Component: "burrito",
		}),
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

type StatusUpdate struct {
	Name      string
	Namespace string
	Status    configv1alpha1.TerraformPullRequestStatus
}

func updateStatus(c client.Client, name string, namespace string, status configv1alpha1.TerraformPullRequestStatus) error {
	pr := &configv1alpha1.TerraformPullRequest{}
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pr)
	if err != nil {
		return err
	}
	pr.Status = status
	err = c.Status().Update(context.Background(), pr)
	if err != nil {
		return err
	}
	return nil
}

func initStatus(c client.Client, statuses []StatusUpdate) error {
	for _, status := range statuses {
		err := updateStatus(c, status.Name, status.Namespace, status.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

func getResult(name types.NamespacedName) (reconcile.Result, *configv1alpha1.TerraformPullRequest, error, error) {
	result, reconcileError := reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: name,
	})
	pr := &configv1alpha1.TerraformPullRequest{}
	err := k8sClient.Get(context.TODO(), name, pr)
	return result, pr, reconcileError, err
}

var _ = Describe("TerraformPullRequest controller", func() {
	var pr *configv1alpha1.TerraformPullRequest
	var reconcileError error
	var err error
	var result reconcile.Result
	var name types.NamespacedName

	Describe("Conditions", func() {
		BeforeEach(func() {
			pr = &configv1alpha1.TerraformPullRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Annotations: map[string]string{
						annotations.LastBranchCommit: "04410b5b7d90b82ad658b86564a9aa4bce411ac9",
					},
				},
				Spec: configv1alpha1.TerraformPullRequestSpec{
					Branch: "test",
					Repository: configv1alpha1.TerraformLayerRepository{
						Name:      "test-repository",
						Namespace: "default",
					},
				},
			}
		})
		Describe("IsLastCommitDiscovered When PR", func() {
			Context("Has no discovered commit annotation", func() {
				It("Should return false", func() {
					_, value := reconciler.IsLastCommitDiscovered(pr)
					Expect(value).To(BeFalse())
				})
			})
			Context("Has discovered commit annotation", func() {
				It("Should return true", func() {
					pr.Status.LastDiscoveredCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
					_, value := reconciler.IsLastCommitDiscovered(pr)
					Expect(value).To(BeTrue())
				})
			})
			Context("Has no branch commit annotation", func() {
				It("Should return false", func() {
					delete(pr.Annotations, annotations.LastBranchCommit)
					_, value := reconciler.IsLastCommitDiscovered(pr)
					Expect(value).To(BeFalse())
				})
			})
		})
		Describe("IsCommentUpToDate When PR", func() {
			Context("Has no comment annotation", func() {
				Context("Has no discovered annotation", func() {
					It("Should return true", func() {
						_, value := reconciler.IsCommentUpToDate(pr)
						Expect(value).To(BeTrue())
					})
				})
				Context("Has discovered annotation", func() {
					It("Should return false", func() {
						pr.Status.LastDiscoveredCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
						_, value := reconciler.IsCommentUpToDate(pr)
						Expect(value).To(BeFalse())
					})
				})
			})
			Context("Has discovered annotation and commented annotation equals", func() {
				It("Should return true", func() {
					pr.Status.LastDiscoveredCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
					pr.Status.LastCommentedCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
					_, value := reconciler.IsCommentUpToDate(pr)
					Expect(value).To(BeTrue())
				})
			})
			Context("Has discovered annotation and commented annotation different", func() {
				It("Should return false", func() {
					pr.Status.LastDiscoveredCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
					pr.Status.LastCommentedCommit = "old"
					_, value := reconciler.IsCommentUpToDate(pr)
					Expect(value).To(BeFalse())
				})
			})
		})
		Describe("AreLayersStillPlanning", func() {
			var layer *configv1alpha1.TerraformLayer
			BeforeEach(func() {
				pr.Status.LastDiscoveredCommit = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
				layer = &configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-layer",
						Namespace: "default",
						Labels: map[string]string{
							"burrito/managed-by": "test",
						},
						Annotations: map[string]string{},
					},
					Spec: configv1alpha1.TerraformLayerSpec{
						Repository: configv1alpha1.TerraformLayerRepository{
							Name:      "test-repository",
							Namespace: "default",
						},
					},
				}
			})
			Context("PR annotations", func() {
				Context("No discovered commit annotation", func() {
					It("Should return true", func() {
						pr.Status.LastDiscoveredCommit = ""
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(condition.Reason).To(Equal("NoCommitDiscovered"))
						Expect(value).To(BeTrue())
					})
				})
				Context("No branch commit annotation", func() {
					It("Should return false", func() {
						delete(pr.Annotations, annotations.LastBranchCommit)
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(condition.Reason).To(Equal("NoBranchCommitOnPR"))
						Expect(value).To(BeTrue())
					})
				})
				Context("Discovered and Last Branch different", func() {
					It("Should return false", func() {
						pr.Status.LastDiscoveredCommit = "other"
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(condition.Reason).To(Equal("StillNeedsDiscovery"))
						Expect(value).To(BeTrue())
					})
				})
			})
			Context("No Layers", func() {
				It("Should return false", func() {
					condition, value := reconciler.AreLayersStillPlanning(pr)
					Expect(condition.Reason).To(Equal("LayersNotPlanning"))
					Expect(value).To(BeFalse())
				})
			})
			Context("Single Layer", func() {
				Context("When layer plan annotation is not set", func() {
					It("Should return true", func() {
						err := k8sClient.Create(context.Background(), layer)
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(err).NotTo(HaveOccurred())
						Expect(condition.Reason).To(Equal("LayersStillPlanning"))
						Expect(value).To(BeTrue())
						Expect(k8sClient.Delete(context.Background(), layer)).To(Succeed())
					})
				})
				Context("When layer relevant commit annotation is not set", func() {
					It("Should return true", func() {
						layer.Annotations[annotations.LastPlanCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
						err := k8sClient.Create(context.Background(), layer)
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(err).NotTo(HaveOccurred())
						Expect(condition.Reason).To(Equal("NoRelevantCommitOnLayer"))
						Expect(value).To(BeTrue())
						Expect(k8sClient.Delete(context.Background(), layer)).To(Succeed())
					})
				})
				Context("When layer plan & relevant commit are equal", func() {
					It("Should return false", func() {
						layer.Annotations[annotations.LastPlanCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
						layer.Annotations[annotations.LastRelevantCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
						err := k8sClient.Create(context.Background(), layer)
						condition, value := reconciler.AreLayersStillPlanning(pr)
						Expect(err).NotTo(HaveOccurred())
						Expect(condition.Reason).To(Equal("LayersNotPlanning"))
						Expect(value).To(BeFalse())
						Expect(k8sClient.Delete(context.Background(), layer)).To(Succeed())
					})
				})
			})
			// TODO ?
			// Context("Multiple Layers", func() {
			// })
		})
	})
	Describe("Reconcile", func() {
		Describe("Nominal Case", func() {
			Describe("When a TerraformPullRequest is created", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-nominal-case-1",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should end in DiscoveryNeeded state", func() {
					Expect(pr.Status.State).To(Equal("DiscoveryNeeded"))
				})
				It("should set RequeueAfter to WaitAction", func() {
					Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
				})
				It("should have a LastDiscoveredCommit annotation", func() {
					Expect(pr.Status.LastDiscoveredCommit).To(Equal(pr.Annotations[annotations.LastBranchCommit]))
				})
				It("should have created 2 temp layers", func() {
					layers, err := controller.GetLinkedLayers(k8sClient, pr)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(layers)).To(Equal(2))
				})
			})
			Describe("When a TerraformPullRequest has all its layers planned", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-nominal-case-2",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should end in CommentNeeded state", func() {
					Expect(pr.Status.State).To(Equal("CommentNeeded"))
				})
				It("should set RequeueAfter to WaitAction", func() {
					Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
				})
				It("should have a LastCommentedCommit annotation", func() {
					Expect(pr.Status.LastCommentedCommit).To(Equal(pr.Status.LastDiscoveredCommit))
				})
			})
			Describe("When a TerraformPullRequest has all its comment up to date", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-nominal-case-3",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should end in Idle state", func() {
					Expect(pr.Status.State).To(Equal("Idle"))
				})
				It("should set RequeueAfter to WaitAction", func() {
					Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
				})
			})
			Describe("When a TerraformPullRequest with no relevant changes is created", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-nominal-case-4",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should not return an error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should end in DiscoveryNeeded state", func() {
					Expect(pr.Status.State).To(Equal("DiscoveryNeeded"))
				})
				It("should set RequeueAfter to WaitAction", func() {
					Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
				})
				It("should have a LastDiscoveredCommit annotation", func() {
					Expect(pr.Status.LastDiscoveredCommit).To(Equal(pr.Annotations[annotations.LastBranchCommit]))
				})
				It("should not have created temp layers", func() {
					layers, err := controller.GetLinkedLayers(k8sClient, pr)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(layers)).To(Equal(0))
				})
			})
		})
		Describe("Error Case", func() {
			Describe("When a TerraformPullRequest is linked to a non existing repo", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-error-case-1",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should return an empty result error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
					Expect(result.IsZero()).To(BeTrue())
				})
				It("should have no state", func() {
					Expect(pr.Status.State).To(Equal(""))
				})
			})
			Describe("When a TerraformPullRequest has an unknown provider", Ordered, func() {
				BeforeAll(func() {
					name = types.NamespacedName{
						Name:      "pr-error-case-2",
						Namespace: "default",
					}
					result, pr, reconcileError, err = getResult(name)
				})
				It("should still exist", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("should return an empty result error", func() {
					Expect(reconcileError).NotTo(HaveOccurred())
				})
				It("should set RequeueAfter to OnError", func() {
					Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.OnError))
				})
				It("should end in DiscoveryNeeded state", func() {
					Expect(pr.Status.State).To(Equal("DiscoveryNeeded"))
				})
				It("should not have a LastDiscoveredCommit annotation", func() {
					Expect(pr.Status.LastDiscoveredCommit).To(Equal(""))
				})
			})
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
