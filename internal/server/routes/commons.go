package routes

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/storage"
	"github.com/padok-team/burrito/internal/storage/redis"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LayerClient struct {
	client  client.Client
	config  *config.Config
	storage storage.Storage
}

func NewLayerClient(c *config.Config) *LayerClient {
	return &LayerClient{
		config: c,
	}
}

func (l *LayerClient) Init() error {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}
	l.client = cl

	// TODO: Fix configuration
	// l.config.Redis.URL = "burrito-redis:6379"
	log.Infof("Connecting to redis at %s", l.config.Redis.URL)
	l.storage = redis.New(l.config.Redis.URL, l.config.Redis.Password, l.config.Redis.Database)
	return nil
}
