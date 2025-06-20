package mock

import (
	"fmt"
	"net/http"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/repository/types"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Mock struct{}

func (m *Mock) GetWebhookProvider() (types.WebhookProvider, error) {
	return &WebhookProvider{}, nil
}

func (m *Mock) GetAPIProvider() (types.APIProvider, error) {
	return &APIProvider{}, nil
}

func (m *Mock) GetGitProvider(repository *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	return &GitProvider{}, nil
}

type GitProvider struct{}

func (p *GitProvider) Bundle(ref string) ([]byte, error) {
	log.Infof("Mock provider created bundle")
	return make([]byte, 1), nil
}

// TODO: Implement a fake list for TerraformRepository Controller tests
func (p *GitProvider) GetChanges(previousCommit, currentCommit string) []string {
	log.Infof("Mock provider get changes previous commit / current commit")
	return []string{}
}

const mock_revision = "MOCK_REVISION"

func GetMockRevision(ref string) string {
	return fmt.Sprintf("%s-%s", mock_revision, ref)
}

func (p *GitProvider) GetLatestRevisionForRef(ref string) (string, error) {
	return GetMockRevision(ref), nil
}

type APIProvider struct{}

func (api *APIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	log.Infof("Mock provider all changed files")
	var allChangedFiles []string
	// Handle not useful PR
	if pr.Spec.ID == "100" {
		allChangedFiles = []string{
			"README.md",
		}
		return allChangedFiles, nil
	}
	allChangedFiles = []string{
		"terraform/main.tf",
		"terragrunt/inputs.hcl",
	}
	return allChangedFiles, nil
}

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	log.Infof("Mock provider comment posted")
	return nil
}

type WebhookProvider struct{}

func (w *WebhookProvider) ParseWebhookPayload(payload *http.Request) (interface{}, bool) {
	log.Infof("Mock provider webhook payload parsed")
	return nil, true
}

func (w *WebhookProvider) GetEventFromWebhookPayload(payload interface{}) (event.Event, error) {
	log.Infof("Mock provider webhook event parsed")
	return nil, nil
}
