package mock

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeComment struct{}

func (c *fakeComment) Generate(commit string) (string, error) {
	return "", nil
}

func TestAPIProvider_ListPullRequests(t *testing.T) {
	api := &APIProvider{}

	t.Run("returns no pull requests for an unrelated repository", func(t *testing.T) {
		repository := &configv1alpha1.TerraformRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{Url: "https://github.com/owner/repo"},
			},
		}
		pullRequests, err := api.ListPullRequests(repository)
		require.NoError(t, err)
		assert.Empty(t, pullRequests)
	})

	t.Run("returns a mock pull request for a burrito-sync repository", func(t *testing.T) {
		repository := &configv1alpha1.TerraformRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "burrito-sync", Namespace: "default"},
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{Url: "https://github.com/padok-team/burrito-sync"},
			},
		}
		pullRequests, err := api.ListPullRequests(repository)
		require.NoError(t, err)
		require.Len(t, pullRequests, 1)

		pr := pullRequests[0]
		assert.Equal(t, "burrito-sync-9001", pr.Name)
		assert.Equal(t, "feature-sync", pr.Spec.Branch)
		assert.Equal(t, "main", pr.Spec.Base)
		assert.Equal(t, "9001", pr.Spec.ID)
		assert.Equal(t, "mock-remote-commit", pr.Annotations[annotations.LastBranchCommit])
	})
}

func TestAPIProvider_GetChanges(t *testing.T) {
	api := &APIProvider{}
	repository := &configv1alpha1.TerraformRepository{ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"}}

	t.Run("returns the readme-only changeset for the reserved PR id", func(t *testing.T) {
		pr := &configv1alpha1.TerraformPullRequest{Spec: configv1alpha1.TerraformPullRequestSpec{ID: "100"}}
		changes, err := api.GetChanges(repository, pr)
		require.NoError(t, err)
		assert.Equal(t, []string{"README.md"}, changes)
	})

	t.Run("returns terraform changes for any other PR id", func(t *testing.T) {
		pr := &configv1alpha1.TerraformPullRequest{Spec: configv1alpha1.TerraformPullRequestSpec{ID: "1"}}
		changes, err := api.GetChanges(repository, pr)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"terraform/main.tf", "terragrunt/inputs.hcl"}, changes)
	})
}

func TestAPIProvider_Comment(t *testing.T) {
	api := &APIProvider{}
	repository := &configv1alpha1.TerraformRepository{}
	pr := &configv1alpha1.TerraformPullRequest{}
	require.NoError(t, api.Comment(repository, pr, &fakeComment{}))
}

func TestGitProvider_Bundle(t *testing.T) {
	t.Run("fails for the reserved unknown URL", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{Url: "https://git.mock.com/unknown"},
			},
		}}
		_, err := p.Bundle("main")
		require.Error(t, err)
	})

	t.Run("returns a deterministic bundle otherwise", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{
			ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
		}}
		bundle, err := p.Bundle("main")
		require.NoError(t, err)
		assert.Equal(t, "bundle:default/repo/main", string(bundle))
	})
}

func TestGitProvider_GetChanges(t *testing.T) {
	t.Run("fails for the reserved unknown URL", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{Url: "https://git.mock.com/unknown"},
			},
		}}
		assert.Nil(t, p.GetChanges("a", "b"))
	})

	t.Run("returns known changes for the reserved revision", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{}}
		changes := p.GetChanges("LAST_RELEVANT_REVISION", "new-rev")
		assert.ElementsMatch(t, []string{"layer-with-files-changed/main.tf", "other-files-changed/inputs.hcl"}, changes)
	})

	t.Run("returns no changes otherwise", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{}}
		assert.Empty(t, p.GetChanges("a", "b"))
	})
}

func TestGitProvider_GetLatestRevisionForRef(t *testing.T) {
	t.Run("fails for the reserved unknown URL", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{
			Spec: configv1alpha1.TerraformRepositorySpec{
				Repository: configv1alpha1.TerraformRepositoryRepository{Url: "https://git.mock.com/unknown"},
			},
		}}
		_, err := p.GetLatestRevisionForRef("main")
		require.Error(t, err)
	})

	t.Run("returns the mock revision otherwise", func(t *testing.T) {
		p := &GitProvider{repository: &configv1alpha1.TerraformRepository{}}
		rev, err := p.GetLatestRevisionForRef("main")
		require.NoError(t, err)
		assert.Equal(t, GetMockRevision("main"), rev)
	})
}
