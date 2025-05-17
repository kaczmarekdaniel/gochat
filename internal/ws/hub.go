package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kaczmarekdaniel/gochat/internal/store"
)

type Hub struct {
	clients      map[*Client]bool
	broadcast    chan *store.Message
	register     chan *Client
	unregister   chan *Client
	roomStore    store.RoomStore    // Add this
	messageStore store.MessageStore // Add this
}

type RoomAction struct {
	client *Client
	room   string
}

func newHub(roomStore store.RoomStore, messageStore store.MessageStore) *Hub {
	return &Hub{
		broadcast:    make(chan *store.Message),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		clients:      make(map[*Client]bool),
		roomStore:    roomStore,
		messageStore: messageStore,
	}
}
func (h *Hub) run() {

	fmt.Println("Hub run started")
	fmt.Printf("roomStore is nil: %v\n", h.roomStore == nil)
	fmt.Printf("messageStore is nil: %v\n", h.messageStore == nil)

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			// Send initial room list to client
			go func() {
				rooms, err := h.roomStore.GetUserRooms(context.Background(), client.userID)
				if err != nil {
					log.Printf("Error getting user rooms: %v", err)
					return
				}

				// Create a new structure to hold rooms with their messages
				type RoomWithMessages struct {
					*store.Room                  // Embed the original Room
					Messages    []*store.Message `json:"messages"`
				}

				roomsWithMessages := make([]RoomWithMessages, 0, len(rooms))

				// Attach messages to each room
				for _, room := range rooms {
					// Get messages for this specific room
					roomMessages, err := h.messageStore.GetMessages(room.ID) // Assuming room has an ID field
					if err != nil {
						log.Printf("Error retrieving messages from room %s: %v", room.ID, err)
						// Continue with other rooms even if we fail to get messages for this one
					}

					// Create a room with messages
					roomWithMessages := RoomWithMessages{
						Room:     room,
						Messages: roomMessages,
					}

					roomsWithMessages = append(roomsWithMessages, roomWithMessages)
				}

				roomsJSON, _ := json.Marshal(roomsWithMessages)
				client.send <- &store.Message{
					Type:    "room_list",
					Content: string(roomsJSON),
					Sender:  "system",
					Time:    time.Now(),
				}
			}()

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			// Save message to database
			if _, err := h.messageStore.CreateMessage(message); err != nil {
				log.Printf("Error saving message: %v", err)
				continue
			}

			// Find clients in the room and send them the message
			h.distributeMessage(message)
		}
	}
}

// Distribute message to clients in the room
func (h *Hub) distributeMessage(message *store.Message) {
	ctx := context.Background()

	// Get all users in the room
	roomUsers, err := h.roomStore.GetRoomUsers(ctx, message.Room)
	if err != nil {
		log.Printf("Error getting room users: %v", err)
		return
	}

	// Create a map for faster lookups
	roomUserMap := make(map[string]bool)
	for _, userID := range roomUsers {
		roomUserMap[userID] = true
	}

	// Send to all clients in the room
	for client := range h.clients {
		if roomUserMap[client.userID] {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
