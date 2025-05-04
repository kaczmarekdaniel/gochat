package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kaczmarekdaniel/gochat/internal/api"
	"github.com/kaczmarekdaniel/gochat/internal/migrations"
	"github.com/kaczmarekdaniel/gochat/internal/store"
)

type Application struct {
	Logger         *log.Logger
	MessageHandler *api.MessageHandler
	DB             *sql.DB
}

func NewApplication() (*Application, error) {

	// our stores will go here
	pgDB, err := store.Open()
	if err != nil {

		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// out handlers will go here
	messageStore := store.NewPostgresMessageStore(pgDB)
	messageHandler := api.NewMessageHandler(messageStore)

	app := &Application{
		Logger:         logger,
		MessageHandler: messageHandler,
		DB:             pgDB,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) { // w - communicate back to the caller. r- request, this is what we get
	fmt.Fprint(w, "status is available") // fprint is specifically used to send data back to the caller/client
}
