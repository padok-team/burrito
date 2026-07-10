package terraformrun

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeAPIProvider struct {
	setStatusCalls []status.CommitStatus
	setStatusErr   error
	getProviderErr error
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

func (p *fakeAPIProvider) GetMergeCommit(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) (string, error) {
	return "", nil
}

func testRepository() *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
	}
}

func testMainLayer() *configv1alpha1.TerraformLayer {
	return &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"},
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
	}
}

func testRun(action string, revision string) *configv1alpha1.TerraformRun {
	return &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{Name: "pwet-" + action + "-abcde", Namespace: "default"},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: action,
			Layer:  configv1alpha1.TerraformRunLayer{Revision: revision},
		},
	}
}

func TestIsPullRequestLayer(t *testing.T) {
	if isPullRequestLayer(testMainLayer()) {
		t.Fatalf("expected a plain main-branch layer to not be considered a pull request layer")
	}
	if !isPullRequestLayer(testPullRequestLayer()) {
		t.Fatalf("expected a layer owned by a TerraformPullRequest to be considered a pull request layer")
	}
}

func TestSetCommitStatusForDirectPushSkipsPullRequestLayers(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.setCommitStatusForDirectPush(context.Background(), testRun("plan", "sha123"), testPullRequestLayer(), testRepository(), status.StateSuccess)
	if len(provider.setStatusCalls) != 0 {
		t.Fatalf("expected no commit status to be set for a pull request layer, got %d calls", len(provider.setStatusCalls))
	}
}

func TestSetCommitStatusForDirectPushSkipsWhenRevisionIsEmpty(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.setCommitStatusForDirectPush(context.Background(), testRun("plan", ""), testMainLayer(), testRepository(), status.StateSuccess)
	if len(provider.setStatusCalls) != 0 {
		t.Fatalf("expected no commit status to be set when the run has no revision, got %d calls", len(provider.setStatusCalls))
	}
}

func TestSetCommitStatusForDirectPushPostsPlanStatus(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	layer := testMainLayer()
	r.setCommitStatusForDirectPush(context.Background(), testRun("plan", "sha123"), layer, testRepository(), status.StateSuccess)

	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one commit status to be set, got %d", len(provider.setStatusCalls))
	}
	got := provider.setStatusCalls[0]
	if got.Phase != status.PhasePlan {
		t.Errorf("expected phase %q, got %q", status.PhasePlan, got.Phase)
	}
	if got.State != status.StateSuccess {
		t.Errorf("expected state %q, got %q", status.StateSuccess, got.State)
	}
	if got.Commit != "sha123" {
		t.Errorf("expected commit %q, got %q", "sha123", got.Commit)
	}
	wantContext := "burrito/plan/pwet"
	if got.Context != wantContext {
		t.Errorf("expected context %q, got %q", wantContext, got.Context)
	}
}

func TestSetCommitStatusForDirectPushPostsApplyFailureStatus(t *testing.T) {
	provider := &fakeAPIProvider{}
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
	r.setCommitStatusForDirectPush(context.Background(), testRun("apply", "sha456"), testMainLayer(), testRepository(), status.StateFailure)

	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one commit status to be set, got %d", len(provider.setStatusCalls))
	}
	got := provider.setStatusCalls[0]
	if got.Phase != status.PhaseApply {
		t.Errorf("expected phase %q, got %q", status.PhaseApply, got.Phase)
	}
	if got.State != status.StateFailure {
		t.Errorf("expected state %q, got %q", status.StateFailure, got.State)
	}
	if got.Description != "Burrito apply failed" {
		t.Errorf("unexpected description: %q", got.Description)
	}
}

func TestSetCommitStatusForDirectPushDoesNotPanicOnProviderError(t *testing.T) {
	r := &Reconciler{
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return nil, errors.New("no provider configured")
		},
	}
	r.setCommitStatusForDirectPush(context.Background(), testRun("plan", "sha123"), testMainLayer(), testRepository(), status.StateSuccess)
}
