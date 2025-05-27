package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/burrito/config"
	"golang.org/x/oauth2"
)

type OAuth struct {
	OidcProvider  *oidc.Provider
	OAuth2Config  *oauth2.Config
	SessionCookie string
}

func New(c *config.Config, sessionCookie string) (*OAuth, error) {
	oauth := &OAuth{}

	// Initialize OIDC provider and OAuth2 config
	provider, err := oidc.NewProvider(context.Background(), c.Server.OIDC.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth.OidcProvider = provider
	oauth.OAuth2Config = &oauth2.Config{
		ClientID:     c.Server.OIDC.ClientID,
		ClientSecret: c.Server.OIDC.ClientSecret,
		RedirectURL:  c.Server.OIDC.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	oauth.SessionCookie = sessionCookie

	return oauth, nil
}

func (o *OAuth) HandleLogin(c echo.Context) error {
	// Generate state parameter for CSRF protection
	state := generateRandomString(32)

	sess, err := session.Get(o.SessionCookie, c)
	if err != nil {
		// Clear session cookie if session is invalid
		http.SetCookie(c.Response(), &http.Cookie{
			Name:     o.SessionCookie,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   c.Request().TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// State is stored in session for verification in callback handler
	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	authURL := o.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (o *OAuth) HandleCallback(c echo.Context) error {
	sess, err := session.Get(o.SessionCookie, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
	}

	// Verify state parameter
	state := c.QueryParam("state")
	expectedState, ok := sess.Values["oauth_state"]
	if !ok || state != expectedState {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state parameter")
	}

	// Exchange code for token
	code := c.QueryParam("code")
	token, err := o.OAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to exchange code for token")
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No id_token in token response")
	}

	// Verify ID token
	verifier := o.OidcProvider.Verifier(&oidc.Config{ClientID: o.OAuth2Config.ClientID})
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify ID token")
	}

	// Extract claims
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to extract claims")
	}

	// Store user info in session
	sess.Values["user_id"] = claims.Sub
	sess.Values["email"] = claims.Email
	sess.Values["name"] = claims.Name
	delete(sess.Values, "oauth_state") // Clean up state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	// Redirect to layers page
	return c.Redirect(http.StatusTemporaryRedirect, "/layers")
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("failed to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
