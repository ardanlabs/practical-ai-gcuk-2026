// Package main implements a small HTTP service exposing a healthcheck, a
// greeting, and a list of users read from Postgres.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // register the "pgx" database/sql driver
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("shutdown", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":3000"
	}

	srv := http.Server{
		Addr:              addr,
		Handler:           routes(log, db),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	}
}

func routes(log *slog.Logger, db *sqlx.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthcheck", healthcheck)
	mux.HandleFunc("GET /hello/{user}", hello)
	mux.HandleFunc("GET /users", listUsers(log, db))

	return mux
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func hello(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{
		"message": "Hello, " + r.PathValue("user"),
	})
}

// User is a user of the system.
type User struct {
	ID        int       `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Email     string    `db:"email"      json:"email"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

func listUsers(log *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const q = `SELECT id, name, email, created_at FROM users ORDER BY id`

	return func(w http.ResponseWriter, r *http.Request) {
		users := []User{}
		if err := db.SelectContext(r.Context(), &users, q); err != nil {
			log.Error("list users", "err", err)
			respond(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		respond(w, http.StatusOK, users)
	}
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("encode response", "err", err)
	}
}
