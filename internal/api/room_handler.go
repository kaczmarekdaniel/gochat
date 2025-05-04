package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kaczmarekdaniel/gochat/internal/store"
	"net/http"
)

type RoomHandler struct {
	roomStore store.RoomStore
}

func NewRoomHandler(roomStore store.RoomStore) *RoomHandler {
	return &RoomHandler{
		roomStore: roomStore,
	}
}

// HandleRooms handles GET and POST requests for rooms
func (rh *RoomHandler) HandleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case http.MethodGet:
		rh.handleGetRooms(w)
	case http.MethodPost:
		rh.handleCreateRoom(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetRooms handles GET requests to retrieve all rooms
func (rh *RoomHandler) handleGetRooms(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := context.Background()
	rooms, err := rh.roomStore.GetRooms(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to retrieve rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

// handleCreateRoom handles POST requests to create a new room
func (rh *RoomHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var roomRequest struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&roomRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if roomRequest.Name == "" {
		http.Error(w, "Room name cannot be empty", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	room, err := rh.roomStore.CreateRoom(ctx, roomRequest.Name)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

// HandleUserRooms gets rooms for a specific user
func (rh *RoomHandler) HandleUserRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	rooms, err := rh.roomStore.GetUserRooms(ctx, userID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to retrieve user rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

// HandleJoinRoom handles requests to join a room
func (rh *RoomHandler) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var joinRequest struct {
		UserID string `json:"user_id"`
		RoomID string `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&joinRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println(joinRequest)
	if joinRequest.UserID == "" || joinRequest.RoomID == "" {
		http.Error(w, "User ID and Room ID are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rh.roomStore.JoinRoom(ctx, joinRequest.UserID, joinRequest.RoomID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully joined room"})
}

// HandleLeaveRoom handles requests to leave a room
func (rh *RoomHandler) HandleLeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var leaveRequest struct {
		UserID string `json:"user_id"`
		RoomID string `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&leaveRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Println(leaveRequest)
	if leaveRequest.UserID == "" || leaveRequest.RoomID == "" {
		http.Error(w, "User ID and Room ID are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rh.roomStore.LeaveRoom(ctx, leaveRequest.UserID, leaveRequest.RoomID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to leave room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully left room"})
}
