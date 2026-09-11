package terraformrun

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/repository/commitstatus"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeAPIProvider struct {
	setStatusCalls []status.CommitStatus
	setStatusErr   error
}

func (p *fakeAPIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error) {
	return nil, nil
}

func (p *fakeAPIProvider) Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, c comment.Comment) error {
	return nil
}

func (p *fakeAPIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	return nil, nil
}

func (p *fakeAPIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	p.setStatusCalls = append(p.setStatusCalls, s)
	return p.setStatusErr
}

func testRepository() *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
	}
}

func testMainLayer() *configv1alpha1.TerraformLayer {
	return &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"},
		Spec:       configv1alpha1.TerraformLayerSpec{Path: "terraform/pwet"},
	}
}

func testPullRequestLayer() *configv1alpha1.TerraformLayer {
	return &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pwet-pr-abcdef",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "TerraformPullRequest", Name: "repo-1"},
			},
		},
		Spec: configv1alpha1.TerraformLayerSpec{Path: "terraform/pwet"},
	}
}

func testRun(action string, revision string) *configv1alpha1.TerraformRun {
	return &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pwet-" + action + "-abcde",
			Namespace:   "default",
			Annotations: map[string]string{annotations.PostCommitStatus: "true"},
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: action,
			Layer:  configv1alpha1.TerraformRunLayer{Revision: revision},
		},
	}
}

func TestPostCommitStatusSkipsWhenRevisionIsEmpty(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.postCommitStatus(context.Background(), testRun("plan", ""), testMainLayer(), testRepository(), status.StateSuccess, "succeeded")
	if len(provider.setStatusCalls) != 0 {
		t.Fatalf("expected no commit status to be set when the run has no revision, got %d calls", len(provider.setStatusCalls))
	}
}

func TestPostCommitStatusSkipsRunsNotMarkedByTheLayerController(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		Config:    &config.Config{},
		Datastore: datastore.NewMockClient(),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	run := testRun("plan", "sha123")
	run.Annotations = nil

	r.postCommitStatus(context.Background(), run, testMainLayer(), testRepository(), status.StateSuccess, commitstatus.Succeeded)

	if len(provider.setStatusCalls) != 0 {
		t.Fatalf("expected no commit status for an unmarked run, got %d calls", len(provider.setStatusCalls))
	}
}

func TestPostCommitStatusPostsForMainLayer(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		Config:    &config.Config{},
		Datastore: datastore.NewMockClient(),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.postCommitStatus(context.Background(), testRun("plan", "sha123"), testMainLayer(), testRepository(), status.StateSuccess, commitstatus.Succeeded)

	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one commit status to be set, got %d", len(provider.setStatusCalls))
	}
	got := provider.setStatusCalls[0]
	if got.Phase != status.PhasePlan {
		t.Errorf("expected phase %q, got %q", status.PhasePlan, got.Phase)
	}
	if got.Commit != "sha123" {
		t.Errorf("expected commit %q, got %q", "sha123", got.Commit)
	}
	wantContext := "Burrito ▶ Plan default/pwet"
	if got.Context != wantContext {
		t.Errorf("expected context %q, got %q", wantContext, got.Context)
	}
}

func TestPostCommitStatusPostsForPullRequestLayer(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		Config:    &config.Config{},
		Datastore: datastore.NewMockClient(),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.postCommitStatus(context.Background(), testRun("apply", "sha456"), testPullRequestLayer(), testRepository(), status.StateFailure, commitstatus.Failed)

	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one commit status to be set for a pull request layer, got %d", len(provider.setStatusCalls))
	}
	got := provider.setStatusCalls[0]
	if got.Phase != status.PhaseApply {
		t.Errorf("expected phase %q, got %q", status.PhaseApply, got.Phase)
	}
	if got.State != status.StateFailure {
		t.Errorf("expected state %q, got %q", status.StateFailure, got.State)
	}
}

func TestPostCommitStatusDoesNotPanicOnProviderError(t *testing.T) {
	r := &Reconciler{
		Config:    &config.Config{},
		Datastore: datastore.NewMockClient(),
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return nil, errors.New("no provider configured")
		},
	}
	r.postCommitStatus(context.Background(), testRun("plan", "sha123"), testMainLayer(), testRepository(), status.StateSuccess, commitstatus.Succeeded)
}

