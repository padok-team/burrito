package api

import (
	"github.com/padok-team/burrito/internal/burrito/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type API struct {
	config *config.Config
	Client client.Client
}

func New(c *config.Config) *API {
	return &API{
		config: c,
	}
}
