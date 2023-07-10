package event_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	utils "github.com/padok-team/burrito/internal/testing"
	"github.com/padok-team/burrito/internal/webhook/event"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestLayer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Handler Suite")
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
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var PushEventNoChanges = event.PushEvent{
	URL:      "https://github.com/padok-team/burrito-examples",
	Revision: "main",
	ChangeInfo: event.ChangeInfo{
		ShaBefore: "b3231e8771591b3864b3c582e85955c1f76aaded",
		ShaAfter:  "6c193d9cad1ddafdb31ff9f733630da9705bfd64",
	},
	Changes: []string{
		"README.md",
	},
}

var PushEventLayerPathChanges = event.PushEvent{
	URL:      "https://github.com/padok-team/burrito-examples",
	Revision: "main",
	ChangeInfo: event.ChangeInfo{
		ShaBefore: "b3231e8771591b3864b3c582e85955c1f76aaded",
		ShaAfter:  "6c193d9cad1ddafdb31ff9f733630da9705bfd64",
	},
	Changes: []string{
		"layer-path-changed/main.tf",
	},
}

var PushEventAdditionalPathChanges = event.PushEvent{
	URL:      "https://github.com/padok-team/burrito-examples",
	Revision: "main",
	ChangeInfo: event.ChangeInfo{
		ShaBefore: "b3231e8771591b3864b3c582e85955c1f76aaded",
		ShaAfter:  "6c193d9cad1ddafdb31ff9f733630da9705bfd64",
	},
	Changes: []string{
		"modules/module-changed/variables.tf",
		"terragrunt/layer-path-changed/module.hcl",
	},
}

var PushEventMultiplePathChanges = event.PushEvent{
	URL:      "https://github.com/padok-team/burrito-examples",
	Revision: "main",
	ChangeInfo: event.ChangeInfo{
		ShaBefore: "b3231e8771591b3864b3c582e85955c1f76aaded",
		ShaAfter:  "6c193d9cad1ddafdb31ff9f733630da9705bfd64",
	},
	Changes: []string{
		"layer-path-changed-2/variables.tf",
		"layer-path-changed-3/inputs.hcl",
	},
}

// var PullRequestEventNotAffected = event.PullRequestEvent{
// 	Provider: "github",
// 	URL:      "https://github.com/example/repo",
// 	Revision: "feature/branch",
// 	Base:     "main",
// 	Action:   "opened",
// 	ID:       "42",
// 	Commit:   "5b2c5e5c6699bf2bf93138205565b85193996572",
// }

var PullRequestEventSingleAffected = event.PullRequestEvent{
	Provider: "github",
	URL:      "https://github.com/padok-team/burrito-examples",
	Revision: "feature/branch",
	Base:     "main",
	Action:   "opened",
	ID:       "42",
	Commit:   "5b2c5e5c6699bf2bf93138205565b85193996572",
}

var PullRequestEventMultipleAffected = event.PullRequestEvent{
	Provider: "github",
	URL:      "https://github.com/example/other-repo",
	Revision: "feature/branch",
	Base:     "main",
	Action:   "opened",
	ID:       "42",
	Commit:   "5b2c5e5c6699bf2bf93138205565b85193996572",
}

