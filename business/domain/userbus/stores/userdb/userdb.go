// Package userdb provides postgres access to the user domain.
package userdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log *slog.Logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(log *slog.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// QueryAll retrieves every user from the database.
func (s *Store) QueryAll(ctx context.Context) ([]userbus.User, error) {
	const q = `
	SELECT
		user_id, name, email, enabled, date_created, date_updated
	FROM
		users
	ORDER BY
		user_id`

	var dbUsrs []dbUser
	if err := s.db.SelectContext(ctx, &dbUsrs, q); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return toBusUsers(dbUsrs)
}
