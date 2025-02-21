package types

import (
	"net/http"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/webhook/event"
)

type Provider interface {
	GetWebhookProvider() (WebhookProvider, error)
	GetAPIProvider() (APIProvider, error)
	GetGitProvider() (GitProvider, error)
}

type GitProvider interface {
	GetLatestRevisionForRef(repository *configv1alpha1.TerraformRepository, ref string) (string, error)
	Bundle(commit string) ([]byte, error)
	GetChanges(previousCommit, currentCommit string) []string
}

type WebhookProvider interface {
	ParseWebhookPayload(r *http.Request) (interface{}, bool)
	GetEventFromWebhookPayload(interface{}) (event.Event, error)
}

type APIProvider interface {
	GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error)
	Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, comment comment.Comment) error
}
