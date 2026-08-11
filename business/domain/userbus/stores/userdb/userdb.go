// Package userdb provides postgres access to the user domain.
package userdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// uniqueViolation is the postgres error code for a unique constraint failure.
const uniqueViolation = "23505"

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

// Create adds a new user to the database and returns the stored row.
func (s *Store) Create(ctx context.Context, usr userbus.User) (userbus.User, error) {
	const q = `
	INSERT INTO users
		(name, email, enabled)
	VALUES
		(:name, :email, :enabled)
	RETURNING
		user_id, name, email, enabled, date_created, date_updated`

	dbUsr, err := s.namedQueryRow(ctx, q, toDBUser(usr))
	if err != nil {
		if isUniqueEmailViolation(err) {
			return userbus.User{}, fmt.Errorf("namedquery: %w", userbus.ErrUniqueEmail)
		}

		return userbus.User{}, fmt.Errorf("namedquery: %w", err)
	}

	return toBusUser(dbUsr)
}

// Update modifies a user in the database and returns the stored row.
func (s *Store) Update(ctx context.Context, usr userbus.User) (userbus.User, error) {
	const q = `
	UPDATE
		users
	SET
		name = :name,
		email = :email,
		enabled = :enabled,
		date_updated = NOW()
	WHERE
		user_id = :user_id
	RETURNING
		user_id, name, email, enabled, date_created, date_updated`

	dbUsr, err := s.namedQueryRow(ctx, q, toDBUser(usr))
	if err != nil {
		if isUniqueEmailViolation(err) {
			return userbus.User{}, fmt.Errorf("namedquery: %w", userbus.ErrUniqueEmail)
		}

		if errors.Is(err, sql.ErrNoRows) {
			return userbus.User{}, fmt.Errorf("namedquery: %w", userbus.ErrNotFound)
		}

		return userbus.User{}, fmt.Errorf("namedquery: %w", err)
	}

	return toBusUser(dbUsr)
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr userbus.User) error {
	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if _, err := s.db.NamedExecContext(ctx, q, toDBUser(usr)); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
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

// QueryByID retrieves the user identified by the specified user id.
func (s *Store) QueryByID(ctx context.Context, userID int64) (userbus.User, error) {
	const q = `
	SELECT
		user_id, name, email, enabled, date_created, date_updated
	FROM
		users
	WHERE
		user_id = $1`

	var dbUsr dbUser
	if err := s.db.GetContext(ctx, &dbUsr, q, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userbus.User{}, fmt.Errorf("get: %w", userbus.ErrNotFound)
		}

		return userbus.User{}, fmt.Errorf("get: %w", err)
	}

	return toBusUser(dbUsr)
}

// namedQueryRow executes a named query that returns exactly one row and scans
// that row into a dbUser value.
func (s *Store) namedQueryRow(ctx context.Context, query string, data dbUser) (dbUser, error) {
	rows, err := s.db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return dbUser{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return dbUser{}, err
		}

		return dbUser{}, sql.ErrNoRows
	}

	var dbUsr dbUser
	if err := rows.StructScan(&dbUsr); err != nil {
		return dbUser{}, err
	}

	return dbUsr, nil
}

// isUniqueEmailViolation reports whether the error represents a postgres
// unique constraint violation on the email column.
func isUniqueEmailViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
