package standard

import (
	"fmt"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/utils/gitprovider/common"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
)

const (
	WorkingDir = "/tmp/burrito/repositories"
	BundleDir  = "/tmp/burrito/gitbundles"
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
	auth, err := s.getGitAuth(repository.Spec.Repository.Url)
	if err != nil {
		return "", fmt.Errorf("failed to get git auth: %w", err)
	}

	// Create an in-memory remote
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repository.Spec.Repository.Url},
	})

	// List references on the remote (equivalent to `git ls-remote <repoURL>`)
	refs, err := remote.List(&git.ListOptions{
		Auth: auth,
	})
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
	auth, err := g.getGitAuth(repository.Spec.Repository.Url)
	if err != nil {
		return nil, err
	}

	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           repository.Spec.Repository.Url,
		Auth:          auth,
	}

	if auth == nil {
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

func (s *Standard) getGitAuth(repoURL string) (transport.AuthMethod, error) {
	isSSH := strings.HasPrefix(repoURL, "git@") || strings.Contains(repoURL, "ssh://")

	if isSSH && s.Config.SSHPrivateKey != "" {
		publicKeys, err := ssh.NewPublicKeys("git", []byte(s.Config.SSHPrivateKey), "")
		if err != nil {
			return nil, err
		}
		return publicKeys, nil
	} else if s.Config.Username != "" && s.Config.Password != "" {
		return &http.BasicAuth{
			Username: s.Config.Username,
			Password: s.Config.Password,
		}, nil
	}
	log.Info("no authentication method provided, falling back to unauthenticated clone")
	return nil, nil
}

func (s *Standard) GetGitBundle(repository *configv1alpha1.TerraformRepository, ref string, revision string) ([]byte, error) {
	repoKey := fmt.Sprintf("%s-%s-%s", repository.Namespace, repository.Name, strings.ReplaceAll(ref, "/", "--"))
	repoDir := filepath.Join(WorkingDir, repoKey)

	auth, err := s.getGitAuth(repository.Spec.Repository.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to get git auth: %w", err)
	}

	// Try to open existing repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return nil, fmt.Errorf("failed to open repository %s: %w", repoKey, err)
		}

		// Clone if it doesn't exist
		log.Infof("Cloning repository %s to %s", repository.Spec.Repository.Url, repoDir)
		cloneOpts := &git.CloneOptions{
			URL:           repository.Spec.Repository.Url,
			Auth:          auth,
			ReferenceName: plumbing.NewBranchReferenceName(ref),
		}

		repo, err = git.PlainClone(repoDir, false, cloneOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository %s: %w", repoKey, err)
		}
	}

	// Fetch latest changes
	fetchOpts := &git.FetchOptions{
		Auth: auth,
	}

	log.Infof("fetching latest changes for repo %s", repoKey)
	err = repo.Fetch(fetchOpts)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Infof("repository %s is already up-to-date", repoKey)
		} else {
			return nil, fmt.Errorf("failed to fetch latest changes: %w", err)
		}
	}

	// Create BundleDir if it doesn't exist
	if _, err := os.Stat(BundleDir); os.IsNotExist(err) {
		if err := os.MkdirAll(BundleDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create BundleDir directory: %v", err)
		}
	}
	bundleDest := filepath.Join(BundleDir, fmt.Sprintf("%s.gitbundle", repoKey))
	bundle, err := common.CreateGitBundle(repoDir, bundleDest, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle: %w", err)
	}

	return bundle, nil
}
