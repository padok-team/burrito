package gitea

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gitea "gitea.dev/sdk"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/repository/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeComment struct {
	body string
}

func (c *fakeComment) Generate(commit string) (string, error) {
	return c.body, nil
}

func newTestAPIProvider(t *testing.T, mux *http.ServeMux) *APIProvider {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client, err := gitea.NewClient(server.URL, gitea.SetToken("test-token"))
	require.NoError(t, err)
	return &APIProvider{client: client}
}

func testRepository() *configv1alpha1.TerraformRepository {
	return &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repo",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://gitea.example.com/owner/repo",
			},
		},
	}
}

func testPullRequest(id string) *configv1alpha1.TerraformPullRequest {
	return &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "abc123",
			},
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			ID: id,
		},
	}
}

func TestAPIProvider_GetChanges(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/pulls/42/files", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `[{"filename":"terraform/main.tf","status":"added"},{"filename":"README.md","status":"added"}]`)
	})

	api := newTestAPIProvider(t, mux)
	changes, err := api.GetChanges(testRepository(), testPullRequest("42"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"terraform/main.tf", "README.md"}, changes)
}

func TestAPIProvider_SetStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/statuses/abc123", func(w http.ResponseWriter, r *http.Request) {
		var opts gitea.CreateStatusOption
		err := json.NewDecoder(r.Body).Decode(&opts)
		require.NoError(t, err)
		assert.Equal(t, gitea.StatusSuccess, opts.State)
		assert.Equal(t, "burrito/plan", opts.Context)
		_, _ = fmt.Fprint(w, `{}`)
	})

	api := newTestAPIProvider(t, mux)
	err := api.SetStatus(testRepository(), testPullRequest("42"), types.CommitStatus{
		Phase:       types.PhasePlan,
		State:       types.StateSuccess,
		Description: "plan succeeded",
	})
	require.NoError(t, err)
}

func TestAPIProvider_Comment_CreatesNewWhenNoneManaged(t *testing.T) {
	created := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = fmt.Fprint(w, `[]`)
		case http.MethodPost:
			created = true
			_, _ = fmt.Fprint(w, `{}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, created, "expected a new comment to be created")
}

func TestAPIProvider_Comment_UpdatesExistingManaged(t *testing.T) {
	updated := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = fmt.Fprint(w, `[{"id":7,"body":"<!-- burrito:pull-request-comment --> old text"}]`)
		case http.MethodPost:
			t.Fatalf("POST should not be called when a managed comment exists")
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/api/v1/repos/owner/repo/issues/comments/7", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		updated = true
		_, _ = fmt.Fprint(w, `{}`)
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, updated, "expected the managed comment to be updated")
}

func TestAPIProvider_GetPullRequestTitle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/pulls/42", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"title":"Add demo layer"}`)
	})

	api := newTestAPIProvider(t, mux)
	title, err := api.GetPullRequestTitle(testRepository(), testPullRequest("42"))
	require.NoError(t, err)
	assert.Equal(t, "Add demo layer", title)
}

func TestAPIProvider_ListPullRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "open", r.URL.Query().Get("state"))
		_, _ = fmt.Fprint(w, `[{"number":42,"title":"Add demo layer","base":{"ref":"main"},"head":{"ref":"demo","sha":"abc123"}}]`)
	})

	api := newTestAPIProvider(t, mux)
	prs, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "repo-42", prs[0].Name)
	assert.Equal(t, "demo", prs[0].Spec.Branch)
	assert.Equal(t, "main", prs[0].Spec.Base)
	assert.Equal(t, "42", prs[0].Spec.ID)
	assert.Equal(t, "abc123", prs[0].Annotations[annotations.LastBranchCommit])
}

func TestAPIProvider_ListPullRequests_SkipsMalformedEntries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		// first entry is missing head/base, second is well-formed
		_, _ = fmt.Fprint(w, `[{"number":43,"title":"broken"},{"number":42,"title":"Add demo layer","base":{"ref":"main"},"head":{"ref":"demo","sha":"abc123"}}]`)
	})

	api := newTestAPIProvider(t, mux)
	prs, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "42", prs[0].Spec.ID)
}
