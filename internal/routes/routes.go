package routes

import (
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/ws"
)

func SetupRoutes(app *app.Application) {

	http.HandleFunc("/hello", app.WS.Hello)
	http.HandleFunc("/ws", ws.Handler)

}
