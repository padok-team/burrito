package burrito

import (
	"io"
	"os"

	"github.com/padok-team/burrito/burrito/config"
	"github.com/padok-team/burrito/controllers"
	"github.com/padok-team/burrito/runner"
	"github.com/padok-team/burrito/webhook"
)

type App struct {
	Config *config.Config

	Runner      Runner
	Controllers Controllers
	Webhook     Webhook

	Out io.Writer
	Err io.Writer
}

type Webhook interface {
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
		Out:         os.Stdout,
		Err:         os.Stderr,
	}
	return app, nil
}
