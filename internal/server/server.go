package server

import (
	"errors"
	"log"
	"net/http"
	"os"

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
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/webhook", s.Webhook.GetHttpHandler())

	err := http.ListenAndServe(s.config.Server.Addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server is closed")
	}
	if err != nil {
		log.Printf("error starting server, exiting: %s", err)
		os.Exit(1)
	}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	// The HTTP server is always healthy.
	// TODO: check it can get terraformlayers and/or repositories
}
