package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/db"
	"github.com/peter/tacticarium/backend/internal/logging"
	"github.com/peter/tacticarium/backend/internal/server"
	"github.com/peter/tacticarium/backend/internal/telemetry"
	"github.com/peter/tacticarium/backend/internal/ws"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// OpenTelemetry (traces, metrics, logs)
	lp, shutdown, err := telemetry.Init(ctx, "tacticarium", "1.0.0")
	if err != nil {
		// Set up logging without OTEL export so we can still log to stdout
		logging.Init(nil)
		slog.Warn("Failed to init telemetry, continuing without observability", "error", err)
	} else {
		defer shutdown(ctx)
		logging.Init(lp)
	}

	// Database
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// WebSocket hub
	hub := ws.NewHub()

	// Router
	r := server.NewRouter(pool, hub, cfg)

	slog.Info("Server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
