package datastore

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/api"
	"github.com/padok-team/burrito/internal/datastore/storage"
	"github.com/padok-team/burrito/internal/utils/authz"
	log "github.com/sirupsen/logrus"
)

type Datastore struct {
	Config *config.Config
	API    *api.API
}

func New(c *config.Config) *Datastore {
	return &Datastore{
		Config: c,
	}
}

func (s *Datastore) Exec() {
	s.API = api.New(s.Config)
	s.API.Storage = storage.New(*s.Config)
	authz := authz.NewAuthz()
	for _, sa := range s.Config.Datastore.AuthorizedServiceAccounts {
		l := strings.Split(sa, "/")
		authz.AddServiceAccount(l[0], l[1])
	}
	authz.SetAudience("burrito")
	log.Infof("starting burrito datastore...")
	e := echo.New()
	e.GET("/healthz", handleHealthz)

	api := e.Group("/api")
	api.Use(middleware.Logger())
	api.Use(authz.Process)
	api.GET("/logs", s.API.GetLogsHandler)
	api.PUT("/logs", s.API.PutLogsHandler)
	api.GET("/plans", s.API.GetPlanHandler)
	api.PUT("/plans", s.API.PutPlanHandler)
	api.GET("/attempts", s.API.GetAttemptsHandler)
	e.Logger.Fatal(e.Start(s.Config.Datastore.Addr))
	log.Infof("burrito datastore started on addr %s", s.Config.Datastore.Addr)
}

func handleHealthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
