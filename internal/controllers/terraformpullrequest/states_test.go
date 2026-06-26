package terraformpullrequest

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/repository/credentials"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"k8s.io/client-go/tools/record"
)

type fakeAPIProvider struct {
	changes         []string
	changesErr      error
	commentErr      error
	pullRequests    []configv1alpha1.TerraformPullRequest
	pullRequestsErr error
	mergeCommit      string
	mergeCommitErr   error
	mergeCommitCalls int
	setStatusCalls   []status.CommitStatus
}

func (p *fakeAPIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error) {
	if p.changesErr != nil {
		return nil, p.changesErr
	}
	return p.changes, nil
}

func (p *fakeAPIProvider) Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, c comment.Comment) error {
	return p.commentErr
}

func (p *fakeAPIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	return p.pullRequests, p.pullRequestsErr
}

func (p *fakeAPIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	p.setStatusCalls = append(p.setStatusCalls, s)
	return nil
}

func (p *fakeAPIProvider) GetMergeCommit(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) (string, error) {
	p.mergeCommitCalls++
	if p.mergeCommitErr != nil {
		return "", p.mergeCommitErr
	}
	if p.mergeCommit != "" {
		return p.mergeCommit, nil
	}
	return "fake-merge-commit", nil
}

func TestDiscoveryNeededHandlerReturnsOnErrorWhenLayerCreationFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	linkedLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "layer",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:   "terraform/",
			Branch: "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: "default",
			},
		},
	}
	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repository-secret",
			Namespace: "default",
		},
		Type: credentials.CredentialsType,
		Data: map[string][]byte{
			"provider": []byte("mock"),
			"url":      []byte(repository.Spec.Repository.Url),
		},
	}
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr, linkedLayer, repositorySecret).
		WithIndex(&corev1.Secret{}, "type", func(obj kclient.Object) []string {
			secret := obj.(*corev1.Secret)
			return []string{string(secret.Type)}
		}).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, c kclient.WithWatch, obj kclient.Object, opts ...kclient.CreateOption) error {
				if _, ok := obj.(*configv1alpha1.TerraformLayer); ok {
					return errors.New("create failed")
				}
				return c.Create(ctx, obj, opts...)
			},
		}).
		Build()

	reconciler := &Reconciler{
		Client:      cl,
		Config:      config.TestConfig(),
		Credentials: credentials.NewCredentialStore(cl, time.Hour),
		Recorder:    record.NewFakeRecorder(10),
	}
	state := &State{}

	result := discoveryNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue when layer creation fails, got %s", result.RequeueAfter)
	}
}

func TestDiscoveryNeededHandlerProceedsWhenDeleteTempLayersFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repository-secret",
			Namespace: "default",
		},
		Type: credentials.CredentialsType,
		Data: map[string][]byte{
			"provider": []byte("mock"),
			"url":      []byte(repository.Spec.Repository.Url),
		},
	}
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr, repositorySecret).
		WithIndex(&corev1.Secret{}, "type", func(obj kclient.Object) []string {
			secret := obj.(*corev1.Secret)
			return []string{string(secret.Type)}
		}).
		WithInterceptorFuncs(interceptor.Funcs{
			DeleteAllOf: func(ctx context.Context, c kclient.WithWatch, obj kclient.Object, opts ...kclient.DeleteAllOfOption) error {
				return errors.New("delete failed")
			},
		}).
		Build()

	reconciler := &Reconciler{
		Client:      cl,
		Config:      config.TestConfig(),
		Credentials: credentials.NewCredentialStore(cl, time.Hour),
		Recorder:    record.NewFakeRecorder(10),
	}
	state := &State{}

	result := discoveryNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.WaitAction {
		t.Fatalf("expected reconciliation to proceed despite temp layer deletion failure, got requeue %s", result.RequeueAfter)
	}
}

func TestCommentNeededHandlerReturnsOnErrorWhenGetLinkedLayersFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		List: func(ctx context.Context, c kclient.WithWatch, list kclient.ObjectList, opts ...kclient.ListOption) error {
			return errors.New("list failed")
		},
	}).Build()
	reconciler := &Reconciler{Client: cl, Config: config.TestConfig(), Recorder: record.NewFakeRecorder(10)}
	state := &State{}

	result := commentNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue when GetLinkedLayers fails, got %s", result.RequeueAfter)
	}
}

func TestCommentNeededHandlerReturnsOnErrorWhenCommentFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(repository, pr).Build()
	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return &fakeAPIProvider{commentErr: errors.New("comment failed")}, nil
		},
	}
	state := &State{}

	result := commentNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue when Comment fails, got %s", result.RequeueAfter)
	}
}

