package store

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
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

func (pg *PostgresUserStore) GetUser(id string) (*User, error) {
	tx, err := pg.db.Begin()

	if err != nil {
		fmt.Println("here")

		return nil, err
	}
	defer tx.Rollback()

	user := &User{}

	query := `
    SELECT id, title, description, duration_minutes, calories_burned
    FROM users
    WHERE id = $1
  `

	err = pg.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.ProfilePicture, &user.CreatedAt)
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
		fmt.Println("here")

		return nil, err
	}
	defer tx.Rollback()

	query :=
		`INSERT INTO Users (type, room, content, sender, time)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING id
  `
	err = tx.QueryRow(query, user.ID, user.Username, user.ProfilePicture, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	fmt.Println("user successfully created", user)

	return user, nil
}
