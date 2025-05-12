package store

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

type UserStore interface {
	CreateUser(*User) (*User, error)
	GetUser(id string) (*User, error)
}

func (pg *PostgresUserStore) GetUser(username string) (*User, error) {
	// Check if id is empty
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	user := &User{}
	query := `
        SELECT id, username, profile_picture, created_at
        FROM users
        WHERE username = $1
    `
	err := pg.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.ProfilePicture, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PostgresUserStore) CreateUser(user *User) (*User, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Set creation time if not already set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO users (username, password, profile_picture, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err = tx.QueryRow(
		query,
		user.Username,
		user.Password,
		user.ProfilePicture,
		user.CreatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	fmt.Println("User successfully created:", user.Username)
	return user, nil
}
