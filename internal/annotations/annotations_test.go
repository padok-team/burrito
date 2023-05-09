package annotations_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	utils "github.com/padok-team/burrito/internal/controllers/testing"
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

func TestAnnotations(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Annotations Controller Suite")
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

var _ = Describe("Annotations", func() {
	var layer *configv1alpha1.TerraformLayer
	var getErr error
	Describe("Add/Remove annotations", Ordered, func() {
		BeforeAll(func() {
			layer = &configv1alpha1.TerraformLayer{}
			getErr = k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, layer)
		})
		It("should exists", func() {
			Expect(getErr).NotTo(HaveOccurred())
		})
		It("should not return an error when adding first annotation", func() {
			err := annotations.Add(context.TODO(), k8sClient, layer, map[string]string{annotations.LastPlanSum: "AuP6pMNxWsbSZKnxZvxD842wy0qaF9JCX8HW1nFeL1I"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should not return an error when adding second annotation", func() {
			err := annotations.Add(context.TODO(), k8sClient, layer, map[string]string{annotations.LastApplySum: "AuP6pMNxWsbSZKnxZvxD842wy0qaF9JCX8HW1nFeL1I"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should not return an error when removing second annotation", func() {
			err := annotations.Remove(context.TODO(), k8sClient, layer, annotations.LastApplySum)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
