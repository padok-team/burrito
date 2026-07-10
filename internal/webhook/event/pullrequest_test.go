package event

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPullRequestEventHandleReturnsNilWhenNoRepositoryMatches(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository := terraformRepository("default", "repo", "https://github.com/acme/repo")
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository).Build()

	event := &PullRequestEvent{
		URL:       "https://github.com/acme/other-repo",
		Reference: "feature/branch",
		Base:      "main",
		Action:    PullRequestOpened,
		ID:        "42",
		Commit:    "new-sha",
	}

	if err := event.Handle(client); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var pullRequestList configv1alpha1.TerraformPullRequestList
	if err := client.List(context.Background(), &pullRequestList); err != nil {
		t.Fatalf("failed to list pull requests: %v", err)
	}
	if len(pullRequestList.Items) != 0 {
		t.Fatalf("expected no pull requests to be created, got %d", len(pullRequestList.Items))
	}
}

func TestPullRequestEventHandleOpenedCreatesAndUpdatesPullRequests(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository1 := terraformRepository("default", "repo-1", "https://github.com/acme/repo")
	repository2 := terraformRepository("default", "repo-2", "https://github.com/acme/repo")
	current := terraformPullRequest("default", "repo-1-42", "feature/old", "stale-sha")
	current.Annotations["example.com/preserved"] = "true"

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository1, &repository2, current).Build()

	event := &PullRequestEvent{
		URL:       "https://github.com/acme/repo",
		Reference: "feature/branch",
		Base:      "main",
		Action:    PullRequestOpened,
		ID:        "42",
		Commit:    "new-sha",
	}

	if err := event.Handle(client); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	updated := &configv1alpha1.TerraformPullRequest{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-1-42", Namespace: "default"}, updated); err != nil {
		t.Fatalf("failed to get updated pull request: %v", err)
	}
	if updated.Spec.Branch != "feature/branch" {
		t.Fatalf("expected branch to be updated, got %q", updated.Spec.Branch)
	}
	if updated.Spec.Base != "main" {
		t.Fatalf("expected base to be preserved, got %q", updated.Spec.Base)
	}
	if updated.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected LastBranchCommit annotation to be updated")
	}
	if updated.Annotations["example.com/preserved"] != "true" {
		t.Fatalf("expected preserved annotation to remain on updated pull request")
	}

	created := &configv1alpha1.TerraformPullRequest{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-2-42", Namespace: "default"}, created); err != nil {
		t.Fatalf("failed to get created pull request: %v", err)
	}
	if created.Spec.Repository.Name != "repo-2" {
		t.Fatalf("expected created pull request to target repo-2, got %q", created.Spec.Repository.Name)
	}
	if created.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected created pull request to carry LastBranchCommit annotation")
	}
}

func TestPullRequestEventHandleClosedDeletesPullRequestsAndRemovesAnnotations(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository1 := terraformRepository("default", "repo-1", "https://github.com/acme/repo")
	repository2 := terraformRepository("default", "repo-2", "https://github.com/acme/repo")
	branchKey := annotations.ComputeKeyForSyncBranchNow("feature/branch")
	repository1.Annotations[branchKey] = "true"
	repository2.Annotations[branchKey] = "true"
	current := terraformPullRequest("default", "repo-1-42", "feature/branch", "stale-sha")

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository1, &repository2, current).Build()

	event := &PullRequestEvent{
		URL:       "https://github.com/acme/repo",
		Reference: "feature/branch",
		Base:      "main",
		Action:    PullRequestClosed,
		ID:        "42",
		Commit:    "new-sha",
	}

	if err := event.Handle(client); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	deleted := &configv1alpha1.TerraformPullRequest{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-1-42", Namespace: "default"}, deleted); !apierrors.IsNotFound(err) {
		t.Fatalf("expected pull request to be deleted, got %v", err)
	}

	missing := &configv1alpha1.TerraformPullRequest{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-2-42", Namespace: "default"}, missing); !apierrors.IsNotFound(err) {
		t.Fatalf("expected missing pull request to stay absent, got %v", err)
	}

	updatedRepository1 := &configv1alpha1.TerraformRepository{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-1", Namespace: "default"}, updatedRepository1); err != nil {
		t.Fatalf("failed to get updated repository1: %v", err)
	}
	if _, ok := updatedRepository1.Annotations[branchKey]; ok {
		t.Fatalf("expected sync annotation to be removed from repository1")
	}

	updatedRepository2 := &configv1alpha1.TerraformRepository{}
	if err := client.Get(context.Background(), ktypes.NamespacedName{Name: "repo-2", Namespace: "default"}, updatedRepository2); err != nil {
		t.Fatalf("failed to get updated repository2: %v", err)
	}
	if _, ok := updatedRepository2.Annotations[branchKey]; ok {
		t.Fatalf("expected sync annotation to be removed from repository2")
	}
}

