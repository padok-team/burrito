package terraformrun

import (
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newLabelsTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core scheme: %v", err)
	}
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add burrito scheme: %v", err)
	}
	return scheme
}

func TestGetDefaultLabelsBoundsManagedByValue(t *testing.T) {
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: "plan",
		},
	}

	labels := getDefaultLabels(run)

	value := labels["burrito/managed-by"]
	if got := len(value); got > 63 {
		t.Fatalf("expected managed-by label length <= 63, got %d", got)
	}
	if value == run.Name {
		t.Fatal("expected managed-by label to use a bounded hash instead of the raw run name")
	}
}

func TestGetDefaultLabelsIsDeterministic(t *testing.T) {
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: "plan",
		},
	}

	if getDefaultLabels(run)["burrito/managed-by"] != getDefaultLabels(run)["burrito/managed-by"] {
		t.Fatal("expected managed-by label to be deterministic for the same run")
	}
}

func TestManagedByLabelSelectorValuesIncludesLegacyRawName(t *testing.T) {
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "short-run-name",
			Namespace: "default",
		},
	}

	values := managedByLabelSelectorValues(run)

	if len(values) != 2 {
		t.Fatalf("expected 2 selector values, got %d: %v", len(values), values)
	}
	if !contains(values, run.Name) {
		t.Fatalf("expected selector values to include the legacy raw run name, got %v", values)
	}
}

func TestManagedByLabelSelectorValuesExcludesInvalidLegacyName(t *testing.T) {
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Repeat("a", 250),
			Namespace: "default",
		},
	}

	values := managedByLabelSelectorValues(run)

	if len(values) != 1 {
		t.Fatalf("expected only the hash-based value for a name that's never been a valid label, got %v", values)
	}
}

func TestGetLinkedPodsFindsLegacyLabeledPods(t *testing.T) {
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "short-run-name",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: "plan",
		},
	}
	legacyPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "legacy-pod",
			Namespace: "default",
			Labels: map[string]string{
				"burrito/component": "runner",
				"burrito/action":    "plan",
				managedByLabel:      run.Name,
			},
		},
	}
	currentPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "current-pod",
			Namespace: "default",
			Labels:    getDefaultLabels(run),
		},
	}
	otherRunPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-pod",
			Namespace: "default",
			Labels: map[string]string{
				"burrito/component": "runner",
				"burrito/action":    "plan",
				managedByLabel:      "some-other-run",
			},
		},
	}
	scheme := newLabelsTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(legacyPod, currentPod, otherRunPod).Build()
	reconciler := &Reconciler{Client: cl}

	pods, err := reconciler.GetLinkedPods(run)
	if err != nil {
		t.Fatalf("GetLinkedPods returned error: %v", err)
	}
	if len(pods.Items) != 2 {
		t.Fatalf("expected to find both the legacy and current pod, got %d", len(pods.Items))
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
