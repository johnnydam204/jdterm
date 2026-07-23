package app

import "jdterm/internal/webserver"

type App struct {
	server *webserver.Server
}

func New() *App {

	return &App{

		server: webserver.New(":8080"),
	}
}

func (a *App) Run() error {

	return a.server.Start()

}
