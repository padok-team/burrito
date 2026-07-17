package server

import (
	"context"
	"crypto/rand"
	"embed"
	"net/http"
	"strings"

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
	"github.com/padok-team/burrito/internal/server/utils"
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
		SameSite: http.SameSiteLaxMode, // Cookie is upgraded to SameSite=Strict after callback
	}
	sessionStore.MaxAge(c.Server.Session.MaxAge)

	return &Server{
		config:       c,
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
	log.SetFormatter(&log.JSONFormatter{})
	bgctx := context.Background()
	datastore := datastore.NewDefaultClient(s.config.Datastore)
	s.API.Datastore = datastore
	client, err := initClient()
	if err != nil {
		log.Fatalf("error initializing client: %s", err)
	}
	s.client = *client
	s.API.Client = s.client
	s.Webhook = webhook.New(s.config, *client)

	// Initialize authentication handlers based on configuration
	var authHandlers a.AuthHandlers
	if s.config.Server.OIDC.Enabled {
		log.Infof("OIDC authentication enabled, issuer: %s", s.config.Server.OIDC.IssuerURL)
		authHandlers, err = oauth.New(s.config, bgctx, *client, cookieName)
		if err != nil {
			log.Fatalf("error initializing OIDC: %s", err)
		}
	} else if s.config.Server.BasicAuth.Enabled {
		log.Info("Basic authentication enabled. Credentials will be stored in burrito-admin-credentials secret in the main burrito namespace.")
		authHandlers, err = basic.New(s.config, bgctx, *client, cookieName)
		if err != nil {
			log.Fatalf("error initializing basic auth: %s", err)
		}
	} else {
		log.Warn("No authentication method enabled! The server is publicly accessible. This is NOT recommended for production environments.")
	}

	// Initialize Echo server
	log.Infof("starting burrito server...")
	e := echo.New()

	// Setup session middleware
	e.Use(session.Middleware(s.sessionStore))

	// Expose static web assets
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
	if s.getAuthEnabled() {
		auth := e.Group("/auth", middleware.RequestLoggerWithConfig(utils.LoggerMiddlewareConfig))
		auth.Add(authHandlers.GetLoginHTTPMethod(), "/login", authHandlers.HandleLogin)
		auth.GET("/callback", authHandlers.HandleCallback)
		auth.POST("/logout", func(c echo.Context) error {
			return a.HandleLogout(c, cookieName)
		})
		auth.GET("/type", func(c echo.Context) error {
			authType := "basic"
			if s.config.Server.OIDC.Enabled {
				authType = "oauth"
			}
			return c.JSON(http.StatusOK, map[string]string{"type": authType})
		})
		auth.GET("/user", s.authMiddleware()(func(c echo.Context) error {
			return a.HandleUserInfo(c)
		}))
	}

	api := e.Group("/api")
	e.GET("/healthz", handleHealthz)

	// Exposed webhook route. Logging middleware is applied here to log unauthenticated requests.
	api.POST("/webhook", middleware.RequestLoggerWithConfig(utils.LoggerMiddlewareConfig)(s.Webhook.GetHttpHandler()))

	if s.getAuthEnabled() {
		api.Use(s.authMiddleware())
	}
	// Logger middleware should be applied after auth middleware to be able log user info
	api.Use(middleware.RequestLoggerWithConfig(utils.LoggerMiddlewareConfig))
	api.GET("/layers", s.API.LayersHandler)
	api.POST("/layers/:namespace/:layer/sync", s.API.SyncLayerHandler)
	api.POST("/layers/:namespace/:layer/apply", s.API.ApplyLayerHandler)
	api.GET("/repositories", s.API.RepositoriesHandler)
	api.GET("/logs/:namespace/:layer/:run/:attempt", s.API.GetLogsHandler)
	api.GET("/run/:namespace/:layer/:run/attempts", s.API.GetAttemptsHandler)

	// Redirect root to layers if authenticated, otherwise to login
	e.GET("/", func(c echo.Context) error {
		if !s.getAuthEnabled() {
			return c.Redirect(http.StatusTemporaryRedirect, "/layers")
		}
		sess, err := session.Get(cookieName, c)
		if err != nil || sess.Values["user_id"] == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/layers")
	})

	e.Logger.Fatal(e.Start(s.config.Server.Addr))
	log.Infof("burrito server started on addr %s", s.config.Server.Addr)
}

func handleHealthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
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
			if picture, ok := sess.Values["picture"].(string); ok {
				c.Set("user_picture", picture)
			}
			return next(c)
		}
	}
}

func (s *Server) getAuthEnabled() bool {
	return s.config.Server.OIDC.Enabled || s.config.Server.BasicAuth.Enabled
}
