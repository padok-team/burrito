package standard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/types"
)

type Standard struct {
	Config credentials.Credential
}

func (s *Standard) GetWebhookProvider() (types.WebhookProvider, error) {
	return nil, fmt.Errorf("webhooks are not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}

func (s *Standard) GetAPIProvider() (types.APIProvider, error) {
	return nil, fmt.Errorf("API is not supported for standard git provider. Provide a specific credentials for providers such as GitHub or GitLab")
}

func (s *Standard) GetGitProvider(repository *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	repoURL := repository.Spec.Repository.Url
	repositoryProvider := &GitProvider{
		RepoURL: repoURL,
	}
	if repoURL == "" {
		return nil, errors.New("repository URL is required")
	}
	isSSH := strings.HasPrefix(repoURL, "git@") || strings.Contains(repoURL, "ssh://")

	if isSSH && s.Config.SSHPrivateKey != "" {
		publicKeys, err := ssh.NewPublicKeys("git", []byte(s.Config.SSHPrivateKey), "")
		if err != nil {
			return nil, err
		}
		repositoryProvider.AuthMethod = publicKeys
	} else if s.Config.Username != "" && s.Config.Password != "" {
		repositoryProvider.AuthMethod = &http.BasicAuth{
			Username: s.Config.Username,
			Password: s.Config.Password,
		}
	}
	return repositoryProvider, nil
}
