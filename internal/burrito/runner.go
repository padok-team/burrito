package burrito

func (app *App) StartRunner() error {
	return app.Runner.Exec()
}
