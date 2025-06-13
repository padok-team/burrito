package terraformrepository_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	controller "github.com/padok-team/burrito/internal/controllers/terraformrepository"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/repository/credentials"
	mock "github.com/padok-team/burrito/internal/repository/providers/mock"
	utils "github.com/padok-team/burrito/internal/testing"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

func TestLayer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "TerraformRun Controller Suite")
}

type MockClock struct{}

const testTime = "Mon May  8 11:21:53 UTC 2023"

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
		Client:      k8sClient,
		Scheme:      scheme.Scheme,
		Config:      config.TestConfig(),
		Clock:       &MockClock{},
		Datastore:   datastore.NewMockClient(),
		Credentials: credentials.NewCredentialStore(k8sClient, config.TestConfig().Controller.Timers.CredentialsTTL),
		Recorder: record.NewBroadcasterForTests(1*time.Second).NewRecorder(scheme.Scheme, corev1.EventSource{
			Component: "burrito",
		}),
	}

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

func getResult(name types.NamespacedName) (reconcile.Result, *configv1alpha1.TerraformRepository, error, error) {
	result, reconcileError := reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: name,
	})
	repo := &configv1alpha1.TerraformRepository{}
	err := k8sClient.Get(context.TODO(), name, repo)
	return result, repo, reconcileError, err
}

var _ = Describe("Run", func() {
	var repo *configv1alpha1.TerraformRepository
	var reconcileError error
	var err error
	var result reconcile.Result
	var name types.NamespacedName

	Describe("Nominal Case", func() {
		Describe("When a TerraformRepository with one TerraformLayer is created", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-never-synced",
					Namespace: "default",
				}
				result, repo, reconcileError, err = getResult(name)
			})
			It("should still exists", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should not return an error", func() {
				Expect(reconcileError).NotTo(HaveOccurred())
			})
			It("should end in SyncNeeded state", func() {
				Expect(repo.Status.State).To(Equal("SyncNeeded"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches[0].Name).To(Equal("branch-not-in-datastore"))
				Expect(repo.Status.Branches[0].LastSyncStatus).To(Equal("success"))
				Expect(repo.Status.Branches[0].LatestRev).To(Equal(mock.MOCK_REVISION))
				Expect(repo.Status.Branches[0].LastSyncDate).To(Equal(testTime))
			})
			It("should update the annotations of the TerraformLayer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-never-synced-layer",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.MOCK_REVISION))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.MOCK_REVISION))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))
			})
			It("should have put the bundle in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-not-in-datastore", mock.MOCK_REVISION)
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})

		})
	})
	// TODO
	// Describe("Error Case", func() {
	// })
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
