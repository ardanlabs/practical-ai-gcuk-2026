package checkapp

import (
	"log/slog"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/api"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *slog.Logger
}

// Routes adds specific routes for this group.
func Routes(mux *http.ServeMux, cfg Config) {
	app := newApp()

	mux.HandleFunc("GET /healthcheck", api.Wrap(cfg.Log, app.liveness))
}
