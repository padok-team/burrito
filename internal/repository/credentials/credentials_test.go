package credentials_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
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

const concurrentRepositoryUrl = "https://github.com/padok-team/burrito"

// listCountingClient counts List calls so specs can assert how many refreshes a
// burst of concurrent readers actually triggered.
type listCountingClient struct {
	client.Client
	lists atomic.Int64
}

func (c *listCountingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	c.lists.Add(1)
	return c.Client.List(ctx, list, opts...)
}

// newConcurrentTestClient builds an in-memory client holding one repository and
// one shared credential, both matching concurrentTestRepository(). The
// concurrency specs drive tens of thousands of List calls, which is far more
// than the envtest API server used by the rest of this suite can serve.
func newConcurrentTestClient() *listCountingClient {
	testScheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(testScheme)).To(Succeed())

	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "repo-creds", Namespace: "default"},
		Type:       corev1.SecretType(credentials.CredentialsType),
		Data: map[string][]byte{
			"provider": []byte("github"),
			"url":      []byte(concurrentRepositoryUrl),
			"username": []byte("repository-user"),
			"password": []byte("repository-password"),
		},
	}
	sharedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "shared-creds",
			Namespace:   "default",
			Annotations: map[string]string{annotations.AllowedTenants: "default"},
		},
		Type: corev1.SecretType(credentials.SharedCredentialsType),
		Data: map[string][]byte{
			"provider": []byte("github"),
			"url":      []byte("https://github.com/padok-team"),
			"username": []byte("shared-user"),
			"password": []byte("shared-password"),
		},
	}

	return &listCountingClient{
		Client: fake.NewClientBuilder().
			WithScheme(testScheme).
			WithObjects(repositorySecret, sharedSecret).
			WithIndex(&corev1.Secret{}, "type", func(obj client.Object) []string {
				return []string{string(obj.(*corev1.Secret).Type)}
			}).
			Build(),
	}
}

func concurrentTestRepository() *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: concurrentRepositoryUrl,
			},
		},
	}
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
	Describe("Concurrent access", func() {
		const (
			readers    = 32
			iterations = 100
		)

		It("should serve consistent credentials while refreshing concurrently", func() {
			// A TTL of 0 makes every call refresh, so reads
			// (GetCredentials/GetAllCredentials) and the refresh they trigger are
			// guaranteed to overlap across goroutines. The counters below fail on a
			// partially observed refresh even without the race detector; run with
			// `go test -race` to also catch unsynchronized access itself.
			store := credentials.NewCredentialStore(newConcurrentTestClient(), 0)
			repository := concurrentTestRepository()

			var lookupErrors, unexpectedCredential, unexpectedCounts atomic.Int64
			var wg sync.WaitGroup
			for i := 0; i < readers; i++ {
				wg.Add(2)
				go func() {
					defer wg.Done()
					for j := 0; j < iterations; j++ {
						credential, err := store.GetCredentials(repository)
						if err != nil {
							lookupErrors.Add(1)
							continue
						}
						// The repository credential is more specific than the shared
						// one, so it must win on every single call.
						if credential.Username != "repository-user" {
							unexpectedCredential.Add(1)
						}
					}
				}()
				go func() {
					defer wg.Done()
					for j := 0; j < iterations; j++ {
						shared, repos := store.GetAllCredentials()
						if len(shared) != 1 || len(repos) != 1 {
							unexpectedCounts.Add(1)
						}
					}
				}()
			}
			wg.Wait()

			Expect(lookupErrors.Load()).To(BeZero(), "GetCredentials stopped matching the repository credential")
			Expect(unexpectedCredential.Load()).To(BeZero(), "GetCredentials returned a credential from an incomplete snapshot")
			Expect(unexpectedCounts.Load()).To(BeZero(), "GetAllCredentials returned an incomplete snapshot")
		})

		It("should refresh once when concurrent readers find the cache stale", func() {
			// Cold store, long TTL: every reader below sees a stale cache, but only
			// one of them should reach the API. A single refresh is two List calls
			// (shared secrets, then repository secrets).
			concurrentClient := newConcurrentTestClient()
			store := credentials.NewCredentialStore(concurrentClient, time.Hour)
			repository := concurrentTestRepository()

			var wg sync.WaitGroup
			for i := 0; i < readers; i++ {
				wg.Add(2)
				go func() {
					defer wg.Done()
					_, _ = store.GetCredentials(repository)
				}()
				go func() {
					defer wg.Done()
					store.GetAllCredentials()
				}()
			}
			wg.Wait()

			Expect(concurrentClient.lists.Load()).To(BeEquivalentTo(2), "concurrent stale readers should share a single refresh")
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
