package types

import (
	"net/http"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	"github.com/padok-team/burrito/internal/webhook/event"
)

type Provider interface {
	GetWebhookProvider() (WebhookProvider, error)
	GetAPIProvider() (APIProvider, error)
	GetGitProvider(repository *configv1alpha1.TerraformRepository) (GitProvider, error)
}

type GitProvider interface {
	GetLatestRevisionForRef(ref string) (string, error)
	Bundle(ref string) ([]byte, error)
	GetChanges(previousCommit, currentCommit string) []string
}

type WebhookProvider interface {
	ParseWebhookPayload(r *http.Request) (interface{}, bool)
	GetEventFromWebhookPayload(interface{}) (event.Event, error)
}

type APIProvider interface {
	GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error)
	Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, comment comment.Comment) error
	ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error)
	SetStatus(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error
	// GetMergeCommit returns the commit SHA that landed on the target branch once
	// pullRequest has been merged (merge commit, or squash commit for squash merges).
	GetMergeCommit(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) (string, error)
}
