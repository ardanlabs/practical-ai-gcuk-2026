package userapp

import (
	"log/slog"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/api"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log     *slog.Logger
	UserBus *userbus.Business
}

// Routes adds specific routes for this group.
func Routes(mux *http.ServeMux, cfg Config) {
	app := newApp(cfg.UserBus)

	mux.HandleFunc("GET /users", api.Wrap(cfg.Log, app.queryAll))
}
