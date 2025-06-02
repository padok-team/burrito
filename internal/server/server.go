package server

import (
	"context"
	"crypto/rand"
	"embed"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/server/api"
	a "github.com/padok-team/burrito/internal/server/auth"
	"github.com/padok-team/burrito/internal/server/auth/basic"
	"github.com/padok-team/burrito/internal/server/auth/oauth"
	"github.com/padok-team/burrito/internal/webhook"
	log "github.com/sirupsen/logrus"

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
}

func New(c *config.Config) *Server {
	// Generate a random session key
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		log.Fatalf("failed to generate session key: %v", err)
	}

	sessionStore := sessions.NewCookieStore(sessionKey)
	// Set session options
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   c.Server.Session.MaxAge,
		Secure:   c.Server.Session.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	sessionStore.MaxAge(c.Server.Session.MaxAge)

	return &Server{
		config:       c,
		Webhook:      webhook.New(c),
		API:          api.New(c),
		staticAssets: http.FS(content),
		sessionStore: sessionStore,
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
	bgctx := context.Background()
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

	var authHandlers a.AuthHandlers
	if s.config.Server.OIDC.Enabled {
		log.Infof("OIDC authentication enabled, issuer: %s", s.config.Server.OIDC.IssuerURL)
		authHandlers, err = oauth.New(s.config, bgctx, *client, cookieName)
		if err != nil {
			log.Fatalf("error initializing OIDC: %s", err)
		}
	} else {
		log.Info("OIDC authentication is disabled, using basic authentication with default secret")
		authHandlers, err = basic.New(s.config, bgctx, *client, cookieName)
		if err != nil {
			log.Fatalf("error initializing basic auth: %s", err)
		}
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
	auth.Add(authHandlers.GetLoginHTTPMethod(), "/login", authHandlers.HandleLogin)
	auth.GET("/callback", authHandlers.HandleCallback)
	auth.POST("/logout", func(c echo.Context) error {
		return a.HandleLogout(c, cookieName)
	})
	// Return the supported auth type: basic or oauth
	auth.GET("/type", func(c echo.Context) error {
		authType := "basic"
		if s.config.Server.OIDC.Enabled {
			authType = "oauth"
		}
		return c.JSON(http.StatusOK, map[string]string{"type": authType})
	})
	// Check if user is authenticated, used to redirect /login to / if already logged in
	auth.GET("/", func(c echo.Context) error {
		sess, err := session.Get(cookieName, c)
		if err != nil || sess.Values["user_id"] == nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		return c.NoContent(http.StatusOK)
	})

	api := e.Group("/api")
	api.Use(middleware.Logger())
	e.GET("/healthz", handleHealthz)
	api.POST("/webhook", s.Webhook.GetHttpHandler())
	api.Use(s.authMiddleware())
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
			if email, ok := sess.Values["email"].(string); ok {
				c.Set("user_email", email)
			}
			if name, ok := sess.Values["name"].(string); ok {
				c.Set("user_name", name)
			}
			return next(c)
		}
	}
}
