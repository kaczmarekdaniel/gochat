package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kaczmarekdaniel/gochat/internal/store"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all connections

}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan *store.Message
	userID string // User's identifier
} // validateMessage checks if a message is valid

func validateMessage(message *store.Message) (bool, string) {
	// Check required fields
	if message.Type == "" {
		return false, "message type is required"
	}

	if message.Content == "" {
		return false, "message content is required"
	}

	if message.Sender == "" {
		return false, "message sender is required"
	}

	validTypes := map[string]bool{
		"chat":         true,
		"notification": true,
		"system":       true,
		"error":        true,
	}

	if !validTypes[message.Type] {
		return false, fmt.Sprintf("invalid message type: %s", message.Type)
	}

	if len(message.Content) > 1000 {
		return false, "message content exceeds maximum length of 1000 characters"
	}

	if len(message.Sender) > 50 {
		return false, "sender name exceeds maximum length of 50 characters"
	}

	// Add other validations as needed:
	// - Check for profanity in content
	// - Validate message format for specific types
	// - Rate limiting (number of messages per minute)

	return true, ""
}

func sanitizeMessage(message *store.Message) {
	// HTML escape the content to prevent XSS attacks
	message.Content = html.EscapeString(message.Content)

	// Trim whitespace
	message.Content = strings.TrimSpace(message.Content)

	message.Sender = html.EscapeString(message.Sender)
	message.Sender = strings.TrimSpace(message.Sender)
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Set connection parameters
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		var message store.Message
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Process message based on type
		ctx := context.Background()
		switch message.Type {
		case "join_room":
			// Add user to room in database
			if err := c.hub.roomStore.JoinRoom(ctx, c.userID, message.Room); err != nil {
				// Send error message to client
				c.send <- &store.Message{
					Type:    "error",
					Content: fmt.Sprintf("Failed to join room: %v", err),
					Sender:  "system",
					Time:    time.Now(),
				}
				continue
			}

			// Send confirmation
			c.send <- &store.Message{
				Type:    "system",
				Content: fmt.Sprintf("Joined room: %s", message.Room),
				Sender:  "system",
				Room:    message.Room,
				Time:    time.Now(),
			}

		case "leave_room":
			// Remove user from room in database
			if err := c.hub.roomStore.LeaveRoom(ctx, c.userID, message.Room); err != nil {
				c.send <- &store.Message{
					Type:    "error",
					Content: fmt.Sprintf("Failed to leave room: %v", err),
					Sender:  "system",
					Time:    time.Now(),
				}
				continue
			}

			// Send confirmation
			c.send <- &store.Message{
				Type:    "system",
				Content: fmt.Sprintf("Left room: %s", message.Room),
				Sender:  "system",
				Room:    message.Room,
				Time:    time.Now(),
			}

		case "chat":
			// Check if user is in this room
			isInRoom, err := c.hub.roomStore.IsUserInRoom(ctx, c.userID, message.Room)
			if err != nil || !isInRoom {
				c.send <- &store.Message{
					Type:    "error",
					Content: "You are not a member of this room",
					Sender:  "system",
					Room:    message.Room,
					Time:    time.Now(),
				}
				continue
			}

			// Set sender ID and broadcast message
			message.Sender = c.userID
			c.hub.broadcast <- &message
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Println("error marshalling message:", err)
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(jsonMessage)

			n := len(c.send)
			for range n {
				w.Write(newline)
				nextMsg := <-c.send
				jsonNext, err := json.Marshal(nextMsg)
				if err != nil {
					log.Println("error marshalling queued message:", err)
					continue
				}
				w.Write(jsonNext)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

//

func createClient(hub *Hub, w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user id is mandatory", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan *store.Message, 256),
		userID: userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
