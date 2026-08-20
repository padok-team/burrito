package credentials_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/repository/credentials"
	utils "github.com/padok-team/burrito/internal/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestCredentialStoreConcurrentGetters(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 to scheme: %v", err)
	}

	repository := &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repo",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://github.com/padok-team/burrito",
			},
		},
	}
	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repo-creds",
			Namespace: "default",
		},
		Type: corev1.SecretType(credentials.CredentialsType),
		Data: map[string][]byte{
			"provider": []byte("github"),
			"url":      []byte("https://github.com/padok-team/burrito"),
			"username": []byte("repository-user"),
			"password": []byte("repository-password"),
		},
	}
	sharedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-creds",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.AllowedTenants: "default",
			},
		},
		Type: corev1.SecretType(credentials.SharedCredentialsType),
		Data: map[string][]byte{
			"provider": []byte("github"),
			"url":      []byte("https://github.com/padok-team"),
			"username": []byte("shared-user"),
			"password": []byte("shared-password"),
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repositorySecret, sharedSecret).
		WithIndex(&corev1.Secret{}, "type", func(obj client.Object) []string {
			return []string{string(obj.(*corev1.Secret).Type)}
		}).
		Build()
	store := credentials.NewCredentialStore(client, 0)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := store.GetCredentials(repository); err != nil {
					t.Errorf("GetCredentials returned error: %v", err)
					return
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				shared, repository := store.GetAllCredentials()
				if len(shared) != 1 || len(repository) != 1 {
					t.Errorf("GetAllCredentials returned %d shared and %d repository credentials", len(shared), len(repository))
					return
				}
			}
		}()
	}
	wg.Wait()
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
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "not-allowed-secret",
					Namespace: "default",
				}, repository)
				Expect(err).NotTo(HaveOccurred())
				_, err = credentialStore.GetCredentials(repository)
				Expect(err).To(HaveOccurred())
			})
		})
		Describe("Two shared secrets are present", Ordered, func() {
			It("should return the one with the longest URL", func() {
				repository := &configv1alpha1.TerraformRepository{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "two-shared-secret-match",
					Namespace: "default",
				}, repository)
				Expect(err).NotTo(HaveOccurred())
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
