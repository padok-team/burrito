package webhook

import (
	"fmt"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/webhook/event"
	"github.com/padok-team/burrito/internal/webhook/github"
	"github.com/padok-team/burrito/internal/webhook/gitlab"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler interface {
	Handle()
}

type Webhook struct {
	client.Client
	Config    *config.Config
	Providers []Provider
}

func New(c *config.Config) *Webhook {
	return &Webhook{
		Config: c,
	}
}

type Provider interface {
	Init(*config.Config) error
	IsFromProvider(*http.Request) bool
	GetEvent(*http.Request) (event.Event, error)
}

func (w *Webhook) Init() error {
	providers := []Provider{}
	for _, p := range []Provider{&github.Github{}, &gitlab.Gitlab{}} {
		err := p.Init(w.Config)
		if err != nil {
			log.Warnf("failed to initialize webhook provider: %s", err)
			continue
		}
		providers = append(providers, p)
	}
	if len(providers) == 0 {
		log.Warnf("no webhook provider initialized, every event will be considered as unknown")
	}
	w.Providers = providers
	return nil
}

func (w *Webhook) GetHttpHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		log.Infof("webhook event received...")
		r := c.Request()
		var err error
		var event event.Event
		for _, p := range w.Providers {
			if p.IsFromProvider(r) {
				event, err = p.GetEvent(r)
				break
			}
		}
		if err != nil {
			log.Errorf("webhook processing failed: %s", err)
			status := http.StatusBadRequest
			if r.Method != "POST" {
				status = http.StatusMethodNotAllowed
			}
			return c.String(status, fmt.Sprintf("webhook processing failed: %s", html.EscapeString(err.Error())))
		}
		if event == nil {
			log.Infof("ignoring unknown webhook event")
			return c.String(http.StatusBadRequest, "Unknown webhook event")
		}

		err = event.Handle(w.Client)
		if err != nil {
			log.Errorf("webhook processing worked but errored during event handling: %s", err)
		}
		return c.String(http.StatusOK, "OK")
	}
}
