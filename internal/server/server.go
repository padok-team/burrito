package server

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/webhook"
)

type Server struct {
	config  *config.Config
	Webhook *webhook.Webhook
}

func New(c *config.Config) *Server {
	webhook := webhook.New(c)
	err := webhook.Init()
	if err != nil {
		log.Printf("error initializing webhook: %s", err)
	}
	return &Server{
		config:  c,
		Webhook: webhook,
	}
}

func (s *Server) Exec() {
	log.Infof("starting burrito server...")
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/webhook", s.Webhook.GetHttpHandler())

	err := http.ListenAndServe(s.config.Server.Addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Errorf("burrito server is closed: %s", err)
	}
	if err != nil {
		log.Fatalf("error starting burrito server, exiting: %s", err)
	}
	log.Infof("burrito server started on addr %s", s.config.Server.Addr)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	// The HTTP server is always healthy.
	// TODO: check it can get terraformlayers and/or repositories
	log.Infof("request received on /healthz")
}
