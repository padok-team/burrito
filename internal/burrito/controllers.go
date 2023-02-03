package burrito

import (
	"errors"
	"time"
)

func (app *App) StartController() error {
	_, err := time.ParseDuration(app.Config.Controller.Timers.DriftDetection)
	if err != nil {
		return errors.New("could not parse drift detection timer")
	}
	_, err = time.ParseDuration(app.Config.Controller.Timers.OnError)
	if err != nil {
		return errors.New("could not parse on error timer")
	}
	_, err = time.ParseDuration(app.Config.Controller.Timers.WaitAction)
	if err != nil {
		return errors.New("could not parse wait action timer")
	}
	app.Controllers.Exec()
	return nil
}
