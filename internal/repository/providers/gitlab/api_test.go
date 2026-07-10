package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
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
	client, err := gitlab.NewClient("test-token", gitlab.WithBaseURL(server.URL), gitlab.WithoutRetries())
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
				Url: "https://gitlab.com/owner/repo",
			},
		},
	}
}

func testPullRequest(id string) *configv1alpha1.TerraformPullRequest {
	return &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pr",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformPullRequestSpec{
			ID: id,
		},
	}
}

func TestAPIProvider_Comment_CreatesNewNoteWhenNoneManaged(t *testing.T) {
	created := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprint(w, "[]")
		case http.MethodPost:
			created = true
			fmt.Fprint(w, "{}")
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, created, "expected a new note to be created")
}

func TestAPIProvider_Comment_EditsExistingManagedNote(t *testing.T) {
	edited := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, `[{"id": 7, "body": "previous body\n\n<!-- burrito:pull-request-comment -->"}]`)
	})
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes/7", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		edited = true
		fmt.Fprint(w, "{}")
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, edited, "expected the managed note to be edited instead of duplicated")
}

func TestAPIProvider_Comment_FindsManagedNoteAcrossPages(t *testing.T) {
	edited := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[{"id": 7, "body": "previous body\n\n<!-- burrito:pull-request-comment -->"}]`)
			return
		}
		w.Header().Set("X-Next-Page", "2")
		w.Header().Set("X-Total-Pages", "2")
		fmt.Fprint(w, `[{"id": 1, "body": "unrelated note"}]`)
	})
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes/7", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		edited = true
		fmt.Fprint(w, "{}")
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, edited, "expected the managed note on the second page to be found and edited")
}

func TestAPIProvider_ListPullRequests_ReturnsMappedMergeRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, `[{"iid": 7, "source_branch": "feature", "target_branch": "main", "sha": "abc123"}]`)
	})

	api := newTestAPIProvider(t, mux)
	pullRequests, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, pullRequests, 1)

	pr := pullRequests[0]
	assert.Equal(t, "repo-7", pr.Name)
	assert.Equal(t, "feature", pr.Spec.Branch)
	assert.Equal(t, "main", pr.Spec.Base)
	assert.Equal(t, "7", pr.Spec.ID)
	assert.Equal(t, "abc123", pr.Annotations[annotations.LastBranchCommit])
}

func TestAPIProvider_ListPullRequests_FollowsPagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[{"iid": 2, "source_branch": "b", "target_branch": "main", "sha": "sha2"}]`)
			return
		}
		w.Header().Set("X-Next-Page", "2")
		w.Header().Set("X-Total-Pages", "2")
		fmt.Fprint(w, `[{"iid": 1, "source_branch": "a", "target_branch": "main", "sha": "sha1"}]`)
	})

	api := newTestAPIProvider(t, mux)
	pullRequests, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, pullRequests, 2)
}

func TestAPIProvider_ListPullRequests_ReturnsErrorWhenListFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	api := newTestAPIProvider(t, mux)
	_, err := api.ListPullRequests(testRepository())
	require.Error(t, err)
}

func TestAPIProvider_Comment_ReturnsErrorWhenListingNotesFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/merge_requests/42/notes", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.Error(t, err)
}
