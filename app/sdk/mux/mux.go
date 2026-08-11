// Package mux provides support to bind domain level routes to the http mux.
package mux

import (
	"log/slog"
	"net/http"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/domain/checkapp"
	"github.com/ardanlabs/practical-ai-gcuk-2026/app/domain/helloapp"
	"github.com/ardanlabs/practical-ai-gcuk-2026/app/domain/userapp"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log     *slog.Logger
	UserBus *userbus.Business
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config) http.Handler {
	mux := http.NewServeMux()

	checkapp.Routes(mux, checkapp.Config{Log: cfg.Log})
	helloapp.Routes(mux, helloapp.Config{Log: cfg.Log})
	userapp.Routes(mux, userapp.Config{Log: cfg.Log, UserBus: cfg.UserBus})

	return mux
}
