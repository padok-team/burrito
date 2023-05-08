// // /*
// // Copyright 2022.

// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at

// //     http://www.apache.org/licenses/LICENSE-2.0

// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.
// // */

package terraformpullrequest

// import (
// 	"path/filepath"
// 	"testing"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"

// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes/scheme"
// 	"k8s.io/client-go/rest"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/envtest"
// 	logf "sigs.k8s.io/controller-runtime/pkg/log"
// 	"sigs.k8s.io/controller-runtime/pkg/log/zap"

// 	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
// 	"github.com/padok-team/burrito/internal/annotations"
// 	//+kubebuilder:scaffold:imports
// )

// // These tests use Ginkgo (BDD-style Go testing framework). Refer to
// // http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// var cfg *rest.Config
// var k8sClient client.Client
// var testEnv *envtest.Environment
// var reconciler *Reconciler

// func TestPullRequest(t *testing.T) {
// 	RegisterFailHandler(Fail)

// 	RunSpecs(t, "TerraformPullRequest Controller Suite")
// }

// var _ = BeforeSuite(func() {
// 	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

// 	By("bootstrapping test environment")
// 	testEnv = &envtest.Environment{
// 		CRDDirectoryPaths:     []string{filepath.Join("../../..", "config", "crd", "bases")},
// 		ErrorIfCRDPathMissing: true,
// 	}

// 	var err error
// 	// cfg is defined in this file globally.
// 	cfg, err = testEnv.Start()
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(cfg).NotTo(BeNil())

// 	err = configv1alpha1.AddToScheme(scheme.Scheme)
// 	Expect(err).NotTo(HaveOccurred())

// 	//+kubebuilder:scaffold:scheme

// 	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
// 	utils.LoadResources(k8sClient, "../testdata")
// 	reconciler = &controller.Reconciler{
// 		Client:  k8sClient,
// 		Scheme:  scheme.Scheme,
// 		Storage: storage.New(),
// 	}
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(k8sClient).NotTo(BeNil())

// })

// var _ = Describe("TerraformPullRequest controller", func() {
// 	var pr *configv1alpha1.TerraformPullRequest

// 	BeforeEach(func() {
// 		pr = &configv1alpha1.TerraformPullRequest{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name:      "test",
// 				Namespace: "test",
// 				Annotations: map[string]string{
// 					annotations.LastBranchCommit: "04410b5b7d90b82ad658b86564a9aa4bce411ac9",
// 				},
// 			},
// 			Spec: configv1alpha1.TerraformPullRequestSpec{
// 				Provider: "gitlab",
// 				Branch:   "test",
// 				Repository: configv1alpha1.TerraformLayerRepository{
// 					Name:      "test-repository",
// 					Namespace: "default",
// 				},
// 			},
// 		}
// 	})

// 	Describe("Conditions", func() {
// 		Describe("IsLastCommitDiscovered When PR", func() {
// 			Context("Has no discovered commit annotation", func() {
// 				It("Should return false", func() {
// 					_, value := reconciler.IsLastCommitDiscovered(pr)
// 					Expect(value).To(BeFalse())
// 				})
// 			})
// 			Context("Has discovered commit annotation", func() {
// 				It("Should return true", func() {
// 					pr.Annotations[annotations.LastDiscoveredCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
// 					_, value := reconciler.IsLastCommitDiscovered(pr)
// 					Expect(value).To(BeTrue())
// 				})
// 			})
// 			Context("Has no branch commit annotation", func() {
// 				It("Should return false", func() {
// 					delete(pr.Annotations, annotations.LastBranchCommit)
// 					_, value := reconciler.IsLastCommitDiscovered(pr)
// 					Expect(value).To(BeFalse())
// 				})
// 			})
// 		})
// 		Describe("IsCommentUpToDate When PR", func() {
// 			Context("Has no comment annotation", func() {
// 				Context("Has no discovered annotation", func() {
// 					It("Should return true", func() {
// 						_, value := reconciler.IsCommentUpToDate(pr)
// 						Expect(value).To(BeTrue())
// 					})
// 				})
// 				Context("Has discovered annotation", func() {
// 					It("Should return false", func() {
// 						pr.Annotations[annotations.LastDiscoveredCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
// 						_, value := reconciler.IsCommentUpToDate(pr)
// 						Expect(value).To(BeFalse())
// 					})
// 				})
// 			})
// 			Context("Has discovered annotation and commented annotation equals", func() {
// 				It("Should return true", func() {
// 					pr.Annotations[annotations.LastDiscoveredCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
// 					pr.Annotations[annotations.LastCommentedCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
// 					_, value := reconciler.IsCommentUpToDate(pr)
// 					Expect(value).To(BeTrue())
// 				})
// 			})
// 			Context("Has discovered annotation and commented annotation different", func() {
// 				It("Should return false", func() {
// 					pr.Annotations[annotations.LastDiscoveredCommit] = "04410b5b7d90b82ad658b86564a9aa4bce411ac9"
// 					pr.Annotations[annotations.LastCommentedCommit] = "old"
// 					_, value := reconciler.IsCommentUpToDate(pr)
// 					Expect(value).To(BeFalse())
// 				})
// 			})
// 		})
// 		Describe("AreLayersStillPlanning", func() {
// 			var layer *configv1alpha1.TerraformLayer
// 			BeforeEach(func() {
// 				layer = &configv1alpha1.TerraformLayer{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Name:      "test",
// 						Namespace: "test",
// 					},
// 					Spec: configv1alpha1.TerraformLayerSpec{
// 						Repository: configv1alpha1.TerraformLayerRepository{
// 							Name:      "test-repository",
// 							Namespace: "default",
// 						},
// 					},
// 				}
// 			})
// 			Context("No Layers", func() {
// 				It("Should return false", func() {
// 					_, value := reconciler.AreLayersStillPlanning(pr)
// 					Expect(value).To(BeFalse())
// 				})
// 			})
// 			Context("Single Layer", func() {
// 				Context("When layer plan annotation is not set", func() {
// 					It("Should return true", func() {
// 						k8sClient.Create(context.Background(), &configv1alpha1.TerraformLayer{
// 							ObjectMeta: metav1.ObjectMeta{
// 								Name:      "test",
// 								Namespace: "test",
// 							},
// 						Expect(value).To(BeTrue())
// 					})
// 				})
// 			})
// 		})

// 	})
// })

// var _ = AfterSuite(func() {
// 	By("tearing down the test environment")
// 	err := testEnv.Stop()
// 	Expect(err).NotTo(HaveOccurred())
// })
