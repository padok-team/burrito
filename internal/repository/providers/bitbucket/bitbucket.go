package bitbucket

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/standard"
	"github.com/padok-team/burrito/internal/repository/types"
)

type Bitbucket struct {
	Config credentials.Credential
}

func (b *Bitbucket) GetWebhookProvider() (types.WebhookProvider, error) {
	return &WebhookProvider{}, nil
}

func (b *Bitbucket) GetAPIProvider() (types.APIProvider, error) {
	client := buildBitbucketClient(b.Config)
	return &APIProvider{
		config: b.Config,
		client: client,
	}, nil
}

func (b *Bitbucket) GetGitProvider(repository *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	auth, err := buildGitCredentials(b.Config)
	if err != nil {
		return nil, err
	}
	return &standard.GitProvider{
		RepoURL:    repository.Spec.Repository.Url,
		AuthMethod: auth,
	}, nil
}

func buildBitbucketClient(config credentials.Credential) *BitbucketClient {
	return &BitbucketClient{
		username: config.Username,
		token:    config.BitbucketToken,
		password: config.Password,
		baseURL:  config.URL,
	}
}

func buildGitCredentials(config credentials.Credential) (transport.AuthMethod, error) {
	if config.BitbucketToken != "" {
		return &http.BasicAuth{
			Username: "x-token-auth",
			Password: config.BitbucketToken,
		}, nil
	} else if config.Username != "" && config.Password != "" {
		return &http.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}, nil
	}
	return nil, fmt.Errorf("no Bitbucket authentication method configured")
}
