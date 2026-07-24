package mock

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	"github.com/padok-team/burrito/internal/repository/types"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Mock struct{}

func (m *Mock) GetWebhookProvider() (types.WebhookProvider, error) {
	return &WebhookProvider{}, nil
}

func (m *Mock) GetAPIProvider() (types.APIProvider, error) {
	return &APIProvider{}, nil
}

func (m *Mock) GetGitProvider(repository *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	return &GitProvider{repository: repository}, nil
}

type GitProvider struct {
	repository *configv1alpha1.TerraformRepository
}

// Internal function for the mock provider to make it fail on specific URL for tests purposes
// in the TerraformRepositoryController tests
func (p *GitProvider) testfail() bool {
	return p.repository.Spec.Repository.Url == "https://git.mock.com/unknown"
}

func (p *GitProvider) Bundle(ref string) ([]byte, error) {
	if p.testfail() {
		return nil, errors.New("mock provider: clone failed")
	}
	// Return a unique bundle per namespace/repo/ref so tests can verify isolation
	content := fmt.Sprintf("bundle:%s/%s/%s", p.repository.Namespace, p.repository.Name, ref)
	return []byte(content), nil
}

func (p *GitProvider) GetChanges(previousCommit, currentCommit string) []string {
	if p.testfail() {
		log.Errorf("mock provider: get changed failed")
		return nil
	}

	// Used in TerraformRepository Controller tests, some layers with this revision as last relevant have changes
	if previousCommit == "LAST_RELEVANT_REVISION" {
		log.Infof("mock gitprovider changes detected")
		return []string{
			"layer-with-files-changed/main.tf",
			"other-files-changed/inputs.hcl",
		}
	}

	return []string{}
}

const mock_revision = "MOCK_REVISION"

func GetMockRevision(ref string) string {
	return fmt.Sprintf("%s-%s", mock_revision, ref)
}

func (p *GitProvider) GetLatestRevisionForRef(ref string) (string, error) {
	if p.testfail() {
		return "", errors.New("mock provider: get latest revision failed")
	}

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

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	log.Infof("Mock provider comment posted")
	return nil
}

func (api *APIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	log.Infof("Mock provider listing open pull requests for %s/%s", repository.Namespace, repository.Name)
	if !strings.Contains(repository.Spec.Repository.Url, "burrito-sync") {
		return []configv1alpha1.TerraformPullRequest{}, nil
	}

	return []configv1alpha1.TerraformPullRequest{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", repository.Name, "9001"),
				Namespace: repository.Namespace,
				Annotations: map[string]string{
					annotations.LastBranchCommit: "mock-remote-commit",
				},
			},
			Spec: configv1alpha1.TerraformPullRequestSpec{
				Branch: "feature-sync",
				Base:   "main",
				ID:     "9001",
				Repository: configv1alpha1.TerraformLayerRepository{
					Name:      repository.Name,
					Namespace: repository.Namespace,
				},
			},
		},
	}, nil
}

func (api *APIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	log.Infof("Mock provider status set: phase=%s state=%s", s.Phase, s.State)
	return nil
}

func (api *APIProvider) GetMergeCommit(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) (string, error) {
	return "mock-merge-commit", nil
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
