package mock

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	log "github.com/sirupsen/logrus"
)

type Mock struct{}

func (m *Mock) Init() error {
	log.Infof("Mock provider initialized")
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
