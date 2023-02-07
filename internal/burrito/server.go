package burrito

func (app *App) StartServer() {
	app.Server.Exec()
}
