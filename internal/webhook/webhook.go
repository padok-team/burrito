package webhook

import (
	"github.com/padok-team/burrito/internal/burrito/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler interface {
	Handle()
}

type Webhook struct {
	config *config.Config
	client client.Client
}

func New(c *config.Config) *Webhook {
	return &Webhook{
		config: c,
	}
}

func (w *Webhook) Exec() {
	return
}
