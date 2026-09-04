package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/server/utils"
	"golang.org/x/oauth2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OAuthAuthHandlers struct {
	OidcProvider    *oidc.Provider
	OAuth2Config    *oauth2.Config
	SessionCookie   string
	LoginHTTPMethod string
	RequiredClaims  map[string][]string
}

func New(c *config.Config, ctx context.Context, cl client.Client, sessionCookie string) (*OAuthAuthHandlers, error) {
	oauth := &OAuthAuthHandlers{}

	// Initialize OIDC provider and OAuth2 config
	provider, err := oidc.NewProvider(context.Background(), c.Server.OIDC.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth.OidcProvider = provider
	oauth.OAuth2Config = &oauth2.Config{
		ClientID:     c.Server.OIDC.ClientID,
		ClientSecret: c.Server.OIDC.ClientSecret, // Should be set from env var BURRITO_SERVER_OIDC_CLIENTSECRET
		RedirectURL:  c.Server.OIDC.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       c.Server.OIDC.Scopes,
	}
	oauth.SessionCookie = sessionCookie
	oauth.LoginHTTPMethod = http.MethodGet
	oauth.RequiredClaims = c.Server.OIDC.RequiredClaims

	return oauth, nil
}

func (o *OAuthAuthHandlers) HandleLogin(c echo.Context) error {
	// Generate state parameter for CSRF protection
	state := generateRandomString(32)

	sess, err := session.Get(o.SessionCookie, c)
	if err != nil {
		// Clear session cookie if session is invalid to prevent stale sessions
		err := utils.RemoveSessionCookie(c, o.SessionCookie)
		if err != nil {
			log.Warnf("Failed to clear session cookie: %v", err)
		}
	}

	// State is stored in session for verification in callback handler
	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	authURL := o.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (o *OAuthAuthHandlers) GetLoginHTTPMethod() string {
	return o.LoginHTTPMethod
}

func (o *OAuthAuthHandlers) HandleCallback(c echo.Context) error {
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
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to extract claims")
	}

	// Extract raw claims to check arbitrary/custom claims (e.g. groups, roles) against
	// RequiredClaims, which the typed struct above does not capture.
	var rawClaims map[string]interface{}
	if err := idToken.Claims(&rawClaims); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to extract claims")
	}

	if !claimsSatisfyRequirements(rawClaims, o.RequiredClaims) {
		log.Warnf("OIDC login denied for subject %s: required claims not satisfied", claims.Sub)
		return c.Redirect(http.StatusTemporaryRedirect, "/login?error=access_denied")
	}

	// Upgrade session cookie to SameSite=Strict after successful login
	sess.Options.SameSite = http.SameSiteStrictMode

	// Store user info in session
	sess.Values["user_id"] = claims.Sub
	sess.Values["email"] = claims.Email
	sess.Values["name"] = claims.Name
	sess.Values["picture"] = claims.Picture
	delete(sess.Values, "oauth_state") // Clean up state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	// Redirect to layers page
	return c.Redirect(http.StatusTemporaryRedirect, "/layers")
}

// claimsSatisfyRequirements checks that, for every claim name in required, the ID token's
// claim value (a string, or an array of strings such as a "groups" claim) contains at least
// one of the allowed values. All required claims must be satisfied. An empty or nil required
// map always passes.
func claimsSatisfyRequirements(claims map[string]interface{}, required map[string][]string) bool {
	for claimName, allowedValues := range required {
		value, ok := claims[claimName]
		if !ok {
			return false
		}

		if !containsAny(claimValues(value), allowedValues) {
			return false
		}
	}
	return true
}

// claimValues normalizes a decoded claim value into a slice of strings, supporting both a
// single string claim and an array claim (e.g. "groups").
func claimValues(value interface{}) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []interface{}:
		values := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
		return values
	default:
		return nil
	}
}

func containsAny(values []string, allowed []string) bool {
	for _, v := range values {
		for _, a := range allowed {
			if v == a {
				return true
			}
		}
	}
	return false
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("failed to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
