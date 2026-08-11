// This program provides the api service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ardanlabs/practical-ai-gcuk-2026/app/sdk/mux"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/domain/userbus/stores/userdb"
	"github.com/ardanlabs/practical-ai-gcuk-2026/business/sdk/sqldb"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(context.Background(), log); err != nil {
		log.Error("startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		apiHost         string
		readTimeout     time.Duration
		writeTimeout    time.Duration
		idleTimeout     time.Duration
		shutdownTimeout time.Duration
		db              sqldb.Config
	}{
		apiHost:         envString("API_HOST", "0.0.0.0:3000"),
		readTimeout:     envDuration("API_READ_TIMEOUT", 5*time.Second),
		writeTimeout:    envDuration("API_WRITE_TIMEOUT", 10*time.Second),
		idleTimeout:     envDuration("API_IDLE_TIMEOUT", 120*time.Second),
		shutdownTimeout: envDuration("API_SHUTDOWN_TIMEOUT", 20*time.Second),
		db: sqldb.Config{
			User:         envString("DB_USER", "postgres"),
			Password:     envString("DB_PASSWORD", "postgres"),
			HostPort:     envString("DB_HOST_PORT", "localhost:5432"),
			Name:         envString("DB_NAME", "postgres"),
			MaxIdleConns: envInt("DB_MAX_IDLE_CONNS", 2),
			MaxOpenConns: envInt("DB_MAX_OPEN_CONNS", 0),
			DisableTLS:   envBool("DB_DISABLE_TLS", true),
		},
	}

	// -------------------------------------------------------------------------
	// Database Support

	log.InfoContext(ctx, "startup", "status", "initializing database support", "hostport", cfg.db.HostPort)

	db, err := sqldb.Open(cfg.db)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	if err := sqldb.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("database status check: %w", err)
	}

	// -------------------------------------------------------------------------
	// Domain Support

	userBus := userbus.NewBusiness(log, userdb.NewStore(log, db))

	// -------------------------------------------------------------------------
	// Start API Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.apiHost,
		Handler:      mux.WebAPI(mux.Config{Log: log, UserBus: userBus}),
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
		IdleTimeout:  cfg.idleTimeout,
		ErrorLog:     slog.NewLogLogger(log.Handler(), slog.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.InfoContext(ctx, "startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		log.InfoContext(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.InfoContext(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.shutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			_ = api.Close()

			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func envString(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return def
}

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}

	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}

	return def
}
