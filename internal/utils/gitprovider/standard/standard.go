package standard

import (
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

type Standard struct {
	Config types.Config
}

func IsAvailable(config types.Config, capabilities []string) bool {
	return len(capabilities) == 1 && capabilities[0] == types.Capabilities.Clone
}
func (s *Standard) Init() error {
	return nil
}

func (s *Standard) InitWebhookHandler() error {
	return fmt.Errorf("InitWebhookHandler not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}

func (s *Standard) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	return nil, fmt.Errorf("GetChanges not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}

func (s *Standard) GetLatestRevisionForRef(repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	// Create an in-memory remote
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repository.Spec.Repository.Url},
	})

	// List references on the remote (equivalent to `git ls-remote <repoURL>`)
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list references: %v", err)
	}

	candidates := []string{
		"refs/heads/" + ref,
		"refs/tags/" + ref,
		ref, // in case someone passes the full ref already
	}

	// Look for the ref in the remote’s references
	for _, c := range candidates {
		for _, r := range refs {
			if r.Name().String() == c {
				return r.Hash().String(), nil
			}
		}
	}

	return "", fmt.Errorf("unable to find commit SHA for ref %q in %q", ref, repository.Spec.Repository.Url)
}

func (s *Standard) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	return fmt.Errorf("Comment not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}

func (s *Standard) CreatePullRequest(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) error {
	return fmt.Errorf("CreatePullRequest not supported for standard git provider.  Provide a specific credentials for providers such as GitHub or GitLab")
}

func (g *Standard) Clone(repository *configv1alpha1.TerraformRepository, branch string, repositoryPath string) (*git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           repository.Spec.Repository.Url,
	}
	isSSH := strings.HasPrefix(repository.Spec.Repository.Url, "git@") || strings.Contains(repository.Spec.Repository.Url, "ssh://")

	if isSSH && g.Config.SSHPrivateKey != "" {
		publicKeys, err := ssh.NewPublicKeys("git", []byte(g.Config.SSHPrivateKey), "")
		if err != nil {
			return nil, err
		}
		cloneOptions.Auth = publicKeys
	} else if g.Config.Username != "" && g.Config.Password != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: g.Config.Username,
			Password: g.Config.Password,
		}
	} else {
		log.Info("No authentication method provided, falling back to unauthenticated clone")
	}

	log.Infof("Cloning remote repository %s on %s branch with git", repository.Spec.Repository.Url, branch)
	repo, err := git.PlainClone(repositoryPath, false, cloneOptions)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (m *Standard) ParseWebhookPayload(payload *nethttp.Request) (interface{}, bool) {
	log.Errorf("ParseWebhookPayload not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
	return nil, false
}

func (m *Standard) GetEventFromWebhookPayload(payload interface{}) (event.Event, error) {
	return nil, fmt.Errorf("GetEventFromWebhookPayload not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}
