package app

import "github.com/kaczmarekdaniel/gochat/internal/ws"

type Application struct {
	WS *ws.WSHandler
}

func NewApplication() (*Application, error) {

	ws_handler := ws.NewWsHandler()

	app := &Application{
		WS: ws_handler,
	}

	return app, nil
}
