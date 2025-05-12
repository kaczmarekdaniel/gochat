package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kaczmarekdaniel/gochat/internal/api"
	"github.com/kaczmarekdaniel/gochat/internal/store"
	"github.com/kaczmarekdaniel/gochat/migrations"
)

type Application struct {
	Logger         *log.Logger
	MessageHandler *api.MessageHandler
	DB             *sql.DB
	MessageStore   store.MessageStore
	RoomStore      store.RoomStore
	RoomHandler    *api.RoomHandler
	UserHandler    *api.UserHandler
	SessionHandler *api.SessionHandler
}

type RoomHandler struct {
	roomStore store.RoomStore
}

func NewRoomHandler(roomStore store.RoomStore) *RoomHandler {
	return &RoomHandler{
		roomStore: roomStore,
	}
}

// Example handler method for rooms
func (rh *RoomHandler) HandleRooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Return list of rooms
		rooms, err := rh.roomStore.GetRooms(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rooms)

	case "POST":
		// Create a new room
		var roomReq struct {
			Name string `json:"name"`
		}

		err := json.NewDecoder(r.Body).Decode(&roomReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		room, err := rh.roomStore.CreateRoom(r.Context(), roomReq.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(room)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
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

	messageStore := store.NewPostgresMessageStore(pgDB)
	roomStore := store.NewPostgresRoomStore(pgDB)
	userStore := store.NewPostgresUserStore(pgDB)
	sessionStore := store.NewPostgresSessionStore(pgDB)

	// Create handlers
	roomHandler := api.NewRoomHandler(roomStore)
	messageHandler := api.NewMessageHandler(messageStore)
	userHandler := api.NewUserHandler(userStore)
	sessionHandler := api.NewSessionHandler(sessionStore)

	app := &Application{
		MessageStore:   messageStore,
		MessageHandler: messageHandler,
		UserHandler:    userHandler,
		SessionHandler: sessionHandler,
		RoomStore:      roomStore,
		Logger:         logger,
		RoomHandler:    roomHandler,
		DB:             pgDB,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "status is available")
}
