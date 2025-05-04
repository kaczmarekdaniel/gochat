package api

import (
	"fmt"

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
