package burrito

func (app *App) StartDatastore() error {
	app.Datastore.Exec()
	return nil
}
