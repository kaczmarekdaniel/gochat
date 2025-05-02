package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

var clients = make(map[*websocket.Conn]bool)

func Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	clients[conn] = true

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Println("read: ", err)
			delete(clients, conn)

			break
		}

		log.Printf("recv: %s", message)

		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println("broadcast err:", err)
				client.Close()
				delete(clients, client)
			}
		}

	}
}
