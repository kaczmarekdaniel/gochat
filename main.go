package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/ws"
)

func main() {

	app, err := app.NewApplication()

	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ws.Start(app)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("fatal error")
	}

}
