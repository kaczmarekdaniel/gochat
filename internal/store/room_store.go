package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Room represents a chat room
type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type RoomStore interface {
	// Get all available rooms
	GetRooms(ctx context.Context) ([]*Room, error)

	// Get rooms that a specific user is in
	GetUserRooms(ctx context.Context, userID string) ([]*Room, error)

	// Get users in a specific room
	GetRoomUsers(ctx context.Context, roomID string) ([]string, error)

	// Create a new room
	CreateRoom(ctx context.Context, name string) (*Room, error)

	// Add a user to a room
	JoinRoom(ctx context.Context, userID, roomID string) error

	// Remove a user from a room
	LeaveRoom(ctx context.Context, userID, roomID string) error

	// Check if a user is in a room
	IsUserInRoom(ctx context.Context, userID, roomID string) (bool, error)
}

type PostgresRoomStore struct {
	db *sql.DB
}

func NewPostgresRoomStore(db *sql.DB) *PostgresRoomStore {
	return &PostgresRoomStore{db: db}
}

// GetRooms returns all available rooms
func (s *PostgresRoomStore) GetRooms(ctx context.Context) ([]*Room, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
        SELECT id, name, created_at
        FROM rooms
        ORDER BY name
    `

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		room := &Room{}
		err = rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetUserRooms returns all rooms a specific user is in
func (s *PostgresRoomStore) GetUserRooms(ctx context.Context, userID string) ([]*Room, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
        SELECT r.id, r.name, r.created_at
        FROM rooms r
        JOIN room_memberships rm ON r.id = rm.room_id
        WHERE rm.user_id = $1
        ORDER BY r.name
    `

	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		room := &Room{}
		err = rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetRoomUsers returns all user IDs in a specific room
func (s *PostgresRoomStore) GetRoomUsers(ctx context.Context, roomID string) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
        SELECT user_id
        FROM room_memberships
        WHERE room_id = $1
    `

	rows, err := tx.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		err = rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

// CreateRoom creates a new chat room
func (s *PostgresRoomStore) CreateRoom(ctx context.Context, name string) (*Room, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO rooms (name)
        VALUES ($1)
        RETURNING id, name, created_at
    `

	row := tx.QueryRowContext(ctx, query, name)

	room := &Room{}
	err = row.Scan(
		&room.ID,
		&room.Name,
		&room.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	fmt.Println("Room successfully created:", room.Name, "with ID:", room.ID)
	return room, nil
}

// JoinRoom adds a user to a room
func (s *PostgresRoomStore) JoinRoom(ctx context.Context, userID, roomID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO room_memberships (user_id, room_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, room_id) DO NOTHING
    `

	_, err = tx.ExecContext(ctx, query, userID, roomID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Println("User", userID, "joined room", roomID)
	return nil
}

// LeaveRoom removes a user from a room
func (s *PostgresRoomStore) LeaveRoom(ctx context.Context, userID, roomID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        DELETE FROM room_memberships
        WHERE user_id = $1 AND room_id = $2
    `

	_, err = tx.ExecContext(ctx, query, userID, roomID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Println("User", userID, "left room", roomID)
	return nil
}

// IsUserInRoom checks if a user is in a room
func (s *PostgresRoomStore) IsUserInRoom(ctx context.Context, userID, roomID string) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query := `
        SELECT EXISTS(
            SELECT 1 FROM room_memberships
            WHERE user_id = $1 AND room_id = $2
        )
    `

	var exists bool
	err = tx.QueryRowContext(ctx, query, userID, roomID).Scan(&exists)
	if err != nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return exists, nil
}
