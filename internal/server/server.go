package server

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/server/routes"
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
	http.HandleFunc("/layers", CORS(routes.GetAllLayers))

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

// Temporary fix for CORS
// https://stackoverflow.com/questions/64062803/how-to-enable-cors-in-go
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
