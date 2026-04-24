package terraformpullrequest

import (
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
