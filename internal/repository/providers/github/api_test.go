package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v84/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
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
	client := github.NewClient(nil)
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client.BaseURL = baseURL
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
				Url: "https://github.com/owner/repo",
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

func TestAPIProvider_Comment_CreatesNewCommentWhenNoneManaged(t *testing.T) {
	created := false
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]*github.IssueComment{})
		case http.MethodPost:
			created = true
			_ = json.NewEncoder(w).Encode(&github.IssueComment{})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, created, "expected a new comment to be created")
}

func TestAPIProvider_Comment_EditsExistingManagedComment(t *testing.T) {
	edited := false
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode([]*github.IssueComment{
			{
				ID:   github.Ptr(int64(7)),
				Body: github.Ptr("previous body\n\n<!-- burrito:pull-request-comment -->"),
			},
		})
	})
	mux.HandleFunc("/repos/owner/repo/issues/comments/7", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		edited = true
		_ = json.NewEncoder(w).Encode(&github.IssueComment{})
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, edited, "expected the managed comment to be edited instead of duplicated")
}

func TestAPIProvider_Comment_FindsManagedCommentAcrossPages(t *testing.T) {
	edited := false
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode([]*github.IssueComment{
				{ID: github.Ptr(int64(7)), Body: github.Ptr("previous body\n\n<!-- burrito:pull-request-comment -->")},
			})
			return
		}
		w.Header().Set("Link", `<https://example.com?page=2>; rel="next"`)
		_ = json.NewEncoder(w).Encode([]*github.IssueComment{{ID: github.Ptr(int64(1)), Body: github.Ptr("unrelated comment")}})
	})
	mux.HandleFunc("/repos/owner/repo/issues/comments/7", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		edited = true
		_ = json.NewEncoder(w).Encode(&github.IssueComment{})
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.NoError(t, err)
	assert.True(t, edited, "expected the managed comment on the second page to be found and edited")
}

func TestAPIProvider_ListPullRequests_ReturnsMappedPullRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode([]*github.PullRequest{
			{
				Number: github.Ptr(7),
				Head: &github.PullRequestBranch{
					Ref: github.Ptr("feature"),
					SHA: github.Ptr("abc123"),
				},
				Base: &github.PullRequestBranch{
					Ref: github.Ptr("main"),
				},
			},
			nil,
		})
	})

	api := newTestAPIProvider(t, mux)
	pullRequests, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, pullRequests, 1)

	pr := pullRequests[0]
	assert.Equal(t, fmt.Sprintf("%s-%d", "repo", 7), pr.Name)
	assert.Equal(t, "feature", pr.Spec.Branch)
	assert.Equal(t, "main", pr.Spec.Base)
	assert.Equal(t, "7", pr.Spec.ID)
	assert.Equal(t, "abc123", pr.Annotations[annotations.LastBranchCommit])
}

func TestAPIProvider_ListPullRequests_FollowsPagination(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode([]*github.PullRequest{{Number: github.Ptr(2), Head: &github.PullRequestBranch{Ref: github.Ptr("b"), SHA: github.Ptr("sha2")}, Base: &github.PullRequestBranch{Ref: github.Ptr("main")}}})
			return
		}
		w.Header().Set("Link", `<https://example.com?page=2>; rel="next"`)
		_ = json.NewEncoder(w).Encode([]*github.PullRequest{{Number: github.Ptr(1), Head: &github.PullRequestBranch{Ref: github.Ptr("a"), SHA: github.Ptr("sha1")}, Base: &github.PullRequestBranch{Ref: github.Ptr("main")}}})
	})

	api := newTestAPIProvider(t, mux)
	pullRequests, err := api.ListPullRequests(testRepository())
	require.NoError(t, err)
	require.Len(t, pullRequests, 2)
}

func TestAPIProvider_ListPullRequests_ReturnsErrorWhenListFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	api := newTestAPIProvider(t, mux)
	_, err := api.ListPullRequests(testRepository())
	require.Error(t, err)
}

func TestAPIProvider_Comment_ReturnsErrorWhenListingCommentsFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/issues/42/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	api := newTestAPIProvider(t, mux)
	err := api.Comment(testRepository(), testPullRequest("42"), &fakeComment{body: "hello"})
	require.Error(t, err)
}

func TestAPIProvider_GetMergeCommit_ReturnsMergeCommitSHA(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls/42", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode(&github.PullRequest{
			MergeCommitSHA: github.Ptr("merge-sha-123"),
		})
	})

	api := newTestAPIProvider(t, mux)
	commit, err := api.GetMergeCommit(testRepository(), testPullRequest("42"))
	require.NoError(t, err)
	assert.Equal(t, "merge-sha-123", commit)
}

func TestAPIProvider_GetMergeCommit_ReturnsErrorWhenGetFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/pulls/42", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	api := newTestAPIProvider(t, mux)
	_, err := api.GetMergeCommit(testRepository(), testPullRequest("42"))
	require.Error(t, err)
}

func TestAPIProvider_SetStatus_MapsRunningToPending(t *testing.T) {
	var gotState string
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/statuses/sha123", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			State string `json:"state"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		gotState = body.State
		fmt.Fprint(w, `{}`)
	})

	api := newTestAPIProvider(t, mux)
	err := api.SetStatus(testRepository(), testPullRequest("42"), status.CommitStatus{
		Phase:  status.PhasePlan,
		State:  status.StateRunning,
		Commit: "sha123",
	})
	require.NoError(t, err)
	assert.Equal(t, "pending", gotState)
}

func TestToGithubState(t *testing.T) {
	assert.Equal(t, "pending", toGithubState(status.StateRunning))
	assert.Equal(t, "pending", toGithubState(status.StatePending))
	assert.Equal(t, "success", toGithubState(status.StateSuccess))
	assert.Equal(t, "failure", toGithubState(status.StateFailure))
}
