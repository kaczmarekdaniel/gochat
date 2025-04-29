package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

var clients = make(map[*websocket.Conn]bool)

func handler(w http.ResponseWriter, r *http.Request) {
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

func main() {

	fmt.Println("Server running on port :8080")

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ws", handler)

	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("fatal error")
	}

}
