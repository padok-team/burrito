package api

import (
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage"
)

type API struct {
	config  *config.Config
	Storage storage.Storage
}

func New(c *config.Config) *API {
	return &API{
		config: c,
	}
}
