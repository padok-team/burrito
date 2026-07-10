package terraformpullrequest

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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

func TestAreLayersStillPlanningCoversBranches(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	tests := []struct {
		name    string
		pr      *configv1alpha1.TerraformPullRequest
		objects []client.Object
		want    bool
		reason  string
	}{
		{
			name: "missing branch commit",
			pr: func() *configv1alpha1.TerraformPullRequest {
				pr := planningPullRequest("default", "repo-1", "1", "feature", "sha")
				delete(pr.Annotations, annotations.LastBranchCommit)
				return pr
			}(),
			want:   true,
			reason: "NoBranchCommitOnPR",
		},
		{
			name:   "no discovered commit",
			pr:     planningPullRequest("default", "repo-1", "1", "feature", "sha"),
			want:   true,
			reason: "NoCommitDiscovered",
		},
		{
			name:   "discovery still needed",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "other"),
			want:   true,
			reason: "StillNeedsDiscovery",
		},
		{
			name:   "no linked layers",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			want:   false,
			reason: "LayersNotPlanning",
		},
		{
			name: "layer missing plan commit",
			pr:   planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), nil, nil),
			},
			want:   true,
			reason: "LayersStillPlanning",
		},
		{
			name: "layer missing relevant commit",
			pr:   planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), strPtr("plan-sha"), nil),
			},
			want:   true,
			reason: "NoRelevantCommitOnLayer",
		},
		{
			name: "layer finished planning",
			pr:   planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), strPtr("plan-sha"), strPtr("plan-sha")),
			},
			want:   false,
			reason: "LayersNotPlanning",
		},
		{
			name: "layer still planning",
			pr:   planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), strPtr("plan-sha"), strPtr("other-sha")),
			},
			want:   true,
			reason: "LayersStillPlanning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tt.objects...).Build()
			reconciler := &Reconciler{Client: cl}
			condition, got := reconciler.AreLayersStillPlanning(tt.pr)
			if got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
			if condition.Reason != tt.reason {
				t.Fatalf("expected reason %q, got %q", tt.reason, condition.Reason)
			}
		})
	}
}

func planningPullRequest(namespace, name, id, branch, commit string) *configv1alpha1.TerraformPullRequest {
	pr := terraformPullRequest(namespace, name, id, branch, commit)
	pr.Status.LastDiscoveredCommit = ""
	return pr
}

func planningPullRequestWithDiscoveredCommit(namespace, name, id, branch, commit, discoveredCommit string) *configv1alpha1.TerraformPullRequest {
	pr := terraformPullRequest(namespace, name, id, branch, commit)
	pr.Status.LastDiscoveredCommit = discoveredCommit
	return pr
}

func planningLayer(namespace, name, managedLabel string, lastPlanCommit, lastRelevantCommit *string) *configv1alpha1.TerraformLayer {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				managedByLabel: managedLabel,
			},
			Annotations: map[string]string{},
		},
	}
	if lastPlanCommit != nil {
		layer.Annotations[annotations.LastPlanCommit] = *lastPlanCommit
	}
	if lastRelevantCommit != nil {
		layer.Annotations[annotations.LastRelevantCommit] = *lastRelevantCommit
	}
	return layer
}

func strPtr(value string) *string {
	return &value
}
