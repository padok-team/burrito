package terraformrepository

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type fakeAPIProvider struct {
	pullRequests   []configv1alpha1.TerraformPullRequest
	err            error
	mergeCommit    string
	mergeCommitErr error
}

func (p *fakeAPIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error) {
	return nil, nil
}

func (p *fakeAPIProvider) Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	return nil
}

func (p *fakeAPIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.pullRequests, nil
}

func (p *fakeAPIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	return nil
}

func (p *fakeAPIProvider) GetMergeCommit(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) (string, error) {
	if p.mergeCommitErr != nil {
		return "", p.mergeCommitErr
	}
	return p.mergeCommit, nil
}

func newTerraformRepositoryTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add burrito scheme: %v", err)
	}
	return scheme
}

func testTerraformRepository(namespace string, name string) *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://github.com/padok-team/burrito.git",
			},
		},
	}
}

func testTerraformPullRequest(namespace string, name string, id string, branch string, commit string) *configv1alpha1.TerraformPullRequest {
	return &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				annotations.LastBranchCommit: commit,
			},
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			ID:     id,
			Branch: branch,
			Base:   "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: namespace,
			},
		},
	}
}

func TestApplyDesiredPullRequestCreatesUpdatesAndPreservesAnnotations(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	current := testTerraformPullRequest("default", "repo-1", "1", "old", "old-sha")
	current.Annotations["example.com/preserved"] = "true"
	reconciler := &Reconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build(),
	}

	desired := testTerraformPullRequest("default", "repo-1", "1", "new", "new-sha")
	if err := reconciler.applyDesiredPullRequest(context.Background(), desired); err != nil {
		t.Fatalf("applyDesiredPullRequest returned error: %v", err)
	}

	updated := &configv1alpha1.TerraformPullRequest{}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: "repo-1", Namespace: "default"}, updated); err != nil {
		t.Fatalf("failed to get updated pull request: %v", err)
	}
	if updated.Spec.Branch != "new" {
		t.Fatalf("expected branch to be updated, got %q", updated.Spec.Branch)
	}
	if updated.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected remote commit annotation to be updated")
	}
	if updated.Annotations["example.com/preserved"] != "true" {
		t.Fatalf("expected unrelated annotation to be preserved")
	}

	missing := testTerraformPullRequest("default", "repo-2", "2", "feature", "sha")
	if err := reconciler.applyDesiredPullRequest(context.Background(), missing); err != nil {
		t.Fatalf("applyDesiredPullRequest should create missing PRs, got %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: "repo-2", Namespace: "default"}, &configv1alpha1.TerraformPullRequest{}); err != nil {
		t.Fatalf("expected missing PR to be created, got %v", err)
	}
}

func TestSyncPullRequestsCreatesDesiredAndDeletesClosedPullRequests(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	openRemote := testTerraformPullRequest("default", "repo-1", "1", "feature", "remote-sha")
	stale := testTerraformPullRequest("default", "repo-2", "2", "closed", "old-sha")
	otherRepositoryPR := testTerraformPullRequest("default", "other-3", "3", "feature", "sha")
	otherRepositoryPR.Spec.Repository.Name = "other"

	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, stale, otherRepositoryPR).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{*openRemote}}, nil
		},
	}

	if err := reconciler.syncPullRequests(context.Background(), repository); err != nil {
		t.Fatalf("syncPullRequests returned error: %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: openRemote.Name, Namespace: openRemote.Namespace}, &configv1alpha1.TerraformPullRequest{}); err != nil {
		t.Fatalf("expected remote PR to be created, got %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: stale.Name, Namespace: stale.Namespace}, &configv1alpha1.TerraformPullRequest{}); !apierrors.IsNotFound(err) {
		t.Fatalf("expected stale PR to be deleted, got %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: otherRepositoryPR.Name, Namespace: otherRepositoryPR.Namespace}, &configv1alpha1.TerraformPullRequest{}); err != nil {
		t.Fatalf("expected other repository PR to be preserved, got %v", err)
	}
}

func TestSyncPullRequestsPreservesMergedPullRequestNotInOpenList(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	merged := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	merged.Annotations[annotations.MergedAt] = "2026-01-01T00:00:00Z"

	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, merged).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			// The remote pull/merge request no longer appears in the "open" list once merged.
			return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{}}, nil
		},
	}

	if err := reconciler.syncPullRequests(context.Background(), repository); err != nil {
		t.Fatalf("syncPullRequests returned error: %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: merged.Name, Namespace: merged.Namespace}, &configv1alpha1.TerraformPullRequest{}); err != nil {
		t.Fatalf("expected merged pull request to survive until the apply flow deletes it, got %v", err)
	}
}

