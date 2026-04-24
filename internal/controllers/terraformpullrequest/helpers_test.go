package terraformpullrequest

import (
	"context"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/repository/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestIsLayerAffected(t *testing.T) {
	pr := configv1alpha1.TerraformPullRequest{
		Spec: configv1alpha1.TerraformPullRequestSpec{
			Base: "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: "default",
			},
		},
	}
	layer := configv1alpha1.TerraformLayer{
		Spec: configv1alpha1.TerraformLayerSpec{
			Branch: "main",
			Path:   "terraform/",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: "default",
			},
		},
	}

	tests := []struct {
		name    string
		layer   configv1alpha1.TerraformLayer
		pr      configv1alpha1.TerraformPullRequest
		changes []string
		want    bool
	}{
		{
			name:    "repository name mismatch",
			layer:   withLayerRepository(layer, "other", "default"),
			pr:      pr,
			changes: []string{"terraform/main.tf"},
			want:    false,
		},
		{
			name:    "repository namespace mismatch",
			layer:   withLayerRepository(layer, "repo", "other"),
			pr:      pr,
			changes: []string{"terraform/main.tf"},
			want:    false,
		},
		{
			name:    "base ref not targeted",
			layer:   withLayerBranch(layer, "develop"),
			pr:      pr,
			changes: []string{"terraform/main.tf"},
			want:    false,
		},
		{
			name:    "changed file under layer path",
			layer:   layer,
			pr:      pr,
			changes: []string{"terraform/main.tf"},
			want:    true,
		},
		{
			name:    "base ref in additional targets",
			layer:   withAdditionalTargetRefs(withLayerBranch(layer, "release"), "main"),
			pr:      pr,
			changes: []string{"terraform/main.tf"},
			want:    true,
		},
		{
			name:    "targeted layer without matching changes",
			layer:   layer,
			pr:      pr,
			changes: []string{"docs/readme.md"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLayerAffected(tt.layer, tt.pr, tt.changes); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}

func TestIsLastCommitDiscoveredWhenDiscoveredCommitIsStale(t *testing.T) {
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "new-sha")
	pr.Status.LastDiscoveredCommit = "old-sha"

	condition, discovered := (&Reconciler{}).IsLastCommitDiscovered(pr)
	if discovered {
		t.Fatalf("expected stale discovered commit to require discovery")
	}
	if condition.Reason != "LastCommitNotDiscovered" {
		t.Fatalf("expected LastCommitNotDiscovered reason, got %q", condition.Reason)
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

func TestGetLinkedLayersForLabelsDeduplicatesMatches(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "layer",
			Namespace: "default",
			Labels: map[string]string{
				managedByLabel: "legacy-pr-name",
			},
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(layer).Build()

	layers, err := getLinkedLayersForLabels(client, terraformPullRequest("default", "repo-1", "1", "feature", "sha"), "legacy-pr-name", "legacy-pr-name")
	if err != nil {
		t.Fatalf("getLinkedLayersForLabels returned error: %v", err)
	}
	if len(layers) != 1 {
		t.Fatalf("expected duplicate label matches to be deduplicated, got %d layers", len(layers))
	}
}

func TestGetLinkedLayersForLabelRejectsInvalidLabelValue(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	if _, err := getLinkedLayersForLabel(client, strings.Repeat("a", 64)); err == nil {
		t.Fatalf("expected invalid label value to return an error")
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

func TestIgnorePredicate(t *testing.T) {
	predicate := ignorePredicate()
	oldPR := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
			Annotations: map[string]string{
				"key": "old",
			},
		},
	}
	newPR := oldPR.DeepCopy()

	if predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: newPR}) {
		t.Fatalf("expected unchanged PR update to be ignored")
	}

	annotatedPR := oldPR.DeepCopy()
	annotatedPR.Annotations["key"] = "new"
	if !predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: annotatedPR}) {
		t.Fatalf("expected annotation change to trigger reconciliation")
	}

	generatedPR := oldPR.DeepCopy()
	generatedPR.Generation = 2
	if !predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: generatedPR}) {
		t.Fatalf("expected generation change to trigger reconciliation")
	}

	if predicate.Delete(event.DeleteEvent{Object: oldPR, DeleteStateUnknown: true}) {
		t.Fatalf("expected confirmed unknown delete state to be ignored")
	}
	if !predicate.Delete(event.DeleteEvent{Object: oldPR, DeleteStateUnknown: false}) {
		t.Fatalf("expected known delete state to trigger reconciliation")
	}
}

func withLayerRepository(layer configv1alpha1.TerraformLayer, name string, namespace string) configv1alpha1.TerraformLayer {
	layer.Spec.Repository.Name = name
	layer.Spec.Repository.Namespace = namespace
	return layer
}

func withLayerBranch(layer configv1alpha1.TerraformLayer, branch string) configv1alpha1.TerraformLayer {
	layer.Spec.Branch = branch
	return layer
}

func withAdditionalTargetRefs(layer configv1alpha1.TerraformLayer, refs ...string) configv1alpha1.TerraformLayer {
	layer.Spec.AdditionalTargetRefs = refs
	return layer
}
