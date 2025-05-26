package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/padok-team/burrito/internal/burrito/config"
	"golang.org/x/oauth2"
)

type AuthConfig struct {
	OidcProvider *oidc.Provider
	OAuth2Config *oauth2.Config
}

func (a *AuthConfig) InitOIDC(c *config.Config) error {
	// Initialize OIDC provider (you'll need to add these to your config)
	provider, err := oidc.NewProvider(context.Background(), c.Server.OIDC.IssuerURL)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %v", err)
	}

	a.OidcProvider = provider
	a.OAuth2Config = &oauth2.Config{
		ClientID:     c.Server.OIDC.ClientID,
		ClientSecret: c.Server.OIDC.ClientSecret,
		RedirectURL:  c.Server.OIDC.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return nil
}
