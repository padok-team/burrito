package terraformpullrequest

import (
	"context"
	"errors"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestGenerateTempLayersUsesBoundedNamesAndLabels(t *testing.T) {
	layer := configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
	}
	pr := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("b", 250),
			Namespace: "default",
			UID:       "7c79cdab-2c41-4af0-8fa4-f9cb73ac800c",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "commit",
			},
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			ID: strings.Repeat("9", 250),
		},
	}

	layers := generateTempLayers(pr, []configv1alpha1.TerraformLayer{layer})
	if len(layers) != 1 {
		t.Fatalf("expected one generated layer, got %d", len(layers))
	}

	generated := layers[0]
	if got := len(generated.GenerateName); got > maxGenerateNamePrefixLength {
		t.Fatalf("expected GenerateName length <= %d, got %d", maxGenerateNamePrefixLength, got)
	}
	if got := len(generated.Labels[managedByLabel]); got > 63 {
		t.Fatalf("expected managed label length <= 63, got %d", got)
	}
	if generated.Labels[managedByLabel] == pr.Name {
		t.Fatal("expected managed label to use a bounded hash instead of the raw pull request name")
	}
}

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

func TestGetLinkedLayersForLabelsPropagatesInvalidLabelError(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()

	if _, err := getLinkedLayersForLabels(cl, terraformPullRequest("default", "repo-1", "1", "feature", "sha"), strings.Repeat("a", 64)); err == nil {
		t.Fatalf("expected invalid label value error to be propagated")
	}
}

func TestGetLinkedLayersForLabelReturnsErrorWhenListFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			return errors.New("list failed")
		},
	}).Build()

	if _, err := getLinkedLayersForLabel(cl, "some-label"); err == nil {
		t.Fatalf("expected list error to be propagated")
	}
}

func TestGetAffectedLayersReturnsErrorWhenListFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			return errors.New("list failed")
		},
	}).Build()
	reconciler := &Reconciler{Client: cl}

	if _, err := reconciler.getAffectedLayers(terraformRepository("default", "repo"), terraformPullRequest("default", "repo-1", "1", "feature", "sha")); err == nil {
		t.Fatalf("expected list error to be propagated")
	}
}

func TestGetAffectedLayersReturnsErrorWhenGetChangesFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	expectedErr := errors.New("get changes failed")
	reconciler := &Reconciler{
		Client: cl,
		APIProviderFactory: func(repository *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return &fakeAPIProvider{changesErr: expectedErr}, nil
		},
	}

	if _, err := reconciler.getAffectedLayers(terraformRepository("default", "repo"), terraformPullRequest("default", "repo-1", "1", "feature", "sha")); !errors.Is(err, expectedErr) {
		t.Fatalf("expected GetChanges error to be propagated, got %v", err)
	}
}

func TestDeleteTempLayersPropagatesFirstDeleteFailure(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	expectedErr := errors.New("delete failed")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		DeleteAllOf: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
			return expectedErr
		},
	}).Build()
	reconciler := &Reconciler{Client: cl}

	if err := reconciler.deleteTempLayers(context.Background(), terraformPullRequest("default", "repo-1", "1", "feature", "sha")); !errors.Is(err, expectedErr) {
		t.Fatalf("expected delete error to be propagated, got %v", err)
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
