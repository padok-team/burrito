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
	Expect(err).NotTo(HaveOccurred())
	utils.LoadResources(k8sClient, "testdata")
	statuses := []repoStatusUpdate{
		{
			Name:      "repo-last-sync-too-old",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch-1",
						LastSyncStatus: "success",
						LatestRev:      "OUTDATED_REVISION",
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago
					},
					{
						Name:           "branch-2",
						LastSyncStatus: "success",
						LatestRev:      "OUTDATED_REVISION",
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago
					},
					{
						Name:           "branch-3",
						LastSyncStatus: "success",
						LatestRev:      "OUTDATED_REVISION",
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago
					},
				},
			},
		},
		{
			Name:      "repo-with-new-layer",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   testTime,
					},
				},
			},
		},
		{
			Name:      "repo-synced-recently",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   testTime,
					},
				},
			},
		},
		{
			Name:      "repo-already-last-revision",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago,
					},
				},
			},
		},
		{
			Name:      "repo-without-layers-2",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago,
					},
				},
			},
		},
		{
			Name:      "repo-sync-now",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   "Sun May  7 11:21:53 UTC 2023", // 24 hours ago,
					},
				},
			},
		},
		{
			Name:      "repo-sync-now-old",
			Namespace: "default",
			Status: configv1alpha1.TerraformRepositoryStatus{
				Branches: []configv1alpha1.BranchState{
					{
						Name:           "branch",
						LastSyncStatus: "success",
						LatestRev:      mock.GetMockRevision("branch"),
						LastSyncDate:   testTime,
					},
				},
			},
		},
	}
	err = initStatus(k8sClient, statuses)
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

// Helper struct to update the status of a TerraformRepository resource (because we cannot define status in the YAML files).
// This is used to initialize the status of the TerraformRepository resource in the tests.
type repoStatusUpdate struct {
	Name      string
	Namespace string
	Status    configv1alpha1.TerraformRepositoryStatus
}

func updateStatus(c client.Client, s *repoStatusUpdate) error {
	pr := &configv1alpha1.TerraformRepository{}
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      s.Name,
		Namespace: s.Namespace,
	}, pr)
	if err != nil {
		return err
	}
	pr.Status = s.Status
	err = c.Status().Update(context.Background(), pr)
	if err != nil {
		return err
	}
	return nil
}

