package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/handler"
	"github.com/peter/tacticarium/backend/internal/ws"
)

func NewRouter(pool *pgxpool.Pool, hub *ws.Hub, cfg *config.Config) chi.Router {
	authHandler := handler.NewAuthHandler(pool, cfg)
	adminAuthHandler := handler.NewAdminAuthHandler(cfg)
	adminHandler := handler.NewAdminHandler(pool)
	factionHandler := handler.NewFactionHandler(pool)
	missionHandler := handler.NewMissionHandler(pool)
	gameHandler := handler.NewGameHandler(pool, hub, cfg.JWTSecret)

	hub.OnStateChange = gameHandler.PersistGameState

	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)

	allowedOrigins := []string{cfg.FrontendURL}
	if cfg.AdminFrontendURL != "" {
		allowedOrigins = append(allowedOrigins, cfg.AdminFrontendURL)
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
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
	r.Get("/api/auth/github", adminAuthHandler.HandleGitHubRedirect)
	r.Get("/api/auth/github/callback", adminAuthHandler.HandleGitHubCallback)

	// Protected routes (player - Discord auth)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))

		r.Get("/api/auth/me", authHandler.HandleMe)
		r.Post("/api/auth/logout", authHandler.HandleLogout)

		r.Get("/api/factions", factionHandler.ListFactions)
		r.Get("/api/factions/{factionId}/detachments", factionHandler.ListDetachments)
		r.Get("/api/factions/{factionId}/stratagems", factionHandler.ListStratagems)
		r.Get("/api/detachments/{detachmentId}/stratagems", factionHandler.ListDetachmentStratagems)

		r.Get("/api/mission-packs", missionHandler.ListMissionPacks)
		r.Get("/api/mission-packs/{packId}/missions", missionHandler.ListMissions)
		r.Get("/api/mission-packs/{packId}/secondaries", missionHandler.ListSecondaries)
		r.Get("/api/mission-packs/{packId}/gambits", missionHandler.ListGambits)
		r.Get("/api/mission-packs/{packId}/rules", missionHandler.ListMissionRules)
		r.Get("/api/mission-packs/{packId}/challenger-cards", missionHandler.ListChallengerCards)

		r.Post("/api/games", gameHandler.CreateGame)
		r.Get("/api/games", gameHandler.ListGames)
		r.Post("/api/games/join/{code}", gameHandler.JoinGame)
		r.Get("/api/games/{gameId}", gameHandler.GetGame)
		r.Get("/api/games/{gameId}/events", gameHandler.GetGameEvents)

		r.Get("/api/users/me/history", gameHandler.GetHistory)
	})

	// Admin routes (GitHub auth)
	r.Group(func(r chi.Router) {
		r.Use(auth.AdminMiddleware(cfg.JWTSecret))

		r.Get("/api/admin/me", adminAuthHandler.HandleAdminMe)

		// Factions
		r.Get("/api/admin/factions", adminHandler.ListFactions)
		r.Get("/api/admin/factions/{id}", adminHandler.GetFaction)
		r.Post("/api/admin/factions", adminHandler.CreateFaction)
		r.Put("/api/admin/factions/{id}", adminHandler.UpdateFaction)
		r.Delete("/api/admin/factions/{id}", adminHandler.DeleteFaction)

		// Detachments
		r.Get("/api/admin/detachments", adminHandler.ListDetachments)
		r.Get("/api/admin/detachments/{id}", adminHandler.GetDetachment)
		r.Post("/api/admin/detachments", adminHandler.CreateDetachment)
		r.Put("/api/admin/detachments/{id}", adminHandler.UpdateDetachment)
		r.Delete("/api/admin/detachments/{id}", adminHandler.DeleteDetachment)

		// Stratagems
		r.Get("/api/admin/stratagems", adminHandler.ListStratagems)
		r.Get("/api/admin/stratagems/{id}", adminHandler.GetStratagem)
		r.Post("/api/admin/stratagems", adminHandler.CreateStratagem)
		r.Put("/api/admin/stratagems/{id}", adminHandler.UpdateStratagem)
		r.Delete("/api/admin/stratagems/{id}", adminHandler.DeleteStratagem)

		// Mission Packs
		r.Get("/api/admin/mission-packs", adminHandler.ListMissionPacks)
		r.Post("/api/admin/mission-packs", adminHandler.CreateMissionPack)
		r.Put("/api/admin/mission-packs/{id}", adminHandler.UpdateMissionPack)
		r.Delete("/api/admin/mission-packs/{id}", adminHandler.DeleteMissionPack)

		// Missions
		r.Get("/api/admin/missions", adminHandler.ListMissions)
		r.Get("/api/admin/missions/{id}", adminHandler.GetMission)
		r.Post("/api/admin/missions", adminHandler.CreateMission)
		r.Put("/api/admin/missions/{id}", adminHandler.UpdateMission)
		r.Delete("/api/admin/missions/{id}", adminHandler.DeleteMission)

		// Secondaries
		r.Get("/api/admin/secondaries", adminHandler.ListSecondaries)
		r.Get("/api/admin/secondaries/{id}", adminHandler.GetSecondary)
		r.Post("/api/admin/secondaries", adminHandler.CreateSecondary)
		r.Put("/api/admin/secondaries/{id}", adminHandler.UpdateSecondary)
		r.Delete("/api/admin/secondaries/{id}", adminHandler.DeleteSecondary)

		// Gambits
		r.Get("/api/admin/gambits", adminHandler.ListGambits)
		r.Get("/api/admin/gambits/{id}", adminHandler.GetGambit)
		r.Post("/api/admin/gambits", adminHandler.CreateGambit)
		r.Put("/api/admin/gambits/{id}", adminHandler.UpdateGambit)
		r.Delete("/api/admin/gambits/{id}", adminHandler.DeleteGambit)

		// Challenger Cards
		r.Get("/api/admin/challenger-cards", adminHandler.ListChallengerCards)
		r.Get("/api/admin/challenger-cards/{id}", adminHandler.GetChallengerCard)
		r.Post("/api/admin/challenger-cards", adminHandler.CreateChallengerCard)
		r.Put("/api/admin/challenger-cards/{id}", adminHandler.UpdateChallengerCard)
		r.Delete("/api/admin/challenger-cards/{id}", adminHandler.DeleteChallengerCard)

		// Mission Rules
		r.Get("/api/admin/mission-rules", adminHandler.ListMissionRules)
		r.Get("/api/admin/mission-rules/{id}", adminHandler.GetMissionRule)
		r.Post("/api/admin/mission-rules", adminHandler.CreateMissionRule)
		r.Put("/api/admin/mission-rules/{id}", adminHandler.UpdateMissionRule)
		r.Delete("/api/admin/mission-rules/{id}", adminHandler.DeleteMissionRule)

		// Bulk import
		r.Post("/api/admin/import/factions", adminHandler.ImportFactions)
		r.Post("/api/admin/import/stratagems", adminHandler.ImportStratagems)
		r.Post("/api/admin/import/missions", adminHandler.ImportMissions)
	})

	// WebSocket (auth via query param)
	r.Get("/ws/game/{gameId}", gameHandler.HandleWebSocket)

	return r
}
