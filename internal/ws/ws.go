package ws

import (
	"net/http"
)

func Serve() {
	hub := newHub()
	go hub.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})
}