func TestResultMessageUsesLastResultWhilePending(t *testing.T) {
	r := &Reconciler{Datastore: datastore.NewMockClient()}
	layer := testMainLayer()
	layer.Status.LastResult = "Plan: 1 to add, 0 to change, 0 to destroy."

	got := r.resultMessage(context.Background(), testRun("plan", "sha123"), layer, testRepository(), commitstatus.Needed)
	if got != layer.Status.LastResult {
		t.Errorf("expected pending outcome to reuse Last Result %q, got %q", layer.Status.LastResult, got)
	}
}

func TestResultMessageReturnsErrorPlaceholderOnDatastoreFailure(t *testing.T) {
	r := &Reconciler{Datastore: &erroringDatastoreClient{}}
	got := r.resultMessage(context.Background(), testRun("plan", "sha123"), testMainLayer(), testRepository(), commitstatus.Succeeded)
	if got != "Error getting last Result" {
		t.Errorf("expected error placeholder, got %q", got)
	}
}

func TestResultMessageNamesTheFailedPhaseWithoutReadingTheDatastore(t *testing.T) {
	// A failed run writes no result artifact: reading one back could only ever fail.
	r := &Reconciler{Datastore: &erroringDatastoreClient{}}
	cases := map[string]string{
		"plan":  "Plan failed",
		"apply": "Apply failed",
	}
	for action, want := range cases {
		got := r.resultMessage(context.Background(), testRun(action, "sha123"), testMainLayer(), testRepository(), commitstatus.Failed)
		if got != want {
			t.Errorf("%s: expected %q, got %q", action, want, got)
		}
	}
}

func TestResultMessageDescribesTheAppliedPlan(t *testing.T) {
	r := &Reconciler{Datastore: &planDatastoreClient{shortDiff: "Plan: 2 to create, 1 to update, 1 to delete"}}
	run := testRun("apply", "sha123")
	run.Spec.Artifact = configv1alpha1.Artifact{Run: "pwet-plan-abcde", Attempt: "0"}

	got := r.resultMessage(context.Background(), run, testMainLayer(), testRepository(), commitstatus.Succeeded)
	want := "Applied: 2 to create, 1 to update, 1 to delete"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestResultMessageFallsBackWhenTheAppliedPlanIsUnavailable(t *testing.T) {
	r := &Reconciler{Datastore: &erroringDatastoreClient{}}
	run := testRun("apply", "sha123")
	run.Spec.Artifact = configv1alpha1.Artifact{Run: "pwet-plan-abcde", Attempt: "0"}

	got := r.resultMessage(context.Background(), run, testMainLayer(), testRepository(), commitstatus.Succeeded)
	if got != applySucceeded {
		t.Errorf("expected %q, got %q", applySucceeded, got)
	}
}

func TestResultMessageIgnoresTheAppliedPlanWhenItWasNotReused(t *testing.T) {
	r := &Reconciler{Datastore: &planDatastoreClient{shortDiff: "Plan: 2 to create, 1 to update, 1 to delete"}}
	run := testRun("apply", "sha123")
	run.Spec.Artifact = configv1alpha1.Artifact{Run: "pwet-plan-abcde", Attempt: "0"}
	applyWithoutPlanArtifact := true
	repository := testRepository()
	repository.Spec.RemediationStrategy.ApplyWithoutPlanArtifact = &applyWithoutPlanArtifact

	got := r.resultMessage(context.Background(), run, testMainLayer(), repository, commitstatus.Succeeded)
	if got != applySucceeded {
		t.Errorf("expected the stale diff to be ignored and %q returned, got %q", applySucceeded, got)
	}
}

type erroringDatastoreClient struct {
	datastore.MockClient
}

func (c *erroringDatastoreClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	return nil, errors.New("datastore unavailable")
}

type planDatastoreClient struct {
	datastore.MockClient
	shortDiff string
}

func (c *planDatastoreClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	return []byte(c.shortDiff), nil
}
