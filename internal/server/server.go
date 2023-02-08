package server

import (
	"errors"
	"fmt"
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
	webhook.Init()
	return &Server{
		config:  c,
		Webhook: webhook,
	}
}

func (s *Server) Exec() {
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/webhook", s.Webhook.GetHttpHandler())

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.config.Server.Port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server is closed")
	}
	if err != nil {
		log.Println("error starting server, exiting: %s", err)
		os.Exit(1)
	}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	// The HTTP server is always healthy.
	// TODO: check it can get terraformlayers and/or repositories
}
