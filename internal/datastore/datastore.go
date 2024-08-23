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

const (
	DefaultCertPath = "/etc/burrito/tls/tls.crt"
	DefaultKeyPath  = "/etc/burrito/tls/tls.key"
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
	api.Use(middleware.RequestLoggerWithConfig(getLoggerConfig()))
	api.Use(authz.Process)
	api.GET("/logs", s.API.GetLogsHandler)
	api.PUT("/logs", s.API.PutLogsHandler)
	api.GET("/plans", s.API.GetPlanHandler)
	api.PUT("/plans", s.API.PutPlanHandler)
	if s.Config.Datastore.TLS {
		e.Logger.Fatal(e.StartTLS(":8080", DefaultCertPath, DefaultKeyPath))
	} else {
		e.Logger.Fatal(e.Start(":8080"))
	}
}

func handleHealthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func getLoggerConfig() middleware.RequestLoggerConfig {
	return middleware.RequestLoggerConfig{
		LogRemoteIP:      true,
		LogMethod:        true,
		LogURI:           true,
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			user, _ := c.Get("serviceAccount").(string)
			log.WithFields(log.Fields{
				"service_account": user,
				"remote_ip":       v.RemoteIP,
				"method":          v.Method,
				"uri":             v.URI,
				"status":          v.Status,
				"error":           v.Error,
				"latency":         v.Latency.String(),
				"bytes_in":        v.ContentLength,
				"bytes_out":       v.ResponseSize,
			}).Info()
			return nil
		},
	}
}
