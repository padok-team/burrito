package burrito

import (
	"io"
	"os"

	"github.com/padok-team/burrito/burrito/config"
	"github.com/padok-team/burrito/controllers"
	"github.com/padok-team/burrito/runner"
)

type App struct {
	Config *config.Config

	Runner      Runner
	Controllers Controllers

	Out io.Writer
	Err io.Writer
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
		Out:         os.Stdout,
		Err:         os.Stderr,
	}
	return app, nil
}