func TestPullRequestEventHandleIgnoresUnsupportedAction(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository := terraformRepository("default", "repo", "https://github.com/acme/repo")
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository).Build()

	event := &PullRequestEvent{
		URL:       "https://github.com/acme/repo",
		Reference: "feature/branch",
		Base:      "main",
		Action:    "synchronized",
		ID:        "42",
		Commit:    "new-sha",
	}

	if err := event.Handle(client); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var pullRequestList configv1alpha1.TerraformPullRequestList
	if err := client.List(context.Background(), &pullRequestList); err != nil {
		t.Fatalf("failed to list pull requests: %v", err)
	}
	if len(pullRequestList.Items) != 0 {
		t.Fatalf("expected unsupported action to skip pull request changes, got %d pull requests", len(pullRequestList.Items))
	}
}

func TestBatchCreatePullRequestsHandlesCreateRace(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository := terraformRepository("default", "repo", "https://github.com/acme/repo")
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository).Build()
	racingClient := &createRaceClient{
		Client:    baseClient,
		targetKey: ktypes.NamespacedName{Name: "repo-42", Namespace: "default"},
	}

	pr := terraformPullRequest("default", "repo-42", "feature/branch", "new-sha")
	pr.Annotations["example.com/preserved"] = "true"

	if err := batchCreatePullRequests(context.Background(), racingClient, []configv1alpha1.TerraformPullRequest{*pr}); err != nil {
		t.Fatalf("batchCreatePullRequests returned error: %v", err)
	}

	created := &configv1alpha1.TerraformPullRequest{}
	if err := baseClient.Get(context.Background(), ktypes.NamespacedName{Name: "repo-42", Namespace: "default"}, created); err != nil {
		t.Fatalf("failed to get created pull request: %v", err)
	}
	if created.Spec.Branch != "feature/branch" {
		t.Fatalf("expected branch to be set after create race, got %q", created.Spec.Branch)
	}
	if created.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected branch commit annotation to be preserved after create race")
	}
	if created.Annotations["example.com/preserved"] != "true" {
		t.Fatalf("expected custom annotation to survive create race handling")
	}
}

func TestBatchDeletePullRequestsReturnsErrorOnUnexpectedDeleteFailure(t *testing.T) {
	scheme := newPullRequestEventTestScheme(t)
	repository := terraformRepository("default", "repo", "https://github.com/acme/repo")
	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&repository).Build()
	failingClient := &deleteErrorClient{
		Client:    baseClient,
		targetKey: ktypes.NamespacedName{Name: "repo-42", Namespace: "default"},
	}

	err := batchDeletePullRequests(context.Background(), failingClient, []configv1alpha1.TerraformPullRequest{*terraformPullRequest("default", "repo-42", "feature/branch", "new-sha")})
	if err == nil {
		t.Fatalf("expected batchDeletePullRequests to return an error")
	}
	if !errors.Is(err, failingClient.err) {
		t.Fatalf("expected returned error to include delete failure, got %v", err)
	}
}

func TestMergePullRequestEventAnnotationsInitializesAndPreservesAnnotations(t *testing.T) {
	current := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"example.com/preserved": "true",
			},
		},
	}
	desired := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotations.LastBranchCommit: "new-sha",
			},
		},
	}

	mergePullRequestEventAnnotations(current, desired)

	if current.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected desired annotation to be merged")
	}
	if current.Annotations["example.com/preserved"] != "true" {
		t.Fatalf("expected existing annotation to be preserved")
	}

	emptyCurrent := &configv1alpha1.TerraformPullRequest{}
	mergePullRequestEventAnnotations(emptyCurrent, desired)
	if emptyCurrent.Annotations == nil {
		t.Fatalf("expected annotations map to be initialized")
	}
	if emptyCurrent.Annotations[annotations.LastBranchCommit] != "new-sha" {
		t.Fatalf("expected desired annotation to be copied into initialized map")
	}
}

func newPullRequestEventTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}
	return scheme
}

func terraformRepository(namespace string, name string, url string) configv1alpha1.TerraformRepository {
	return configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: url,
			},
		},
	}
}

func terraformPullRequest(namespace string, name string, branch string, commit string) *configv1alpha1.TerraformPullRequest {
	return &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			Annotations: map[string]string{annotations.LastBranchCommit: commit},
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			Branch: branch,
			Base:   "main",
			ID:     "42",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}

type createRaceClient struct {
	client.Client
	targetKey ktypes.NamespacedName
}

func (c *createRaceClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	pullRequest, ok := obj.(*configv1alpha1.TerraformPullRequest)
	key := ktypes.NamespacedName{Name: pullRequest.Name, Namespace: pullRequest.Namespace}
	if !ok || key != c.targetKey {
		return c.Client.Create(ctx, obj, opts...)
	}

	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	return apierrors.NewAlreadyExists(schema.GroupResource{Group: "config.terraform.padok.cloud", Resource: "terraformpullrequests"}, pullRequest.Name)
}

type deleteErrorClient struct {
	client.Client
	targetKey ktypes.NamespacedName
	err       error
}

func (c *deleteErrorClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	pullRequest, ok := obj.(*configv1alpha1.TerraformPullRequest)
	key := ktypes.NamespacedName{Name: pullRequest.Name, Namespace: pullRequest.Namespace}
	if !ok || key != c.targetKey {
		return c.Client.Delete(ctx, obj, opts...)
	}

	if c.err == nil {
		c.err = errors.New("delete failed")
	}
	return c.err
}
