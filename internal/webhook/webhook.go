package webhook

import (
	"context"
	"fmt"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	"github.com/padok-team/burrito/internal/webhook/event"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler interface {
	Handle()
}

type Webhook struct {
	client.Client
	Config    *config.Config
	Providers map[string][]gitprovider.Provider
}

func New(c *config.Config) *Webhook {
	return &Webhook{
		Config:    c,
		Providers: make(map[string][]gitprovider.Provider),
	}
}

type Provider interface {
	Init() error
	ParseFromProvider(*http.Request) (interface{}, bool)
	GetEvent(interface{}) (event.Event, error)
}

func (w *Webhook) Init() error {
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := w.Client.List(context.Background(), repositories)
	if err != nil {
		return fmt.Errorf("failed to list TerraformRepository objects: %w", err)
	}
	err = w.initializeDefaultProvider()
	if err != nil {
		return fmt.Errorf("Some legacy webhook configuration was found but default providers could not be initialized: %w", err)
	}
	for _, r := range repositories.Items {
		if _, ok := w.Providers[fmt.Sprintf("%s/%s", r.Namespace, r.Name)]; !ok {
			provider, err := w.initializeProviders(r)
			if err != nil {
				log.Errorf("could not initialize provider for repository %s/%s: %s", r.Namespace, r.Name, err)
			}
			if provider != nil {
				w.Providers[fmt.Sprintf("%s/%s", r.Namespace, r.Name)] = provider
				log.Infof("initialized webhook handlers for repository %s/%s", r.Namespace, r.Name)
			}
		}
	}
	return nil
}

func (w *Webhook) GetHttpHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		log.Infof("webhook event received...")
		r := c.Request()
		var err error
		var event event.Event
		for _, ps := range w.Providers {
			for _, p := range ps {
				parsed, ok := p.ParseWebhookPayload(r)
				if ok {
					event, err = p.GetEventFromWebhookPayload(parsed)
					break
				}
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

func (w *Webhook) initializeProviders(r configv1alpha1.TerraformRepository) ([]gitprovider.Provider, error) {
	if r.Spec.Repository.SecretName == "" {
		log.Debugf("Tried to initialize default providers, but no webhook secret configured for repository %s/%s", r.Namespace, r.Name)
		return nil, nil
	}
	secret := &corev1.Secret{}
	err := w.Client.Get(context.Background(), types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.Repository.SecretName,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook secret for repository %s/%s: %w", r.Namespace, r.Spec.Repository.SecretName, err)
	}
	value, ok := secret.Data["webhookSecret"]
	if !ok {
		return nil, fmt.Errorf("webhook secret not found in secret %s/%s", r.Namespace, r.Spec.Repository.SecretName)
	}
	webhookSecret := string(value)

	availableProviders, err := gitprovider.ListAvailable(gitprovider.Config{WebhookSecret: webhookSecret}, []string{"webhook"})
	if err != nil {
		return nil, fmt.Errorf("failed to list available providers: %w", err)
	}

	providers := make([]gitprovider.Provider, 0)
	for _, availableProvider := range availableProviders {
		provider, err := gitprovider.NewWithName(gitprovider.Config{WebhookSecret: webhookSecret}, availableProvider)
		if err != nil {
			log.Errorf("failed to create provider %s: %s", availableProvider, err)
			continue
		}
		err = provider.InitWebhookHandler()
		if err != nil {
			log.Errorf("failed to initialize provider %s: %s", availableProvider, err)
			continue
		}
		providers = append(providers, provider)
	}

	for _, p := range providers {
		err := p.InitWebhookHandler()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize provider: %w", err)
		}
	}

	return providers, nil
}

func (w *Webhook) initializeDefaultProvider() error {
	if w.Providers["githubDefault"] != nil && w.Config.Server.Webhook.Github.Secret != "" {
		provider, err := gitprovider.NewWithName(gitprovider.Config{WebhookSecret: w.Config.Server.Webhook.Github.Secret}, "github")
		if err != nil {
			return fmt.Errorf("failed to create default provider: %w", err)
		}
		err = provider.InitWebhookHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize default provider: %w", err)
		}
		w.Providers["githubDefault"] = []gitprovider.Provider{provider}
		log.Info("initialized default GitHub webhook handler")
	}
	if w.Providers["gitlabDefault"] != nil && w.Config.Server.Webhook.Gitlab.Secret != "" {
		provider, err := gitprovider.NewWithName(gitprovider.Config{WebhookSecret: w.Config.Server.Webhook.Gitlab.Secret}, "gitlab")
		if err != nil {
			return fmt.Errorf("failed to create default provider: %w", err)
		}
		err = provider.InitWebhookHandler()
		if err != nil {
			return fmt.Errorf("failed to initialize default provider: %w", err)
		}
		w.Providers["gitlabDefault"] = []gitprovider.Provider{provider}
		log.Info("initialized default GitLab webhook handler")
	}
	return nil
}
