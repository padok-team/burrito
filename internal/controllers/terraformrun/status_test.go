package terraformrun

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/commitstatus"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
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
		ObjectMeta: metav1.ObjectMeta{Name: "pwet-" + action + "-abcde", Namespace: "default"},
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

	got := r.resultMessage(testRun("plan", "sha123"), layer, commitstatus.Needed)
	if got != layer.Status.LastResult {
		t.Errorf("expected pending outcome to reuse Last Result %q, got %q", layer.Status.LastResult, got)
	}
}

func TestResultMessageReturnsErrorPlaceholderOnDatastoreFailure(t *testing.T) {
	r := &Reconciler{Datastore: &erroringDatastoreClient{}}
	got := r.resultMessage(testRun("apply", "sha123"), testMainLayer(), commitstatus.Succeeded)
	if got != "Error getting last Result" {
		t.Errorf("expected error placeholder, got %q", got)
	}
}

type erroringDatastoreClient struct {
	datastore.MockClient
}

func (c *erroringDatastoreClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	return nil, errors.New("datastore unavailable")
}
