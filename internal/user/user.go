// Package user provides access to the users stored in the database.
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// User represents a single user in the system.
type User struct {
	ID        int       `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Email     string    `db:"email"      json:"email"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// Store provides read access to users.
type Store struct {
	db *sqlx.DB
}

// NewStore constructs a Store backed by db.
func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

// List returns every user in the system, ordered by id.
func (s *Store) List(ctx context.Context) ([]User, error) {
	const q = `SELECT id, name, email, created_at FROM users ORDER BY id`

	var users []User
	if err := s.db.SelectContext(ctx, &users, q); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return users, nil
}
