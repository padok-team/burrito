package standard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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

func (s *Standard) GetRepositoryProvider() (types.GitProvider, error) {
	repositoryProvider := &GitProvider{
		URL: s.Config.URL,
	}
	repoURL := s.Config.URL
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
