package webhook

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/burrito/config"
	repo "github.com/padok-team/burrito/internal/repository"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler interface {
	Handle()
}

type Webhook struct {
	client.Client
	Config      *config.Config
	credentials *credentials.CredentialStore
}

func New(config *config.Config, client client.Client) *Webhook {
	return &Webhook{
		Client: client,
		Config: config,
		// TODO: get TTL value from config
		credentials: credentials.NewCredentialStore(client, 5*time.Second),
	}
}

func (w *Webhook) GetHttpHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		log.Infof("webhook event received...")
		r := c.Request()
		event, err := w.tryGetEventFromPayload(r)
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

// Iterate over all webhook secrets we have to decode the payload
func (w *Webhook) tryGetEventFromPayload(r *http.Request) (event.Event, error) {
	allShared, allRepository := w.credentials.GetAllCredentials()

	for _, i := range allShared {
		cred := i.Credential
		provider, err := repo.GetProviderFromCredentials(cred)
		if err != nil {
			continue
		}
		whProvider, err := provider.GetWebhookProvider()
		if err != nil {
			log.Errorf("failed to get webhook provider: %s", err)
			continue
		}
		parsed, ok := whProvider.ParseWebhookPayload(r)
		if ok {
			return whProvider.GetEventFromWebhookPayload(parsed)
		}
	}

	for _, i := range allRepository {
		cred := i.Credential
		provider, err := repo.GetProviderFromCredentials(cred)
		if err != nil {
			continue
		}
		whProvider, err := provider.GetWebhookProvider()
		if err != nil {
			log.Errorf("failed to get webhook provider: %s", err)
			continue
		}
		parsed, ok := whProvider.ParseWebhookPayload(r)
		if ok {
			return whProvider.GetEventFromWebhookPayload(parsed)
		}
	}

	return nil, nil
}
