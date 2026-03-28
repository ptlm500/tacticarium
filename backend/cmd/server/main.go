package main

import (
	"context"
	"log"
	"net/http"

	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/db"
	"github.com/peter/tacticarium/backend/internal/server"
	"github.com/peter/tacticarium/backend/internal/ws"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	// Database
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations (Now passes the URL instead of the pool)
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// WebSocket hub
	hub := ws.NewHub()

	// Router
	r := server.NewRouter(pool, hub, cfg)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
