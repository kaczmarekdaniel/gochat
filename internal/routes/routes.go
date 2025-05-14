package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/store"
)

func withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*") // Or specific origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusOK) // Must return 200 OK for preflight
			return                       // Important: don't call the next handler for OPTIONS
		}

		fmt.Println("Middleware: Before handling request")

		next(w, r)

		fmt.Println("Middleware: After handling request")
	}
}

func SetupRoutes(app *app.Application) {

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Or specific origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusOK) // Must return 200 OK for preflight
			return                       // Important: don't call the next handler for OPTIONS
		}

		user, err := app.UserHandler.HandleGetUser(w, r)

		if err != nil {
			fmt.Println(err)

			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusUnauthorized)

			errorResponse := map[string]string{
				"message": "Authentication failed",
				"success": "false",
			}

			json.NewEncoder(w).Encode(errorResponse)
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

	http.HandleFunc("/create-user", app.UserHandler.HandleCreateUser)

	// dev endpoints
	http.HandleFunc("/join-room", withMiddleware(app.RoomHandler.HandleJoinRoom))
	http.HandleFunc("/leave-room", app.RoomHandler.HandleLeaveRoom)
	http.HandleFunc("/user-rooms", withMiddleware(app.RoomHandler.HandleUserRooms))
	http.HandleFunc("/rooms", withMiddleware(app.RoomHandler.HandleRooms))
	http.HandleFunc("/init", withMiddleware(app.MessageHandler.HandleGetAllMesssages))

}
