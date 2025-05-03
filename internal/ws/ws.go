package ws

import (
	"net/http"
	"time"
)

type Message struct {
	Type    string    `json:"type"`    // e.g., "chat", "notification", "error"
	Content string    `json:"content"` // The actual message content
	Sender  string    `json:"sender"`  // Who sent the message
	Time    time.Time `json:"time"`    // When the message was sent
}

func Start() {
	hub := newHub()
	go hub.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})
}
