// Package web wires the HTTP routes and handlers for the service.
package web

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/internal/user"
)

// Handlers holds the dependencies shared by the HTTP handlers.
type Handlers struct {
	log   *slog.Logger
	users *user.Store
}

// NewMux returns a mux with all routes registered.
func NewMux(log *slog.Logger, users *user.Store) *http.ServeMux {
	h := Handlers{log: log, users: users}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /hello/{user}", h.hello)
	mux.HandleFunc("GET /users", h.listUsers)

	return mux
}

func (h Handlers) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h Handlers) hello(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("user")

	h.respond(w, r, http.StatusOK, map[string]string{
		"message": "Hello, " + name,
	})
}

func (h Handlers) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "listing users", "err", err)
		h.respond(w, r, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	if users == nil {
		users = []user.User{}
	}

	h.respond(w, r, http.StatusOK, users)
}

func (h Handlers) respond(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.ErrorContext(r.Context(), "encoding response", "err", err)
	}
}
