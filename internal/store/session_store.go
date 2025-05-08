package store

import (
	"database/sql"
	"errors"
	"time"
)

// Session represents a user session in the system
type Session struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id"`
	UserID       int64     `json:"user_id"`
	Token        string    `json:"token"`
	IPAddress    string    `json:"ip_address"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SessionStore interface defines the session-related operations
type SessionStore interface {
	CreateSession(*Session) (*Session, error)
	GetSessionByID(sessionID string) (*Session, error)
	GetSessionByToken(token string) (*Session, error)
	GetActiveSessionsByUserID(userID int64) ([]*Session, error)
	UpdateLastActivity(sessionID string) error
	DeactivateSession(sessionID string) error
	CleanupExpiredSessions() (int64, error)
}

// PostgresSessionStore implements SessionStore interface
type PostgresSessionStore struct {
	db *sql.DB
}

// NewPostgresSessionStore creates a new PostgresSessionStore
func NewPostgresSessionStore(db *sql.DB) *PostgresSessionStore {
	return &PostgresSessionStore{db: db}
}

// CreateSession creates a new session in the database
func (pg *PostgresSessionStore) CreateSession(session *Session) (*Session, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO sessions (session_id, user_id, token, ip_address, is_active, created_at, last_activity, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = tx.QueryRow(
		query,
		session.SessionID,
		session.UserID,
		session.Token,
		session.IPAddress,
		session.IsActive,
		session.CreatedAt,
		session.LastActivity,
		session.ExpiresAt,
	).Scan(&session.ID)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByID retrieves a session by its session_id
func (pg *PostgresSessionStore) GetSessionByID(sessionID string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, session_id, user_id, token, ip_address, is_active, created_at, last_activity, expires_at
		FROM sessions
		WHERE session_id = $1
	`
	err := pg.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.UserID,
		&session.Token,
		&session.IPAddress,
		&session.IsActive,
		&session.CreatedAt,
		&session.LastActivity,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByToken retrieves a session by its token
func (pg *PostgresSessionStore) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, session_id, user_id, token, ip_address, is_active, created_at, last_activity, expires_at
		FROM sessions
		WHERE token = $1
	`
	err := pg.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.SessionID,
		&session.UserID,
		&session.Token,
		&session.IPAddress,
		&session.IsActive,
		&session.CreatedAt,
		&session.LastActivity,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetActiveSessionsByUserID gets all active sessions for a user
func (pg *PostgresSessionStore) GetActiveSessionsByUserID(userID int64) ([]*Session, error) {
	query := `
		SELECT id, session_id, user_id, token, ip_address, is_active, created_at, last_activity, expires_at
		FROM sessions
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
	`
	rows, err := pg.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(
			&session.ID,
			&session.SessionID,
			&session.UserID,
			&session.Token,
			&session.IPAddress,
			&session.IsActive,
			&session.CreatedAt,
			&session.LastActivity,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// UpdateLastActivity updates the last_activity timestamp of a session
func (pg *PostgresSessionStore) UpdateLastActivity(sessionID string) error {
	query := `
		UPDATE sessions
		SET last_activity = CURRENT_TIMESTAMP
		WHERE session_id = $1 AND is_active = true
	`
	result, err := pg.db.Exec(query, sessionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("session not found or already inactive")
	}

	return nil
}

// DeactivateSession marks a session as inactive (logout)
func (pg *PostgresSessionStore) DeactivateSession(sessionID string) error {
	query := `
		UPDATE sessions
		SET is_active = false
		WHERE session_id = $1
	`
	result, err := pg.db.Exec(query, sessionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("session not found")
	}

	return nil
}

// CleanupExpiredSessions removes or deactivates expired sessions
func (pg *PostgresSessionStore) CleanupExpiredSessions() (int64, error) {
	// Option 1: Mark as inactive
	query := `
		UPDATE sessions
		SET is_active = false
		WHERE expires_at < NOW() AND is_active = true
	`

	result, err := pg.db.Exec(query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
