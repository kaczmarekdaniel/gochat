package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/store"
)

func Start(app *app.Application) {
	hub := newHub(app.RoomStore, app.MessageStore)

	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		createClient(hub, w, r)
	})

	// Login handler that uses HandleGetUser for authentication
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		user, err := app.UserHandler.HandleGetUser(w, r)
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		sessionToken := uuid.New().String()
		sessionID := uuid.New().String()
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		// Get client IP address
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		session := &store.Session{
			SessionID:    sessionID,
			UserID:       user.ID,
			Token:        sessionToken,
			IPAddress:    ip,
			IsActive:     true,
			CreatedAt:    now,
			LastActivity: now,
			ExpiresAt:    expiresAt,
		}

		createdSession, err := app.SessionHandler.HandleCreateSession(session)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"user":    user,
			"session": map[string]any{
				"id":         createdSession.SessionID,
				"expires_at": createdSession.ExpiresAt.Format(time.RFC3339),
			},
		})
	})
	http.HandleFunc("/init", app.MessageHandler.HandleGetAllMesssages)

	http.HandleFunc("/create-user", app.UserHandler.HandleCreateUser)
	http.HandleFunc("/rooms", app.RoomHandler.HandleRooms)
	http.HandleFunc("/user-rooms", app.RoomHandler.HandleUserRooms)
	http.HandleFunc("/join-room", app.RoomHandler.HandleJoinRoom)
	http.HandleFunc("/leave-room", app.RoomHandler.HandleLeaveRoom)

}