func TestResolveMergeCommitCachesAnnotationAfterFirstLookup(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).Build()
	reconciler := &Reconciler{Client: cl}
	provider := &fakeAPIProvider{mergeCommit: "merge-sha-abc"}

	first := reconciler.resolveMergeCommit(context.Background(), provider, terraformRepository("default", "repo"), pr)
	if first != "merge-sha-abc" {
		t.Fatalf("expected resolved merge commit %q, got %q", "merge-sha-abc", first)
	}
	if provider.mergeCommitCalls != 1 {
		t.Fatalf("expected provider to be queried once, got %d calls", provider.mergeCommitCalls)
	}

	updated := &configv1alpha1.TerraformPullRequest{}
	if err := cl.Get(context.Background(), kclient.ObjectKeyFromObject(pr), updated); err != nil {
		t.Fatalf("failed to get pull request: %v", err)
	}
	if updated.Annotations[annotations.MergeCommit] != "merge-sha-abc" {
		t.Fatalf("expected merge commit annotation to be persisted, got %q", updated.Annotations[annotations.MergeCommit])
	}

	second := reconciler.resolveMergeCommit(context.Background(), provider, terraformRepository("default", "repo"), pr)
	if second != "merge-sha-abc" {
		t.Fatalf("expected cached merge commit %q, got %q", "merge-sha-abc", second)
	}
	if provider.mergeCommitCalls != 1 {
		t.Fatalf("expected provider not to be queried again once cached, got %d calls", provider.mergeCommitCalls)
	}
}

func TestResolveMergeCommitReturnsEmptyWhenProviderErrors(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).Build()
	reconciler := &Reconciler{Client: cl}
	provider := &fakeAPIProvider{mergeCommitErr: errors.New("lookup failed")}

	commit := reconciler.resolveMergeCommit(context.Background(), provider, terraformRepository("default", "repo"), pr)
	if commit != "" {
		t.Fatalf("expected empty commit when provider errors, got %q", commit)
	}
}

func TestWaitingForApplyHandlerSetsMergeCommitOnStatus(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).Build()
	provider := &fakeAPIProvider{mergeCommit: "merge-sha-abc"}
	reconciler := &Reconciler{
		Client: cl,
		Config: config.TestConfig(),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}

	result := waitingForApplyHandler(context.Background(), reconciler, repository, pr, &State{})
	if result.RequeueAfter != reconciler.Config.Controller.Timers.WaitAction {
		t.Fatalf("expected wait requeue, got %s", result.RequeueAfter)
	}
	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one SetStatus call, got %d", len(provider.setStatusCalls))
	}
	if provider.setStatusCalls[0].Commit != "merge-sha-abc" {
		t.Fatalf("expected status to target the merge commit, got %q", provider.setStatusCalls[0].Commit)
	}
}

func TestMakeApplyCommentNeededHandlerSetsMergeCommitOnStatus(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).Build()
	provider := &fakeAPIProvider{mergeCommit: "merge-sha-abc"}
	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}

	handler := makeApplyCommentNeededHandler(nil)
	result := handler(context.Background(), reconciler, repository, pr, &State{})
	if !result.IsZero() {
		t.Fatalf("expected empty result after deleting the pull request, got %+v", result)
	}
	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one SetStatus call, got %d", len(provider.setStatusCalls))
	}
	if provider.setStatusCalls[0].Commit != "merge-sha-abc" {
		t.Fatalf("expected status to target the merge commit, got %q", provider.setStatusCalls[0].Commit)
	}
	if provider.setStatusCalls[0].State != status.StateSuccess {
		t.Fatalf("expected success state with no failed layers, got %q", provider.setStatusCalls[0].State)
	}
}

func TestGetStateSelectsExpectedState(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	planningPR := terraformPullRequest("default", "repo-planning", "1", "feature", "sha")
	planningPR.Status.LastDiscoveredCommit = "sha"
	planningLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "planning-layer",
			Namespace: "default",
			Labels: map[string]string{
				managedByLabel: managedByLabelValue(planningPR),
			},
			Annotations: map[string]string{},
		},
	}
	reconciler := &Reconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(planningLayer).Build(),
	}

	tests := []struct {
		name string
		pr   *configv1alpha1.TerraformPullRequest
		want string
	}{
		{
			name: "discovery needed",
			pr:   terraformPullRequest("default", "repo-discovery", "2", "feature", "sha"),
			want: DiscoveryNeeded,
		},
		{
			name: "planning",
			pr:   planningPR,
			want: Planning,
		},
		{
			name: "comment needed",
			pr: func() *configv1alpha1.TerraformPullRequest {
				pr := terraformPullRequest("default", "repo-comment", "3", "feature", "sha")
				pr.Status.LastDiscoveredCommit = "sha"
				return pr
			}(),
			want: CommentNeeded,
		},
		{
			name: "idle",
			pr: func() *configv1alpha1.TerraformPullRequest {
				pr := terraformPullRequest("default", "repo-idle", "4", "feature", "sha")
				pr.Status.LastDiscoveredCommit = "sha"
				pr.Status.LastCommentedCommit = "sha"
				return pr
			}(),
			want: Idle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := reconciler.GetState(context.Background(), tt.pr)
			if state.Status.State != tt.want {
				t.Fatalf("expected state %q, got %q", tt.want, state.Status.State)
			}
		})
	}
}

