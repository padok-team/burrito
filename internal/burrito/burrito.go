package burrito

import (
	"io"
	"os"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers"
	"github.com/padok-team/burrito/internal/runner"
	"github.com/padok-team/burrito/internal/server"
	"github.com/padok-team/burrito/internal/webhook"
)

type App struct {
	Config *config.Config

	Runner      Runner
	Controllers Controllers
	Webhook     Webhook
	Server      Server

	Out io.Writer
	Err io.Writer
}

type Webhook interface {
	Exec()
}

type Server interface {
	Exec()
}

type Runner interface {
	Exec()
}

type Controllers interface {
	Exec()
}

func New() (*App, error) {
	c := &config.Config{}
	app := &App{
		Config:      c,
		Runner:      runner.New(c),
		Controllers: controllers.New(c),
		Webhook:     webhook.New(c),
		Server:      server.New(c),
		Out:         os.Stdout,
		Err:         os.Stderr,
	}
	return app, nil
}
