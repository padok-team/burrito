package repository_test

import (
	"testing"

	"github.com/padok-team/burrito/internal/repository"
	"github.com/padok-team/burrito/internal/repository/credentials"

	"github.com/stretchr/testify/assert"
)

// allProviderKeys lists every provider key handled by
// repository.GetProviderFromCredentials. That switch statement is the single,
// mandatory registration point for providers (see internal/repository/AGENTS.md:
// "Adding a provider = add a case here"), so keeping this list in sync with it
// guarantees that every provider reachable in production is exercised by the
// non-regression test below.
var allProviderKeys = []string{"github", "gitlab", "standard", "mock"}

// TestGetWebhookProvider_RejectsEmptyWebhookSecret is a security non-regression
// test: no provider must ever hand back a usable webhook provider when the
// webhook secret is empty. An empty secret makes go-playground/webhooks skip
// HMAC/token verification entirely, which would let anyone POST forged,
// unsigned events to /api/webhook. The invariant is enforced centrally in
// GetProviderFromCredentials, so this test drives every provider through that
// factory to prove the enforcement is wired in for all of them.
func TestGetWebhookProvider_RejectsEmptyWebhookSecret(t *testing.T) {
	for _, key := range allProviderKeys {
		t.Run(key, func(t *testing.T) {
			provider, err := repository.GetProviderFromCredentials(credentials.Credential{
				Provider:      key,
				WebhookSecret: "",
			})
			assert.NoError(t, err, "provider %q should be constructible", key)

			_, err = provider.GetWebhookProvider()
			assert.Error(t, err, "provider %q must return an error when the webhook secret is empty", key)
		})
	}
}

// TestGetWebhookProvider_AcceptsNonEmptyWebhookSecret guards the legitimate,
// correctly-configured path: providers that support webhooks must still return
// a working webhook provider when a secret is configured.
func TestGetWebhookProvider_AcceptsNonEmptyWebhookSecret(t *testing.T) {
	for _, key := range []string{"github", "gitlab"} {
		t.Run(key, func(t *testing.T) {
			provider, err := repository.GetProviderFromCredentials(credentials.Credential{
				Provider:      key,
				WebhookSecret: "a-real-secret",
			})
			assert.NoError(t, err, "provider %q should be constructible", key)

			whProvider, err := provider.GetWebhookProvider()
			assert.NoError(t, err, "provider %q should accept a non-empty webhook secret", key)
			assert.NotNil(t, whProvider)
		})
	}
}