func initStatus(c client.Client, statuses []repoStatusUpdate) error {
	for _, status := range statuses {
		err := updateStatus(c, &status)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ = Describe("Run", func() {
	var repo *configv1alpha1.TerraformRepository
	var reconcileError error
	var err error
	var result reconcile.Result
	var name types.NamespacedName

	Describe("Nominal Case", func() {
		Describe("When a TerraformRepository without TerraformLayer is created", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-without-layers",
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
			It("should end in Synced state", func() {
				Expect(repo.Status.State).To(Equal("Synced"))
			})
			It("should have the condition LastSyncTooOld to False with NoBranches reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("NoBranches"))
			})
			It("should have the condition HasLastSyncFailed to False with NoBranches reason", func() {
				Expect(repo.Status.Conditions[1].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[1].Reason).To(Equal("NoBranches"))
			})
			It("should not have branches in the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(0))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository has not TerraformLayers anymore", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-without-layers-2",
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
			It("should end in Synced state", func() {
				Expect(repo.Status.State).To(Equal("Synced"))
			})
			It("should have the condition LastSyncTooOld to False with NoBranches reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("NoBranches"))
			})
			It("should have the condition HasLastSyncFailed to False with NoBranches reason", func() {
				Expect(repo.Status.Conditions[1].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[1].Reason).To(Equal("NoBranches"))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
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
			It("should have the condition LastSyncTooOld to True with NewLayer reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("NewLayer"))
			})
			It("should have the condition HasLastSyncFailed to False with NoSyncYet reason", func() {
				Expect(repo.Status.Conditions[1].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[1].Reason).To(Equal("NoSyncYet"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches[0].Name).To(Equal("branch-not-in-datastore"))
				Expect(repo.Status.Branches[0].LastSyncStatus).To(Equal("success"))
				Expect(repo.Status.Branches[0].LatestRev).To(Equal(mock.GetMockRevision("branch-not-in-datastore")))
				Expect(repo.Status.Branches[0].LastSyncDate).To(Equal(testTime))
			})
			It("should update the annotations of the TerraformLayer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-never-synced-layer",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-not-in-datastore")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-not-in-datastore")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))
			})
			It("should have put the bundle in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-not-in-datastore", mock.GetMockRevision("branch-not-in-datastore"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository with multiple TerraformLayer is created", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-with-multiple-layers",
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
				Expect(repo.Status.Branches).To(HaveLen(2))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch-1",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch-1"),
					LastSyncDate:   testTime,
				}))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch-2",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch-2"),
					LastSyncDate:   testTime,
				}))
			})
			It("should update the annotations of the TerraformLayers", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-with-multiple-layers-branch-1-1",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))

				layer = &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-with-multiple-layers-branch-1-2",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))

				layer = &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-with-multiple-layers-branch-2-1",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-2")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-2")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))

			})
			It("should have put multiple bundles in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-1", mock.GetMockRevision("branch-1"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")

				check, err = reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-2", mock.GetMockRevision("branch-2"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository has not been synced in the last 24h and changes are detected for some layers", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-last-sync-too-old",
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
			It("should have the condition LastSyncTooOld to True", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("SyncTooOld"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(3))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch-1",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch-1"),
					LastSyncDate:   testTime,
				}))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch-2",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch-2"),
					LastSyncDate:   testTime,
				}))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch-3",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch-3"),
					LastSyncDate:   testTime,
				}))
			})
			It("should update the annotations of the Terraform layers WITH changes", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-last-sync-too-old-layer-1",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-1")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))

				layer = &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-last-sync-too-old-layer-2",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-2")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch-2")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))
			})
			It("should NOT update the LastRelevantCommit annotation of the Terraform layers with NO changes", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-last-sync-too-old-layer-3",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch-3")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, "LAST_RELEVANT_REVISION"))
			})
			It("should have put multiple bundles in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-1", mock.GetMockRevision("branch-1"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")

				check, err = reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-2", mock.GetMockRevision("branch-2"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")

				check, err = reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch-3", mock.GetMockRevision("branch-3"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a new TerraformLayer is created for a existent TerraformRepository", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-with-new-layer",
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
			It("should have the condition LastSyncTooOld to True with NewLayer reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("NewLayer"))
			})
			It("should have the condition HasLastSyncFailed to False with NoSyncYet reason", func() {
				Expect(repo.Status.Conditions[1].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[1].Reason).To(Equal("NoSyncYet"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(2))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch"),
					LastSyncDate:   testTime,
				}))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "new-branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("new-branch"),
					LastSyncDate:   testTime,
				}))
			})
			It("should update the annotations of the newly created TerraformLayer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-with-new-layer-new",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("new-branch")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("new-branch")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommitDate, testTime))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))
			})
			It("should have put the bundle of the new branch in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "new-branch", mock.GetMockRevision("new-branch"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository is already synced recently", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-synced-recently",
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
			It("should end in Synced state", func() {
				Expect(repo.Status.State).To(Equal("Synced"))
			})
			It("should have the condition LastSyncTooOld to False with SyncRecent reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("SyncRecent"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch"),
					LastSyncDate:   testTime,
				}))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository has not been synced in the last 24h but is already on last revision", Ordered, func() {
			BeforeAll(func() {
				// Put a fake git bundle
				_ = reconciler.Datastore.PutGitBundle("default", "repo-already-last-revision", "branch", mock.GetMockRevision("branch"), []byte("fake"))
				name = types.NamespacedName{
					Name:      "repo-already-last-revision",
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
			It("should have the condition LastSyncTooOld to True with SyncTooOld reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("SyncTooOld"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch"),
					LastSyncDate:   testTime,
				}))
			})
			It("should update the annotations of the TerraformLayer", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-already-last-revision-layer",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, mock.GetMockRevision("branch")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommitDate, testTime))
			})
			It("should NOT have changed the bundle of the branch in the datastore", func() {
				bundle, err := reconciler.Datastore.GetGitBundle(repo.Namespace, repo.Name, "branch", mock.GetMockRevision("branch"))
				Expect(err).NotTo(HaveOccurred())
				Expect(bundle).To(Equal([]byte("fake")))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository has a recent Sync Now annotation for a branch", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-sync-now",
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
			It("should have the condition LastSyncTooOld to True and SyncNowRequested reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("SyncNowRequested"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch"),
					LastSyncDate:   testTime,
				}))
			})
			It("should update the annotations of the Terraform layers WITH changes", func() {
				layer := &configv1alpha1.TerraformLayer{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      "repo-sync-now-layer",
					Namespace: "default",
				}, layer)).To(Succeed())
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastBranchCommit, mock.GetMockRevision("branch")))
				Expect(layer.Annotations).To(HaveKeyWithValue(annotations.LastRelevantCommit, "LAST_RELEVANT_REVISION"))
			})
			It("should have put multiple bundles in the datastore", func() {
				check, err := reconciler.Datastore.CheckGitBundle(repo.Namespace, repo.Name, "branch", mock.GetMockRevision("branch"))
				Expect(err).NotTo(HaveOccurred())
				Expect(check).To(BeTrue(), "the bundle should be in the datastore")
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
		Describe("When a TerraformRepository has an OLD Sync Now annotation for a branch", Ordered, func() {
			BeforeAll(func() {
				name = types.NamespacedName{
					Name:      "repo-sync-now-old",
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
			It("should end in Synced state", func() {
				Expect(repo.Status.State).To(Equal("Synced"))
			})
			It("should have the condition LastSyncTooOld to False and SyncRecent reason", func() {
				Expect(repo.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
				Expect(repo.Status.Conditions[0].Reason).To(Equal("SyncRecent"))
			})
			It("should update the status of the TerraformRepository", func() {
				Expect(repo.Status.Branches).To(HaveLen(1))
				Expect(repo.Status.Branches).To(ContainElement(configv1alpha1.BranchState{
					Name:           "branch",
					LastSyncStatus: "success",
					LatestRev:      mock.GetMockRevision("branch"),
					LastSyncDate:   testTime,
				}))
			})
			It("should set RequeueAfter to WaitAction", func() {
				Expect(result.RequeueAfter).To(Equal(reconciler.Config.Controller.Timers.WaitAction))
			})
		})
	})
	// Describe("Error Case", func() {
	// })
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
