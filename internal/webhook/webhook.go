package webhook

import (
	"context"
	"fmt"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/webhook/event"
	"github.com/padok-team/burrito/internal/webhook/github"
	"github.com/padok-team/burrito/internal/webhook/gitlab"
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
	Providers map[string][]Provider
}

func New(c *config.Config) *Webhook {
	return &Webhook{
		Config:    c,
		Providers: make(map[string][]Provider),
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
				parsed, ok := p.ParseFromProvider(r)
				if ok {
					event, err = p.GetEvent(parsed)
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

func (w *Webhook) initializeProviders(r configv1alpha1.TerraformRepository) ([]Provider, error) {
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

	providers := []Provider{
		&github.Github{Secret: webhookSecret},
		&gitlab.Gitlab{Secret: webhookSecret},
	}

	for _, p := range providers {
		err := p.Init()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize provider: %w", err)
		}
	}

	return providers, nil
}

func (w *Webhook) initializeDefaultProvider() error {
	if w.Providers["githubDefault"] != nil && w.Config.Server.Webhook.Github.Secret != "" {
		provider := &github.Github{Secret: w.Config.Server.Webhook.Github.Secret}
		err := provider.Init()
		if err != nil {
			return fmt.Errorf("failed to initialize default provider: %w", err)
		}
		w.Providers["githubDefault"] = []Provider{provider}
		log.Info("initialized default GitHub webhook handler")
	}
	if w.Providers["gitlabDefault"] != nil && w.Config.Server.Webhook.Gitlab.Secret != "" {
		provider := &gitlab.Gitlab{Secret: w.Config.Server.Webhook.Gitlab.Secret}
		err := provider.Init()
		if err != nil {
			return fmt.Errorf("failed to initialize default provider: %w", err)
		}
		w.Providers["gitlabDefault"] = []Provider{provider}
		log.Info("initialized default Gitlab webhook handler")
	}
	return nil
}
