package terraformpullrequest

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/repository/credentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/client-go/tools/record"
)

func TestReconcileIgnoresMissingResources(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: ktypes.NamespacedName{Name: "missing", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("expected no requeue for missing resources, got %s", result.RequeueAfter)
	}
}

func TestReconcilePullRequestReturnsErrorWhenStatusUpdateFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	pr.Status.LastDiscoveredCommit = "sha"
	pr.Status.LastCommentedCommit = "sha"
	expectedErr := errors.New("status update failed")
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr).
		WithStatusSubresource(pr).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				if subResourceName == "status" {
					return expectedErr
				}
				return nil
			},
		}).
		Build()
	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
	}

	result, err := reconciler.reconcilePullRequest(context.Background(), pr)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected status update error, got %v", err)
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
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
			name:   "missing branch commit",
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
			name:   "layer missing plan commit",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), nil, nil),
			},
			want:   true,
			reason: "LayersStillPlanning",
		},
		{
			name:   "layer missing relevant commit",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), strPtr("plan-sha"), nil),
			},
			want:   true,
			reason: "NoRelevantCommitOnLayer",
		},
		{
			name:   "layer finished planning",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
			objects: []client.Object{
				planningLayer("default", "layer-1", managedByLabelValue(planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha")), strPtr("plan-sha"), strPtr("plan-sha")),
			},
			want:   false,
			reason: "LayersNotPlanning",
		},
		{
			name:   "layer still planning",
			pr:     planningPullRequestWithDiscoveredCommit("default", "repo-1", "1", "feature", "sha", "sha"),
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

func TestDiscoveryNeededHandlerCreatesTempLayers(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	linkedLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "layer",
			Namespace: "default",
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
		Recorder:     record.NewFakeRecorder(10),
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
