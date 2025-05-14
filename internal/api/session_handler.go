package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/store"
)

type SessionHandler struct {
	sessionStore store.SessionStore
}

func NewSessionHandler(sessionStore store.SessionStore) *SessionHandler {
	return &SessionHandler{
		sessionStore: sessionStore,
	}
}

func (wh *SessionHandler) HandleGetSessionByToken(w http.ResponseWriter, r *http.Request) {

	var session store.Session
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionID := session.Token

	sessionData, err := wh.sessionStore.GetSessionByToken(sessionID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to retrieve the session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessionData)
}

func (wh *SessionHandler) HandleCreateSession(sessionRaw *store.Session) (*store.Session, error) {

	sessionCreated, err := wh.sessionStore.CreateSession(sessionRaw)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return sessionCreated, nil
}
