package ws

import (
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/app"
)

func Start(app *app.Application) {
	hub := newHub()
	go hub.run(app)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})
}
