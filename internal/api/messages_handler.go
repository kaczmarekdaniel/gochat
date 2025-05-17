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

func (wh *MessageHandler) HandleGetMesssages(w http.ResponseWriter, r *http.Request) {

	messages, err := wh.messageStore.GetMessages("123") // TODO: FIX IT
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

	createdMessage, err := wh.messageStore.CreateMessage(messageRaw)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return createdMessage, nil
}
