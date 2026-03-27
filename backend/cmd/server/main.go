package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/db"
	"github.com/peter/tacticarium/backend/internal/handler"
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

	// Run migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// WebSocket hub
	hub := ws.NewHub()

	// Handlers
	authHandler := handler.NewAuthHandler(pool, cfg)
	factionHandler := handler.NewFactionHandler(pool)
	missionHandler := handler.NewMissionHandler(pool)
	gameHandler := handler.NewGameHandler(pool, hub, cfg.JWTSecret)

	// Wire up state persistence
	hub.OnStateChange = gameHandler.PersistGameState

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes (public)
	r.Get("/api/auth/discord", authHandler.HandleDiscordRedirect)
	r.Get("/api/auth/discord/callback", authHandler.HandleDiscordCallback)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))

		// Auth
		r.Get("/api/auth/me", authHandler.HandleMe)
		r.Post("/api/auth/logout", authHandler.HandleLogout)

		// Factions
		r.Get("/api/factions", factionHandler.ListFactions)
		r.Get("/api/factions/{factionId}/detachments", factionHandler.ListDetachments)
		r.Get("/api/factions/{factionId}/stratagems", factionHandler.ListStratagems)
		r.Get("/api/detachments/{detachmentId}/stratagems", factionHandler.ListDetachmentStratagems)

		// Missions
		r.Get("/api/mission-packs", missionHandler.ListMissionPacks)
		r.Get("/api/mission-packs/{packId}/missions", missionHandler.ListMissions)
		r.Get("/api/mission-packs/{packId}/secondaries", missionHandler.ListSecondaries)
		r.Get("/api/mission-packs/{packId}/gambits", missionHandler.ListGambits)

		// Games
		r.Post("/api/games", gameHandler.CreateGame)
		r.Get("/api/games", gameHandler.ListGames)
		r.Post("/api/games/join/{code}", gameHandler.JoinGame)
		r.Get("/api/games/{gameId}", gameHandler.GetGame)
		r.Get("/api/games/{gameId}/events", gameHandler.GetGameEvents)

		// History
		r.Get("/api/users/me/history", gameHandler.GetHistory)
	})

	// WebSocket (auth via query param)
	r.Get("/ws/game/{gameId}", gameHandler.HandleWebSocket)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
