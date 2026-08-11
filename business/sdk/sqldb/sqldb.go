// Package sqldb provides support for access the database.
package sqldb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Calls init function.
	"github.com/jmoiron/sqlx"
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	HostPort     string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := url.Values{}
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.HostPort,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("pgx", u.String())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	for attempts := 1; ; attempts++ {
		pingError := db.PingContext(ctx)
		if pingError == nil {
			break
		}

		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), pingError)
		case <-time.After(time.Duration(attempts) * 100 * time.Millisecond):
		}
	}

	const q = `SELECT true`
	var tmp bool

	return db.QueryRowContext(ctx, q).Scan(&tmp)
}
