package terraformpullrequest

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newTerraformPullRequestTestScheme(t *testing.T) *runtime.Scheme {
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

func terraformRepository(namespace string, name string) *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://github.com/padok-team/burrito.git",
			},
		},
	}
}

func terraformPullRequest(namespace string, name string, id string, branch string, commit string) *configv1alpha1.TerraformPullRequest {
	return &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				annotations.LastBranchCommit: commit,
			},
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			ID:     id,
			Branch: branch,
			Base:   "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "repo",
				Namespace: namespace,
			},
		},
	}
}
