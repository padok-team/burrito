package burrito

import (
	"io"
	"os"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers"
	"github.com/padok-team/burrito/internal/runner"
	"github.com/padok-team/burrito/internal/server"
)

type App struct {
	Config *config.Config

	Runner      Runner
	Controllers Controllers
	Server      Server

	Out io.Writer
	Err io.Writer
}

type Server interface {
	Exec()
}

type Runner interface {
	Exec() error
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
		Server:      server.New(c),
		Out:         os.Stdout,
		Err:         os.Stderr,
	}
	return app, nil
}
