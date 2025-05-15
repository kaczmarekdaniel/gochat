package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kaczmarekdaniel/gochat/internal/store"
)

// AuthHandler manages authentication-related HTTP handlers
type AuthHandler struct {
	UserStore    store.UserStore
	SessionStore store.SessionStore
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userStore store.UserStore, sessionStore store.SessionStore) *AuthHandler {
	return &AuthHandler{
		UserStore:    userStore,
		SessionStore: sessionStore,
	}
}

// HandleLogin authenticates a user and creates a session
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse login request
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid request format",
			"success": "false",
		})
		return
	}

	// Get user by username
	user, err := h.UserStore.GetUser(loginRequest.Username)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Authentication failed",
			"success": "false",
		})
		return
	}

	// Verify password
	// Make sure to import "golang.org/x/crypto/bcrypt"
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	// if err != nil {
	//    w.Header().Set("Content-Type", "application/json")
	//    w.WriteHeader(http.StatusUnauthorized)
	//    json.NewEncoder(w).Encode(map[string]string{
	//        "message": "Authentication failed",
	//        "success": "false",
	//    })
	//    return
	// }

	// Create session
	sessionToken := uuid.New().String()
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

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

	createdSession, err := h.SessionStore.CreateSession(session)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set cookie
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
		"session": map[string]interface{}{
			"id":         createdSession.SessionID,
			"expires_at": createdSession.ExpiresAt.Format(time.RFC3339),
		},
	})
}
