package repository

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/github"
	"github.com/padok-team/burrito/internal/repository/providers/gitlab"
	"github.com/padok-team/burrito/internal/repository/providers/mock"
	"github.com/padok-team/burrito/internal/repository/providers/standard"
	"github.com/padok-team/burrito/internal/repository/types"
)

func GetGitProviderFromRepository(store *credentials.CredentialStore, repo *configv1alpha1.TerraformRepository) (types.GitProvider, error) {
	creds, err := store.GetCredentials(repo)
	// If no credentials, it may be a standard public repository
	if err != nil {
		return getStandardGitNoAuth(repo.Spec.Repository.Url), nil
	}
	provider, err := GetProviderFromCredentials(*creds)
	if err != nil {
		return nil, err
	}

	return provider.GetGitProvider(repo)
}

func GetAPIProviderFromRepository(store *credentials.CredentialStore, repo *configv1alpha1.TerraformRepository) (types.APIProvider, error) {
	creds, err := store.GetCredentials(repo)
	if err != nil {
		return nil, err
	}
	provider, err := GetProviderFromCredentials(*creds)
	if err != nil {
		return nil, err
	}

	return provider.GetAPIProvider()
}

func GetProviderFromCredentials(RepositoryCredentials credentials.Credential) (types.Provider, error) {
	switch RepositoryCredentials.Provider {
	case "github":
		return &github.Github{Config: RepositoryCredentials}, nil
	case "gitlab":
		return &gitlab.Gitlab{Config: RepositoryCredentials}, nil
	case "standard":
		return &standard.Standard{Config: RepositoryCredentials}, nil
	case "mock":
		return &mock.Mock{}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", RepositoryCredentials.Provider)
	}
}

func getStandardGitNoAuth(URL string) types.GitProvider {
	return &standard.GitProvider{
		RepoURL: URL,
	}
}
