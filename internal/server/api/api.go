package api

import (
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type API struct {
	config    *config.Config
	Client    client.Client
	Datastore datastore.Client
}

func New(c *config.Config) *API {
	client := datastore.NewDefaultClient()
	if c.Datastore.TLS {
		client.Scheme = "https"
	}
	return &API{
		config:    c,
		Datastore: client,
	}
}
