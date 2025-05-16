package credentials_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/repository/credentials"
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
var credentialStore *credentials.CredentialStore

func TestCredentials(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Credentials Suite")
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
	Expect(k8sClient).NotTo(BeNil())
	utils.LoadResources(k8sClient, "testdata")
	credentialStore = credentials.NewCredentialStore(
		k8sClient,
		5*time.Second,
	)
})

var _ = Describe("Credentials", func() {
	Describe("Repository secret is present", Ordered, func() {
		It("should return repository secret", func() {
			repository := &configv1alpha1.TerraformRepository{}
			err := k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      "repository-secret-present",
				Namespace: "default",
			}, repository)
			Expect(err).NotTo(HaveOccurred())
			credentials, err := credentialStore.GetCredentials(repository)
			fmt.Println(credentials.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(credentials.Username).To(Equal("username-present"))
			Expect(credentials.Password).To(Equal("password-present"))
		})
	})
	Describe("Repository secret is not present", Ordered, func() {
		Describe("Shared secret is present", Ordered, func() {
			It("should return shared secret", func() {
				repository := &configv1alpha1.TerraformRepository{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repository-secret-not-present",
					Namespace: "default",
				}, repository)
				Expect(err).NotTo(HaveOccurred())
				credentials, err := credentialStore.GetCredentials(repository)
				Expect(err).NotTo(HaveOccurred())
				Expect(credentials.Username).To(Equal("username-shared"))
				Expect(credentials.Password).To(Equal("password-shared"))
			})
		})
		Describe("Shared secret is not present", Ordered, func() {
			It("should return error", func() {
				repository := &configv1alpha1.TerraformRepository{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "no-secret-present",
					Namespace: "default",
				}, repository)
				Expect(err).NotTo(HaveOccurred())
				_, err = credentialStore.GetCredentials(repository)
				Expect(err).To(HaveOccurred())
			})
		})
		Describe("Shared secret is present but not allowed", Ordered, func() {
			It("should return error", func() {
				repository := &configv1alpha1.TerraformRepository{}
				k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "not-allowed-secret",
					Namespace: "default",
				}, repository)
				_, err := credentialStore.GetCredentials(repository)
				Expect(err).To(HaveOccurred())
			})
		})
		Describe("Two shared secrets are present", Ordered, func() {
			It("should return the one with the longest URL", func() {
				repository := &configv1alpha1.TerraformRepository{}
				k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "two-shared-secret-match",
					Namespace: "default",
				}, repository)
				credentials, err := credentialStore.GetCredentials(repository)
				Expect(err).NotTo(HaveOccurred())
				Expect(credentials.Username).To(Equal("username-match-1"))
				Expect(credentials.Password).To(Equal("password-match-1"))
			})
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
