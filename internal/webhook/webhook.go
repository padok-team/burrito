package webhook

import (
	"github.com/go-playground/webhooks/v6/github"
	"github.com/padok-team/burrito/internal/burrito/config"
)

type Handler interface {
	Handle()
}

type WebhookHandler struct {
	config *config.Config
	github *github.Webhook
}

func New(c *config.Config) *WebhookHandler {
	return &WebhookHandler{
		config: c,
	}
}

func (w *WebhookHandler) Init() error {
	githubWebhook, err := github.New(github.Options.Secret(w.config.Webhook.Github.Secret))
	if err != nil {
		return err
	}
	w.github = githubWebhook
	return nil
}

func (w *WebhookHandler) Handle() {
	return
}
