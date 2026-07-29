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
	var provider types.Provider
	switch RepositoryCredentials.Provider {
	case "github":
		provider = &github.Github{Config: RepositoryCredentials}
	case "gitlab":
		provider = &gitlab.Gitlab{Config: RepositoryCredentials}
	case "standard":
		provider = &standard.Standard{Config: RepositoryCredentials}
	case "mock":
		provider = &mock.Mock{}
	default:
		return nil, fmt.Errorf("unknown provider: %s", RepositoryCredentials.Provider)
	}
	// Wrap every provider so the webhook-secret invariant is enforced centrally,
	// in the single code path all providers are constructed through, rather than
	// relying on each provider (present or future) to remember it. See
	// webhookSecretEnforcer for details.
	return &webhookSecretEnforcer{Provider: provider, credential: RepositoryCredentials}, nil
}

// webhookSecretEnforcer wraps a types.Provider to guarantee that a webhook
// provider is never handed out when the credential has no webhook secret.
//
// This exists for security reasons. An empty webhook secret makes the
// underlying go-playground/webhooks parser skip signature/token verification
// entirely (`if len(hook.secret) > 0`), which would let anyone POST forged,
// unsigned events to /api/webhook and drive real Terraform plans under the
// operator's own credentials. Because GetProviderFromCredentials is the single
// choke point every provider is built through, enforcing the check here makes
// it structurally impossible for any provider — including ones added in the
// future — to serve webhooks without a configured secret. Providers therefore
// do not (and must not) need to re-implement this guard themselves.
type webhookSecretEnforcer struct {
	types.Provider
	credential credentials.Credential
}

// GetWebhookProvider delegates to the wrapped provider first (so providers that
// do not support webhooks keep returning their own error), then refuses to hand
// back a working webhook provider unless a non-empty webhook secret is set.
func (w *webhookSecretEnforcer) GetWebhookProvider() (types.WebhookProvider, error) {
	inner, err := w.Provider.GetWebhookProvider()
	if err != nil {
		return nil, err
	}
	if w.credential.WebhookSecret == "" {
		return nil, fmt.Errorf("webhook secret is required but not configured for this repository: refusing to serve webhooks without signature verification")
	}
	return inner, nil
}

func getStandardGitNoAuth(URL string) types.GitProvider {
	return &standard.GitProvider{
		RepoURL: URL,
	}
}
