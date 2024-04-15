package datastore

// import (
// 	"net/http"

// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// 	"github.com/padok-team/burrito/internal/burrito/config"
// 	"github.com/padok-team/burrito/internal/server/api"
// 	"github.com/padok-team/burrito/internal/storage"
// 	log "github.com/sirupsen/logrus"

// 	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
// 	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// )

// type Datastore struct {
// 	config  *config.Config
// 	API     *api.API
// 	client  client.Client
// 	storage *storage.Storage
// }

// func New(c *config.Config) *Datastore {
// 	return &Datastore{
// 		config: c,
// 	}
// }

// func initClient() (*client.Client, error) {
// 	scheme := runtime.NewScheme()
// 	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
// 	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
// 	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
// 		Scheme: scheme,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &cl, nil
// }

// func (s *Datastore) Exec() {
// 	client, err := initClient()
// 	if err != nil {
// 		log.Fatalf("error initializing client: %s", err)
// 	}
// 	s.client = *client
// 	s.API.Client = s.client
// 	log.Infof("starting burrito server...")
// 	e := echo.New()
// 	e.Use(middleware.Logger())
// 	e.Use(middleware.StaticWithConfig(
// 		middleware.StaticConfig{
// 			Filesystem: s.staticAssets,
// 			Root:       "dist",
// 			Index:      "index.html",
// 			HTML5:      true,
// 		},
// 	))
// 	e.GET("/healthz", handleHealthz)
// 	e.POST("/api/webhook", s.Webhook.GetHttpHandler())
// 	e.GET("/api/layers", s.API.LayersHandler)
// 	e.GET("/api/repositories", s.API.RepositoriesHandler)
// 	e.Logger.Fatal(e.Start(s.config.Server.Addr))
// 	log.Infof("burrito server started on addr %s", s.config.Server.Addr)
// }

// func handleHealthz(c echo.Context) error {
// 	return c.String(http.StatusOK, "OK")
// }
