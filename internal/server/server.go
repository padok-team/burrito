package server

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/server/api"
	"github.com/padok-team/burrito/internal/server/auth"
	"github.com/padok-team/burrito/internal/webhook"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const cookieName = "burrito_session"

//go:embed all:dist
var content embed.FS

type Server struct {
	config       *config.Config
	Webhook      *webhook.Webhook
	API          *api.API
	staticAssets http.FileSystem
	client       client.Client
	sessionStore sessions.Store
	sessionKey   []byte
	authConfig   *auth.AuthConfig
}

func New(c *config.Config) *Server {
	// Generate a random session key
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		log.Fatalf("failed to generate session key: %v", err)
	}

	return &Server{
		config:       c,
		Webhook:      webhook.New(c),
		API:          api.New(c),
		staticAssets: http.FS(content),
		sessionStore: sessions.NewCookieStore(sessionKey),
	}
}

func initClient() (*client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return &cl, nil
}

func (s *Server) Exec() {
	datastore := datastore.NewDefaultClient(s.config.Datastore)
	s.API.Datastore = datastore
	client, err := initClient()
	if err != nil {
		log.Fatalf("error initializing client: %s", err)
	}
	s.client = *client
	s.API.Client = s.client
	s.Webhook.Client = s.client
	log.Info("initializing webhook handlers...")
	err = s.Webhook.Init()
	if err != nil {
		log.Fatalf("error initializing webhook handler: %s", err)
	}

	s.authConfig = &auth.AuthConfig{}
	err = s.authConfig.InitOIDC(s.config)
	if err != nil {
		log.Fatalf("error initializing OIDC: %s", err)
	}
	// Initialize Echo server
	log.Infof("starting burrito server...")
	e := echo.New()

	// Setup session middleware
	e.Use(session.Middleware(s.sessionStore))

	e.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Filesystem: s.staticAssets,
			Root:       "dist",
			Index:      "index.html",
			HTML5:      true,
			Skipper: func(c echo.Context) bool {
				p := c.Request().URL.Path
				return strings.HasPrefix(p, "/auth") ||
					strings.HasPrefix(p, "/api")
			},
		},
	))

	// Auth routes (no authentication required)
	auth := e.Group("/auth")
	auth.GET("/login", s.handleLogin)
	auth.GET("/callback", s.handleCallback)
	auth.POST("/logout", s.handleLogout)

	api := e.Group("/api")
	api.Use(middleware.Logger())
	e.GET("/healthz", handleHealthz)
	api.Use(s.authMiddleware())
	api.POST("/webhook", s.Webhook.GetHttpHandler())
	api.GET("/layers", s.API.LayersHandler)
	api.POST("/layers/:namespace/:layer/sync", s.API.SyncLayerHandler)
	api.GET("/repositories", s.API.RepositoriesHandler)
	api.GET("/logs/:namespace/:layer/:run/:attempt", s.API.GetLogsHandler)
	api.GET("/run/:namespace/:layer/:run/attempts", s.API.GetAttemptsHandler)

	// Redirect root to layers if authenticated, otherwise to login
	e.GET("/", func(c echo.Context) error {
		sess, err := session.Get(cookieName, c)
		if err != nil || sess.Values["user_id"] == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/layers")
	})

	// start a goroutine to refresh webhook handlers every minute
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.refreshWebhookHandlers(ctx)

	e.Logger.Fatal(e.Start(s.config.Server.Addr))
	log.Infof("burrito server started on addr %s", s.config.Server.Addr)
}

func handleHealthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (s *Server) refreshWebhookHandlers(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debug("refreshing webhook handlers...")
			err := s.Webhook.Init()
			if err != nil {
				log.Errorf("error refreshing webhook handlers: %s", err)
			} else {
				log.Debug("webhook handlers refreshed successfully")
			}
		case <-ctx.Done():
			log.Info("stopping refresh of webhook handlers")
			return
		}
	}
}

func (s *Server) authMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(cookieName, c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			userID, ok := sess.Values["user_id"]
			if !ok || userID == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			// Store user info in context for use in handlers
			c.Set("user_id", userID)
			return next(c)
		}
	}
}

func (s *Server) handleLogin(c echo.Context) error {
	// Generate state parameter for CSRF protection
	state := generateRandomString(32)

	sess, err := session.Get(cookieName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
	}

	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	authURL := s.authConfig.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (s *Server) handleCallback(c echo.Context) error {
	sess, err := session.Get(cookieName, c)
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
	token, err := s.authConfig.OAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to exchange code for token")
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No id_token in token response")
	}

	// Verify ID token
	verifier := s.authConfig.OidcProvider.Verifier(&oidc.Config{ClientID: s.authConfig.OAuth2Config.ClientID})
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

func (s *Server) handleLogout(c echo.Context) error {
	sess, err := session.Get(cookieName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
	}

	// Clear session
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("failed to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
