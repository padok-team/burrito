package webhook

import (
	"fmt"
	"html"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/webhook/event"
	"github.com/padok-team/burrito/internal/webhook/github"
	"github.com/padok-team/burrito/internal/webhook/gitlab"

	"k8s.io/apimachinery/pkg/runtime"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
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
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}
	w.Client = cl
	providers := []Provider{}
	for _, p := range []Provider{&github.Github{}, &gitlab.Gitlab{}} {
		err = p.Init(w.Config)
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

func (w *Webhook) GetHttpHandler() func(http.ResponseWriter, *http.Request) {
	log.Infof("webhook event received...")
	return func(writer http.ResponseWriter, r *http.Request) {
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
			http.Error(writer, fmt.Sprintf("webhook processing failed: %s", html.EscapeString(err.Error())), status)
			return
		}
		if event == nil {
			log.Infof("ignoring unknown webhook event")
			http.Error(writer, "Unknown webhook event", http.StatusBadRequest)
		}

		err = event.Handle(w.Client)
		if err != nil {
			log.Errorf("webhook processing worked but errored during event handling: %s", err)
		}
	}
}
