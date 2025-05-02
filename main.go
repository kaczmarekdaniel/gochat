package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kaczmarekdaniel/gochat/internal/ws"
)

func main() {

	fmt.Println("Server running on port :8080")

	http.HandleFunc("/hello", ws.Hello)
	http.HandleFunc("/ws", ws.Handler)

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
