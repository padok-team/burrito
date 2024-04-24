package burrito

func (app *App) StartDatastore() error {
	app.Controllers.Exec()
	return nil
}
