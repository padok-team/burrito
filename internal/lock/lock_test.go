package lock_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/lock"
	utils "github.com/padok-team/burrito/internal/testing"
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

func TestLock(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Lock Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../..", "manifests", "crds")},
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

var _ = Describe("Lock", func() {
	var layer *configv1alpha1.TerraformLayer
	var run *configv1alpha1.TerraformRun
	var getErrLayer error
	var getErrRun error
	Describe("Add check remove flow", Ordered, func() {
		BeforeAll(func() {
			layer = &configv1alpha1.TerraformLayer{}
			getErrLayer = k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			}, layer)
			run = &configv1alpha1.TerraformRun{}
			getErrRun = k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test-run",
			}, run)
		})
		It("layer and run should exist", func() {
			Expect(getErrLayer).NotTo(HaveOccurred())
			Expect(getErrRun).NotTo(HaveOccurred())
		})
		It("should return false since layer is not locked", func() {
			locked, err := lock.IsLayerLocked(context.TODO(), k8sClient, layer)
			Expect(err).NotTo(HaveOccurred())
			Expect(locked).To(Equal(false))
		})
		It("should not return error when creating Lease object", func() {
			err := lock.CreateLock(context.TODO(), k8sClient, layer, run)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return true since layer is locked", func() {
			locked, err := lock.IsLayerLocked(context.TODO(), k8sClient, layer)
			Expect(err).NotTo(HaveOccurred())
			Expect(locked).To(Equal(true))
		})
		It("should not return error when deleting Lease object", func() {
			err := lock.DeleteLock(context.TODO(), k8sClient, layer, run)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return false since layer is not locked anymore", func() {
			locked, err := lock.IsLayerLocked(context.TODO(), k8sClient, layer)
			Expect(err).NotTo(HaveOccurred())
			Expect(locked).To(Equal(false))
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
