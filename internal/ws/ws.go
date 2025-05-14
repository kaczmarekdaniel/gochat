package ws

import (
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/routes"
)

func Start(app *app.Application) {
	hub := newHub(app.RoomStore, app.MessageStore)

	go hub.run()

	routes.SetupRoutes(app)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})

}
