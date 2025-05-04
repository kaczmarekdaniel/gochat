package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/store"
)

type MessageHandler struct {
	messageStore store.MessageStore
}

func NewMessageHandler(messageStore store.MessageStore) *MessageHandler {
	return &MessageHandler{
		messageStore: messageStore,
	}
}

func (wh *MessageHandler) HandleGetAllMesssages(w http.ResponseWriter, r *http.Request) {

	messages, err := wh.messageStore.GetAllMessages()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to retrieve the messages", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

func (wh *MessageHandler) HandleCreateMessage(messageRaw *store.Message) (*store.Message, error) {
	if messageRaw.Content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// Now you can pass the message to your store
	createdMessage, err := wh.messageStore.CreateMessage(messageRaw)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return createdMessage, nil
}
