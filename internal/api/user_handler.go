package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kaczmarekdaniel/gochat/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userStore store.UserStore
}

func NewUserHandler(userStore store.UserStore) *UserHandler {
	return &UserHandler{
		userStore: userStore,
	}
}

func (uh *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) (*store.User, error) {

	var user store.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	fmt.Println(user)

	userData, err := uh.userStore.GetUser(user.Username)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if userData == nil {
		return nil, fmt.Errorf("user not found")
	}

	return userData, nil
}

func (uh *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var userRaw store.User
	err := json.NewDecoder(r.Body).Decode(&userRaw)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username
	if userRaw.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate password
	if userRaw.Password == "" {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRaw.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userRaw.Password = string(hashedPassword)

	// Create the user in the database
	userCreated, err := uh.userStore.CreateUser(&userRaw)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}

		fmt.Println(err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return success response with the created user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created is more appropriate than 200 OK

	// Create a sanitized response (don't include password)
	// Adjust this based on your actual User struct fields
	userResponse := map[string]interface{}{
		"id":         userCreated.ID,
		"username":   userCreated.Username,
		"created_at": userCreated.CreatedAt,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user":    userResponse,
	})
}