func TestPlanningHandlerRequeuesWithWaitAction(t *testing.T) {
	reconciler := &Reconciler{Config: config.TestConfig()}

	result := planningHandler(context.Background(), reconciler, nil, nil, nil)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.WaitAction {
		t.Fatalf("expected RequeueAfter %s, got %s", reconciler.Config.Controller.Timers.WaitAction, result.RequeueAfter)
	}
}

func TestCommentNeededHandlerWaitsWhenProviderIsUnavailable(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	cfg := config.TestConfig()
	reconciler := &Reconciler{
		Client:      client,
		Config:      cfg,
		Credentials: credentials.NewCredentialStore(client, cfg.Controller.Timers.CredentialsTTL),
		Recorder:    record.NewFakeRecorder(10),
	}
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	state := &State{}

	result := commentNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != cfg.Controller.Timers.WaitAction {
		t.Fatalf("expected wait requeue when provider is unavailable, got %s", result.RequeueAfter)
	}
	if state.Status.LastCommentedCommit != "" {
		t.Fatalf("expected comment status to stay unchanged")
	}
}

func TestDiscoveryNeededHandlerCreatesTempLayers(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	linkedLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "layer",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:   "terraform/",
			Branch: "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: "default",
			},
		},
	}
	legacyTempLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "legacy-temp",
			Namespace: "default",
			Labels: map[string]string{
				managedByLabel: pr.Name,
			},
			Annotations: map[string]string{},
		},
	}
	currentTempLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "current-temp",
			Namespace: "default",
			Labels: map[string]string{
				managedByLabel: managedByLabelValue(pr),
			},
			Annotations: map[string]string{},
		},
	}
	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repository-secret",
			Namespace: "default",
		},
		Type: credentials.CredentialsType,
		Data: map[string][]byte{
			"provider": []byte("mock"),
			"url":      []byte(repository.Spec.Repository.Url),
		},
	}
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr, linkedLayer, legacyTempLayer, currentTempLayer, repositorySecret).
		WithIndex(&corev1.Secret{}, "type", func(obj kclient.Object) []string {
			secret := obj.(*corev1.Secret)
			return []string{string(secret.Type)}
		}).
		Build()

	reconciler := &Reconciler{
		Client:      cl,
		Config:      config.TestConfig(),
		Credentials: credentials.NewCredentialStore(cl, time.Hour),
		Recorder:    record.NewFakeRecorder(10),
	}
	state := &State{}

	result := discoveryNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.WaitAction {
		t.Fatalf("expected wait requeue, got %s", result.RequeueAfter)
	}
	if state.Status.LastDiscoveredCommit != "sha" {
		t.Fatalf("expected last discovered commit to be set, got %q", state.Status.LastDiscoveredCommit)
	}

	layers := &configv1alpha1.TerraformLayerList{}
	if err := cl.List(context.Background(), layers); err != nil {
		t.Fatalf("failed to list layers: %v", err)
	}
	if len(layers.Items) != 2 {
		t.Fatalf("expected linked layer plus one generated temp layer, got %d", len(layers.Items))
	}
	for _, layer := range layers.Items {
		if layer.Labels[managedByLabel] == pr.Name {
			t.Fatalf("expected legacy temp layers to be removed")
		}
	}
	var generatedCount int
	for _, layer := range layers.Items {
		if layer.Labels[managedByLabel] == managedByLabelValue(pr) {
			generatedCount++
		}
	}
	if generatedCount != 1 {
		t.Fatalf("expected one generated temp layer to keep the managed label, got %d", generatedCount)
	}
}

func TestCommentNeededHandlerPostsCommentAndUpdatesStatus(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	pr.Status.LastDiscoveredCommit = "sha"
	linkedLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "layer",
			Namespace: "default",
			Labels: map[string]string{
				managedByLabel: managedByLabelValue(pr),
			},
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:   "terraform/",
			Branch: "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: "default",
			},
		},
	}
	repositorySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repository-secret",
			Namespace: "default",
		},
		Type: credentials.CredentialsType,
		Data: map[string][]byte{
			"provider": []byte("mock"),
			"url":      []byte(repository.Spec.Repository.Url),
		},
	}
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr, linkedLayer, repositorySecret).
		WithIndex(&corev1.Secret{}, "type", func(obj kclient.Object) []string {
			secret := obj.(*corev1.Secret)
			return []string{string(secret.Type)}
		}).
		Build()

	reconciler := &Reconciler{
		Client:      cl,
		Config:      config.TestConfig(),
		Credentials: credentials.NewCredentialStore(cl, time.Hour),
		Datastore:   datastore.NewMockClient(),
		Recorder:    record.NewFakeRecorder(10),
	}
	state := &State{}

	result := commentNeededHandler(context.Background(), reconciler, repository, pr, state)
	if result.RequeueAfter != reconciler.Config.Controller.Timers.WaitAction {
		t.Fatalf("expected wait requeue, got %s", result.RequeueAfter)
	}
	if state.Status.LastCommentedCommit != "sha" {
		t.Fatalf("expected last commented commit to be set, got %q", state.Status.LastCommentedCommit)
	}
}
