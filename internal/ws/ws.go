package ws

import (
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/app"
)

func Start(app *app.Application) {
	hub := newHub(app.RoomStore, app.MessageStore)

	// Start the hub in a goroutine
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})

	http.HandleFunc("/init", app.MessageHandler.HandleGetAllMesssages)

	http.HandleFunc("/rooms", app.RoomHandler.HandleRooms)
	http.HandleFunc("/user-rooms", app.RoomHandler.HandleUserRooms)
	http.HandleFunc("/join-room", app.RoomHandler.HandleJoinRoom)
	http.HandleFunc("/leave-room", app.RoomHandler.HandleLeaveRoom)

}
