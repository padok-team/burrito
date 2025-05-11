package server

import (
	"embed"
	"net/http"

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
	s.Webhook = webhook.New(s.config, s.client)
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
	api.GET("/repositories", s.API.RepositoriesHandler)
	api.GET("/logs/:namespace/:layer/:run/:attempt", s.API.GetLogsHandler)
	api.GET("/run/:namespace/:layer/:run/attempts", s.API.GetAttemptsHandler)

	e.Logger.Fatal(e.Start(s.config.Server.Addr))
	log.Infof("burrito server started on addr %s", s.config.Server.Addr)
}

func handleHealthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
