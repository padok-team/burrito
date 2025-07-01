package server

import (
	"context"
	"embed"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/server/api"
	"github.com/padok-team/burrito/internal/webhook"
	log "github.com/sirupsen/logrus"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed all:dist
var content embed.FS

type Server struct {
	config       *config.Config
	Webhook      *webhook.Webhook
	API          *api.API
	staticAssets http.FileSystem
	client       client.Client
}

func New(c *config.Config) *Server {
	return &Server{
		config:       c,
		Webhook:      webhook.New(c),
		API:          api.New(c),
		staticAssets: http.FS(content),
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
	log.Infof("starting burrito server...")
	e := echo.New()
	e.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Filesystem: s.staticAssets,
			Root:       "dist",
			Index:      "index.html",
			HTML5:      true,
		},
	))
	api := e.Group("/api")
	api.Use(middleware.Logger())
	e.GET("/healthz", handleHealthz)
	api.POST("/webhook", s.Webhook.GetHttpHandler())
	api.GET("/layers", s.API.LayersHandler)
	api.POST("/layers/:namespace/:layer/sync", s.API.SyncLayerHandler)
	api.POST("/layers/:namespace/:layer/apply", s.API.ApplyLayerHandler)
	api.GET("/repositories", s.API.RepositoriesHandler)
	api.GET("/logs/:namespace/:layer/:run/:attempt", s.API.GetLogsHandler)
	api.GET("/run/:namespace/:layer/:run/attempts", s.API.GetAttemptsHandler)

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