func TestSyncPullRequestsDetectsMergeViaPollingWhenNoWebhookAnnotatedIt(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	// Not yet annotated MergedAt: no webhook is configured, so polling is the only signal.
	disappeared := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")

	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, disappeared).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{}, mergeCommit: "merge-sha-abc"}, nil
		},
	}

	if err := reconciler.syncPullRequests(context.Background(), repository); err != nil {
		t.Fatalf("syncPullRequests returned error: %v", err)
	}

	updated := &configv1alpha1.TerraformPullRequest{}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: disappeared.Name, Namespace: disappeared.Namespace}, updated); err != nil {
		t.Fatalf("expected pull request to survive as merged, got %v", err)
	}
	if updated.Annotations[annotations.MergedAt] == "" {
		t.Fatalf("expected MergedAt annotation to be set")
	}
	if updated.Annotations[annotations.MergeCommit] != "merge-sha-abc" {
		t.Fatalf("expected merge commit annotation %q, got %q", "merge-sha-abc", updated.Annotations[annotations.MergeCommit])
	}
}

func TestSyncPullRequestsDeletesGenuinelyClosedPullRequestWithoutMergeCommit(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	closed := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")

	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, closed).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			// No merge commit: the pull request was closed without being merged.
			return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{}}, nil
		},
	}

	if err := reconciler.syncPullRequests(context.Background(), repository); err != nil {
		t.Fatalf("syncPullRequests returned error: %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: closed.Name, Namespace: closed.Namespace}, &configv1alpha1.TerraformPullRequest{}); !apierrors.IsNotFound(err) {
		t.Fatalf("expected closed-without-merge pull request to be deleted, got %v", err)
	}
}

func TestSyncPullRequestsReturnsErrors(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	expectedFactoryErr := errors.New("factory failed")
	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return nil, expectedFactoryErr
		},
	}

	err := reconciler.syncPullRequests(context.Background(), repository)
	if !errors.Is(err, expectedFactoryErr) {
		t.Fatalf("expected provider factory error, got %v", err)
	}

	expectedListErr := errors.New("list failed")
	reconciler.APIProviderFactory = func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
		return &fakeAPIProvider{err: expectedListErr}, nil
	}
	err = reconciler.syncPullRequests(context.Background(), repository)
	if !errors.Is(err, expectedListErr) {
		t.Fatalf("expected list error, got %v", err)
	}
}

func TestSyncPullRequestsSkipsUpToDatePullRequestForDeletion(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	upToDate := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")

	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, upToDate).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{*upToDate}}, nil
		},
	}

	if err := reconciler.syncPullRequests(context.Background(), repository); err != nil {
		t.Fatalf("syncPullRequests returned error: %v", err)
	}
	if err := reconciler.Get(context.Background(), ktypes.NamespacedName{Name: upToDate.Name, Namespace: upToDate.Namespace}, &configv1alpha1.TerraformPullRequest{}); err != nil {
		t.Fatalf("expected up-to-date PR to be preserved, got %v", err)
	}
}

func TestSyncPullRequestsPropagatesDownstreamErrors(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	remotePR := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	stale := testTerraformPullRequest("default", "repo-2", "2", "closed", "old-sha")

	t.Run("existing list failure", func(t *testing.T) {
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository).WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				if _, ok := list.(*configv1alpha1.TerraformPullRequestList); ok {
					return errors.New("list failed")
				}
				return c.List(ctx, list, opts...)
			},
		}).Build()
		reconciler := &Reconciler{
			Client: cl,
			APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
				return &fakeAPIProvider{}, nil
			},
		}
		if err := reconciler.syncPullRequests(context.Background(), repository); err == nil {
			t.Fatalf("expected error when listing existing pull requests fails")
		}
	})

	t.Run("apply desired failure", func(t *testing.T) {
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key ktypes.NamespacedName, obj client.Object, opts ...client.GetOption) error {
				if _, ok := obj.(*configv1alpha1.TerraformPullRequest); ok {
					return errors.New("get failed")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()
		reconciler := &Reconciler{
			Client: cl,
			APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
				return &fakeAPIProvider{pullRequests: []configv1alpha1.TerraformPullRequest{*remotePR}}, nil
			},
		}
		if err := reconciler.syncPullRequests(context.Background(), repository); err == nil {
			t.Fatalf("expected error when applying a desired pull request fails")
		}
	})

	t.Run("delete failure", func(t *testing.T) {
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, stale).WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				return errors.New("delete failed")
			},
		}).Build()
		reconciler := &Reconciler{
			Client: cl,
			APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
				return &fakeAPIProvider{}, nil
			},
		}
		if err := reconciler.syncPullRequests(context.Background(), repository); err == nil {
			t.Fatalf("expected error when deleting a stale pull request fails")
		}
	})
}

func TestApplyDesiredPullRequestReturnsErrorOnInitialGetFailure(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	desired := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key ktypes.NamespacedName, obj client.Object, opts ...client.GetOption) error {
			return errors.New("get failed")
		},
	}).Build()
	reconciler := &Reconciler{Client: cl}

	if err := reconciler.applyDesiredPullRequest(context.Background(), desired); err == nil {
		t.Fatalf("expected error when the initial Get fails with a non-NotFound error")
	}
}

func TestApplyDesiredPullRequestReturnsErrorWhenCreateRaceGetFails(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	desired := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	baseClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	racingClient := &createRaceGetFailureClient{
		Client:    baseClient,
		targetKey: ktypes.NamespacedName{Name: desired.Name, Namespace: desired.Namespace},
	}
	reconciler := &Reconciler{Client: racingClient}

	if err := reconciler.applyDesiredPullRequest(context.Background(), desired); err == nil {
		t.Fatalf("expected error when the post-create-race Get fails")
	}
}