var _ = Describe("Webhook", func() {
	var handleErr error
	Describe("Push Event", func() {
		Describe("Layer", func() {
			Describe("No paths are relevant to layer", Ordered, func() {
				BeforeAll(func() {
					handleErr = PushEventNoChanges.Handle(k8sClient)
				})
				It("should have only set the LastBranchCommit annotation", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "no-path-changed-1",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					_, ok := layer.Annotations[annotations.LastRelevantCommit]
					Expect(ok).To(BeFalse())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventNoChanges.ChangeInfo.ShaAfter))
				})
				It("should not have changed the LastRelevantCommit annotation", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "no-path-changed-2",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Not(Equal(PushEventNoChanges.ChangeInfo.ShaAfter)))
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventNoChanges.ChangeInfo.ShaAfter))
				})
			})
			Describe("Layer path has been modified", Ordered, func() {
				BeforeAll(func() {
					handleErr = PushEventLayerPathChanges.Handle(k8sClient)
				})
				It("should have updated the LastBranchCommit and LastRelevantCommit annotations", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "layer-path-changed-1",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventLayerPathChanges.ChangeInfo.ShaAfter))
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Equal(PushEventLayerPathChanges.ChangeInfo.ShaAfter))
				})
			})
			Describe("Additional path has been modified", Ordered, func() {
				BeforeAll(func() {
					handleErr = PushEventAdditionalPathChanges.Handle(k8sClient)
				})
				It("should have updated commit annotations for a absolute change path", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "layer-additional-paths-1",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventAdditionalPathChanges.ChangeInfo.ShaAfter))
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Equal(PushEventAdditionalPathChanges.ChangeInfo.ShaAfter))
				})
				// TODO: make this test pass
				It("should have updated commit annotations for a relative change path", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "layer-additional-paths-2",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventAdditionalPathChanges.ChangeInfo.ShaAfter))
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Equal(PushEventAdditionalPathChanges.ChangeInfo.ShaAfter))
				})
			})
			Describe("Multiple paths have been modified", Ordered, func() {
				BeforeAll(func() {
					handleErr = PushEventMultiplePathChanges.Handle(k8sClient)
				})
				It("should have updated commit annotations for layer-path-changed-2", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "layer-path-changed-2",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventMultiplePathChanges.ChangeInfo.ShaAfter))
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Equal(PushEventMultiplePathChanges.ChangeInfo.ShaAfter))
				})
				It("should have updated commit annotations for layer-path-changed-3", func() {
					layer := &configv1alpha1.TerraformLayer{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "layer-path-changed-3",
					}, layer)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(layer.Annotations[annotations.LastBranchCommit]).To(Equal(PushEventMultiplePathChanges.ChangeInfo.ShaAfter))
					Expect(layer.Annotations[annotations.LastRelevantCommit]).To(Equal(PushEventMultiplePathChanges.ChangeInfo.ShaAfter))
				})
			})
		})
		Describe("PullRequest", func() {
			// TODO
			// Describe("No pull request have been created", Ordered, func() {})
			Describe("A single pull request have been affected", Ordered, func() {
				BeforeAll(func() {
					handleErr = PullRequestEventSingleAffected.Handle(k8sClient)
				})
				It("should have created a TerraformPullRequest", func() {
					pr := &configv1alpha1.TerraformPullRequest{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      fmt.Sprintf("%s-%s", "burrito", PullRequestEventSingleAffected.ID),
					}, pr)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(pr.Annotations[annotations.LastBranchCommit]).To(Equal(PullRequestEventSingleAffected.Commit))
				})
			})
			Describe("Multiple pull request have been affected", Ordered, func() {
				BeforeAll(func() {
					handleErr = PullRequestEventMultipleAffected.Handle(k8sClient)
				})
				It("should have created a TerraformPullRequest for other-repo-1", func() {
					pr := &configv1alpha1.TerraformPullRequest{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      fmt.Sprintf("%s-%s", "other-repo-1", PullRequestEventMultipleAffected.ID),
					}, pr)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(pr.Annotations[annotations.LastBranchCommit]).To(Equal(PullRequestEventMultipleAffected.Commit))
				})
				It("should have created a TerraformPullRequest for other-repo-2", func() {
					pr := &configv1alpha1.TerraformPullRequest{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      fmt.Sprintf("%s-%s", "other-repo-2", PullRequestEventMultipleAffected.ID),
					}, pr)
					Expect(err).NotTo(HaveOccurred())
					Expect(handleErr).NotTo(HaveOccurred())
					Expect(pr.Annotations[annotations.LastBranchCommit]).To(Equal(PullRequestEventMultipleAffected.Commit))
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
