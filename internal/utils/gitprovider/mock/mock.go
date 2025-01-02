package mock

import (
	"net/http"
	"slices"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Mock struct {
	Config types.Config
}

func IsAvailable(config types.Config, capabilities []string) bool {
	allCapabilities := []string{types.Capabilities.Clone, types.Capabilities.Comment, types.Capabilities.Changes, types.Capabilities.Webhook}
	if !config.EnableMock {
		return false
	}
	for _, c := range capabilities {
		if !slices.Contains(allCapabilities, c) {
			return false
		}
	}
	return true
}

func (m *Mock) Init() error {
	log.Infof("Mock provider initialized")
	return nil
}

func (m *Mock) InitWebhookHandler() error {
	log.Infof("Mock provider webhook handler initialized")
	return nil
}

func (m *Mock) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
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

func (m *Mock) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	log.Infof("Mock provider comment posted")
	return nil
}

func (g *Mock) Clone(repository *configv1alpha1.TerraformRepository, branch string, repositoryPath string) (*git.Repository, error) {
	log.Infof("Mock provider repository cloned")
	return nil, nil
}

func (m *Mock) ParseWebhookPayload(payload *http.Request) (interface{}, bool) {
	log.Infof("Mock provider webhook payload parsed")
	return nil, true
}

func (m *Mock) GetEventFromWebhookPayload(payload interface{}) (event.Event, error) {
	log.Infof("Mock provider webhook event parsed")
	return nil, nil
}
