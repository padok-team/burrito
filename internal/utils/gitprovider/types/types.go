package types

import (
	"net/http"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/webhook/event"
)

// Config holds all possible authentication configurations
type Config struct {
	// Basic auth
	Username string
	Password string

	// SSH auth
	SSHPrivateKey string

	// GitHub App auth
	AppID             int64
	AppInstallationID int64
	AppPrivateKey     string

	// Token auth
	GitHubToken string
	GitLabToken string

	// Repository URL
	URL string

	// Mock provider
	EnableMock bool

	// Secret for webhook handling
	WebhookSecret string
}

// Provider interface defines methods that must be implemented by git providers
type Provider interface {
	Init() error
	InitWebhookHandler() error
	GetChanges(*configv1alpha1.TerraformRepository, *configv1alpha1.TerraformPullRequest) ([]string, error)
	Comment(*configv1alpha1.TerraformRepository, *configv1alpha1.TerraformPullRequest, comment.Comment) error
	Clone(*configv1alpha1.TerraformRepository, string, string) (*git.Repository, error)
	GetLatestRevisionForRef(*configv1alpha1.TerraformRepository, string) (string, error)
	ParseWebhookPayload(r *http.Request) (interface{}, bool)
	GetEventFromWebhookPayload(interface{}) (event.Event, error)
}

var Capabilities = struct {
	Clone   string
	Comment string
	Changes string
	Webhook string
}{
	Clone:   "clone",
	Comment: "comment",
	Changes: "changes",
	Webhook: "webhook",
}
