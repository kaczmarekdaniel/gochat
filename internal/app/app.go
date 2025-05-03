package app

import "github.com/kaczmarekdaniel/gochat/internal/ws"

type Application struct {
	client *ws.Client
}

func NewApplication() (*Application, error) {

	app := &Application{}

	return app, nil
}