type createRaceGetFailureClient struct {
	client.Client
	targetKey ktypes.NamespacedName
	created   bool
}

func (c *createRaceGetFailureClient) Get(ctx context.Context, key ktypes.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if key == c.targetKey && c.created {
		return errors.New("get after create race failed")
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *createRaceGetFailureClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	pullRequest, ok := obj.(*configv1alpha1.TerraformPullRequest)
	if !ok {
		return c.Client.Create(ctx, obj, opts...)
	}
	key := ktypes.NamespacedName{Name: pullRequest.Name, Namespace: pullRequest.Namespace}
	if key != c.targetKey {
		return c.Client.Create(ctx, obj, opts...)
	}
	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	c.created = true
	return apierrors.NewAlreadyExists(schema.GroupResource{Group: "config.terraform.padok.cloud", Resource: "terraformpullrequests"}, pullRequest.Name)
}

func TestApplyDesiredPullRequestHandlesCreateRace(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")
	desired := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	desired.Annotations["example.com/preserved"] = "true"
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository).Build()
	racingClient := &createRacePullRequestClient{
		Client:    baseClient,
		targetKey: ktypes.NamespacedName{Name: desired.Name, Namespace: desired.Namespace},
	}
	reconciler := &Reconciler{Client: racingClient}

	if err := reconciler.applyDesiredPullRequest(context.Background(), desired); err != nil {
		t.Fatalf("applyDesiredPullRequest returned error: %v", err)
	}

	created := &configv1alpha1.TerraformPullRequest{}
	if err := baseClient.Get(context.Background(), ktypes.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, created); err != nil {
		t.Fatalf("failed to get created pull request: %v", err)
	}
	if created.Spec.Branch != "feature" {
		t.Fatalf("expected branch to be created, got %q", created.Spec.Branch)
	}
	if created.Annotations["example.com/preserved"] != "true" {
		t.Fatalf("expected annotations to survive create race handling")
	}
}

func TestPollingAnnotationHelpers(t *testing.T) {
	current := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"owned":     "old",
				"preserved": "true",
			},
		},
	}
	desired := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"owned": "new",
			},
		},
	}

	if pollingAnnotationsEqual(current, desired) {
		t.Fatalf("expected polling annotations to be different")
	}

	mergePollingAnnotations(current, desired)
	if current.Annotations["owned"] != "new" {
		t.Fatalf("expected owned annotation to be updated")
	}
	if current.Annotations["preserved"] != "true" {
		t.Fatalf("expected unrelated annotation to be preserved")
	}
	if !pollingAnnotationsEqual(current, desired) {
		t.Fatalf("expected polling annotations to match after merge")
	}

	noAnnotations := &configv1alpha1.TerraformPullRequest{}
	mergePollingAnnotations(noAnnotations, desired)
	if noAnnotations.Annotations["owned"] != "new" {
		t.Fatalf("expected merge to initialize annotations")
	}
}

func TestDeleteRemotePullRequestIsIdempotent(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	pr := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr-1",
			Namespace: "default",
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).Build()
	reconciler := &Reconciler{Client: cl}

	if err := reconciler.deleteRemotePullRequest(context.Background(), pr); err != nil {
		t.Fatalf("deleteRemotePullRequest returned error: %v", err)
	}
	if err := cl.Get(context.Background(), ktypes.NamespacedName{Name: pr.Name, Namespace: pr.Namespace}, &configv1alpha1.TerraformPullRequest{}); !apierrors.IsNotFound(err) {
		t.Fatalf("expected pull request to be deleted, got error %v", err)
	}
	if err := reconciler.deleteRemotePullRequest(context.Background(), pr); err != nil {
		t.Fatalf("deleteRemotePullRequest should ignore missing objects, got %v", err)
	}
}

func TestApplyDesiredPullRequestNoopsWhenCurrentMatchesDesired(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	current := testTerraformPullRequest("default", "repo-1", "1", "feature", "sha")
	reconciler := &Reconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(current).Build(),
	}

	desired := current.DeepCopy()
	if err := reconciler.applyDesiredPullRequest(context.Background(), desired); err != nil {
		t.Fatalf("applyDesiredPullRequest returned error: %v", err)
	}
}

type createRacePullRequestClient struct {
	client.Client
	targetKey ktypes.NamespacedName
}

func (c *createRacePullRequestClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	pullRequest, ok := obj.(*configv1alpha1.TerraformPullRequest)
	if !ok {
		return c.Client.Create(ctx, obj, opts...)
	}
	key := ktypes.NamespacedName{Name: pullRequest.Name, Namespace: pullRequest.Namespace}
	if key != c.targetKey {
		return c.Client.Create(ctx, obj, opts...)
	}
	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	return apierrors.NewAlreadyExists(schema.GroupResource{Group: "config.terraform.padok.cloud", Resource: "terraformpullrequests"}, pullRequest.Name)
}
