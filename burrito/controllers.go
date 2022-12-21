package burrito

func (app *App) StartController() error {
	app.Controllers.Exec()
	return nil
}
