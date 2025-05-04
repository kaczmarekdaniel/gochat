package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Message struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"` // e.g., "chat", "notification", "error"
	Room    string    `json:"room"`
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
	GetAllMessages() ([]*Message, error)
}

func (pg *PostgresMessagesStore) GetAllMessages() ([]*Message, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		fmt.Println("here")

		return nil, err
	}
	defer tx.Rollback()

	query := `
        SELECT id, type, room, content, sender, time 
        FROM messages
        ORDER BY time DESC
    `

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		err = rows.Scan(
			&message.ID,
			&message.Type,
			&message.Room,
			&message.Content,
			&message.Sender,
			&message.Time,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (pg *PostgresMessagesStore) CreateMessage(message *Message) (*Message, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		fmt.Println("here")

		return nil, err
	}
	defer tx.Rollback()

	query :=
		`INSERT INTO Messages (type, room, content, sender, time)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING id
  `
	err = tx.QueryRow(query, message.Type, message.Room, message.Content, message.Sender, message.Time).Scan(&message.ID)
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
