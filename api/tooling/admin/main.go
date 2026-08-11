// This program provides administrative support for the project, currently the
// database migrations.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/env"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/sdk/migrate"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/sdk/sqldb"
	db "github.com/ardanlabs/practical-ai-gcuk-2026/zarf/schemas/db"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

const usage = `Usage: admin <command>

Commands:
  migrate         apply all outstanding migrations
  migrate-down    roll back the most recently applied migration
  migrate-status  show the state of every migration
`

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "admin: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) != 2 {
		fmt.Print(usage)

		return errors.New("a single command is required")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	sqlDB, err := sqldb.Open(env.DBConfig())
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer sqlDB.Close()

	if err := sqldb.StatusCheck(ctx, sqlDB); err != nil {
		return fmt.Errorf("database status check: %w", err)
	}

	switch cmd := os.Args[1]; cmd {
	case "migrate":
		return migrateUp(ctx, sqlDB)

	case "migrate-down":
		return migrateDown(ctx, sqlDB)

	case "migrate-status":
		return migrateStatus(ctx, sqlDB)

	default:
		fmt.Print(usage)

		return fmt.Errorf("unknown command %q", cmd)
	}
}

func migrateUp(ctx context.Context, sqlDB *sqlx.DB) error {
	results, err := migrate.Up(ctx, sqlDB, db.Migrations())
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("migrations up to date, nothing to apply")

		return nil
	}

	for _, r := range results {
		fmt.Printf("applied %s in %s\n", r.Source.Path, r.Duration)
	}

	return nil
}

func migrateDown(ctx context.Context, sqlDB *sqlx.DB) error {
	result, err := migrate.Down(ctx, sqlDB, db.Migrations())
	if err != nil {
		return err
	}

	fmt.Printf("rolled back %s in %s\n", result.Source.Path, result.Duration)

	return nil
}

func migrateStatus(ctx context.Context, sqlDB *sqlx.DB) error {
	status, err := migrate.Status(ctx, sqlDB, db.Migrations())
	if err != nil {
		return err
	}

	for _, s := range status {
		state := string(s.State)
		if s.State == goose.StateApplied {
			state = s.AppliedAt.Format(time.RFC3339)
		}

		fmt.Printf("%-40s %s\n", s.Source.Path, state)
	}

	return nil
}
