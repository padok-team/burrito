package terraformlayer

import (
	"context"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newRunTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add burrito scheme: %v", err)
	}
	return scheme
}

func TestGetDefaultLabelsBoundsManagedByValue(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
	}

	labels := GetDefaultLabels(layer)

	value := labels["burrito/managed-by"]
	if got := len(value); got > 63 {
		t.Fatalf("expected managed-by label length <= 63, got %d", got)
	}
	if value == layer.Name {
		t.Fatal("expected managed-by label to use a bounded hash instead of the raw layer name")
	}
}

func TestGetDefaultLabelsIsDeterministic(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
	}

	if GetDefaultLabels(layer)["burrito/managed-by"] != GetDefaultLabels(layer)["burrito/managed-by"] {
		t.Fatal("expected managed-by label to be deterministic for the same layer")
	}
}

func TestManagedByLabelSelectorValuesIncludesLegacyRawName(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "short-layer-name",
			Namespace: "default",
		},
	}

	values := managedByLabelSelectorValues(layer)

	if len(values) != 2 {
		t.Fatalf("expected 2 selector values, got %d: %v", len(values), values)
	}
	if values[0] != managedByLabelValue(layer) {
		t.Fatalf("expected first value to be the hash-based value, got %q", values[0])
	}
	if !contains(values, layer.Name) {
		t.Fatalf("expected selector values to include the legacy raw layer name, got %v", values)
	}
}

func TestManagedByLabelSelectorValuesExcludesInvalidLegacyName(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
	}

	values := managedByLabelSelectorValues(layer)

	if len(values) != 1 {
		t.Fatalf("expected only the hash-based value for a name that's never been a valid label, got %v", values)
	}
}

func TestGetAllRunsFindsLegacyLabeledRuns(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "short-layer-name",
			Namespace: "default",
		},
	}
	legacyRun := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "legacy-run",
			Namespace: "default",
			Labels:    map[string]string{managedByLabel: layer.Name},
		},
	}
	currentRun := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "current-run",
			Namespace: "default",
			Labels:    GetDefaultLabels(layer),
		},
	}
	otherLayerRun := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-run",
			Namespace: "default",
			Labels:    map[string]string{managedByLabel: "some-other-layer"},
		},
	}
	scheme := newRunTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(legacyRun, currentRun, otherLayerRun).Build()
	reconciler := &Reconciler{Client: cl}

	runs, err := reconciler.getAllRuns(context.Background(), layer)
	if err != nil {
		t.Fatalf("getAllRuns returned error: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected to find both the legacy and current run, got %d", len(runs))
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
