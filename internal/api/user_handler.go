package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/store"
)

type UserHandler struct {
	userStore store.UserStore
}

func NewUserHandler(userStore store.UserStore) *UserHandler {
	return &UserHandler{
		userStore: userStore,
	}
}

func (wh *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {

	var user store.User
	err := json.NewDecoder(r.Body).Decode(&r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := user.ID

	messages, err := wh.userStore.GetUser(userID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to retrieve the messages", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

func (wh *UserHandler) HandleCreateUser(userRaw *store.User) (*store.User, error) {
	if userRaw.Username == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	userCreated, err := wh.userStore.CreateUser(userRaw)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return userCreated, nil
}

func (wh *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {

	// 1. Find the user by id
	// 2. Compare passwords
	// 3. if it's ok, generate the session

}
