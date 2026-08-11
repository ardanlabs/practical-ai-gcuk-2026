// Package migrate applies the database migrations using goose.
package migrate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

func newProvider(db *sqlx.DB, fsys fs.FS) (*goose.Provider, error) {
	p, err := goose.NewProvider(goose.DialectPostgres, db.DB, fsys)
	if err != nil {
		return nil, fmt.Errorf("new provider: %w", err)
	}

	return p, nil
}

// Up applies all outstanding migrations and reports the ones that ran.
func Up(ctx context.Context, db *sqlx.DB, fsys fs.FS) ([]*goose.MigrationResult, error) {
	p, err := newProvider(db, fsys)
	if err != nil {
		return nil, err
	}

	results, err := p.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("up: %w", err)
	}

	return results, nil
}

// Down rolls back the most recently applied migration.
func Down(ctx context.Context, db *sqlx.DB, fsys fs.FS) (*goose.MigrationResult, error) {
	p, err := newProvider(db, fsys)
	if err != nil {
		return nil, err
	}

	result, err := p.Down(ctx)
	if err != nil {
		return nil, fmt.Errorf("down: %w", err)
	}

	return result, nil
}

// Status reports the state of every known migration.
func Status(ctx context.Context, db *sqlx.DB, fsys fs.FS) ([]*goose.MigrationStatus, error) {
	p, err := newProvider(db, fsys)
	if err != nil {
		return nil, err
	}

	status, err := p.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}

	return status, nil
}
