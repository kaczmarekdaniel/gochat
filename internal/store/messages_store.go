package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Message struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`    // e.g., "chat", "notification", "error"
	Content string    `json:"content"` // The actual message content
	Sender  string    `json:"sender"`  // Who sent the message
	Time    time.Time `json:"time"`    // When the message was sent
}

type PostgresMessagesStore struct {
	db *sql.DB
}

func NewPostgresMessageStore(db *sql.DB) *PostgresMessagesStore {
	return &PostgresMessagesStore{db: db}
}

type MessageStore interface {
	CreateMessage(*Message) (*Message, error)
}

func (pg *PostgresMessagesStore) CreateMessage(message *Message) (*Message, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		fmt.Println("here")

		return nil, err
	}
	defer tx.Rollback()

	query :=
		`INSERT INTO Messages (type, content, sender, time)
  VALUES ($1, $2, $3, $4)
  RETURNING id
  `
	err = tx.QueryRow(query, message.Type, message.Content, message.Sender, message.Time).Scan(&message.ID)
	if err != nil {
		fmt.Println("here 1")
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	fmt.Println("message successfully created", message)

	return message, nil
}
