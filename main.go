package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type user struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type server struct {
	db *sqlx.DB
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer db.Close()

	srv := server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthcheck", healthcheck)
	mux.HandleFunc("GET /hello/{user}", hello)
	mux.HandleFunc("GET /users", srv.listUsers)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s", r.PathValue("user"))
}

func (s server) listUsers(w http.ResponseWriter, r *http.Request) {
	users := []user{}
	const q = `SELECT id, name, email, created_at FROM users ORDER BY id`
	if err := s.db.SelectContext(r.Context(), &users, q); err != nil {
		log.Printf("list users: %v", err)
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("encode users: %v", err)
	}
}
