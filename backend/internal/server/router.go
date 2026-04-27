package server

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/handler"
	"github.com/peter/tacticarium/backend/internal/ws"
)

// Handlers groups all handler instances needed by the router.
type Handlers struct {
	Auth      *handler.AuthHandler
	AdminAuth *handler.AdminAuthHandler
	Admin     *handler.AdminHandler
	Faction   *handler.FactionHandler
	Mission   *handler.MissionHandler
	Game      *handler.GameHandler
}

// NewHandlers constructs all handler instances. Pass nil for pool/hub when
// only the OpenAPI spec is needed (handlers won't be invoked).
func NewHandlers(pool *pgxpool.Pool, hub *ws.Hub, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:      handler.NewAuthHandler(pool, cfg),
		AdminAuth: handler.NewAdminAuthHandler(cfg),
		Admin:     handler.NewAdminHandler(pool),
		Faction:   handler.NewFactionHandler(pool),
		Mission:   handler.NewMissionHandler(pool),
		Game:      handler.NewGameHandler(pool, hub, cfg.JWTSecret),
	}
}

// NewAPI registers all huma operations on the given chi router and returns
// the huma.API. This is used both by NewRouter (for serving) and by the
// openapi command (for spec extraction without a running server or database).
func NewAPI(r chi.Router, h *Handlers, jwtSecret string) huma.API {

	humaConfig := huma.DefaultConfig("Tacticarium API", "1.0.0")
	humaConfig.Info.Description = "API for the Tacticarium turn tracker"
	api := humachi.New(r, humaConfig)

	// Security schemes
	api.OpenAPI().Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Player JWT token (Discord OAuth)",
		},
		"adminBearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Admin JWT token (GitHub OAuth)",
		},
	}

	playerSecurity := []map[string][]string{{"bearerAuth": {}}}
	adminSecurity := []map[string][]string{{"adminBearerAuth": {}}}

	playerMiddleware := auth.HumaMiddleware(jwtSecret)
	adminMiddleware := auth.HumaAdminMiddleware(jwtSecret)

	// --- Health check ---
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "Health check",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*handler.HealthOutput, error) {
		out := &handler.HealthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	// --- Player auth (huma, protected) ---
	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/auth/me",
		Summary:     "Get current user info",
		Tags:        []string{"Auth"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Auth.HandleMe)

	// --- Player faction endpoints ---
	huma.Register(api, huma.Operation{
		OperationID: "list-factions",
		Method:      http.MethodGet,
		Path:        "/api/factions",
		Summary:     "List all factions",
		Tags:        []string{"Factions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Faction.ListFactions)

	huma.Register(api, huma.Operation{
		OperationID: "list-detachments",
		Method:      http.MethodGet,
		Path:        "/api/factions/{factionId}/detachments",
		Summary:     "List detachments for a faction",
		Tags:        []string{"Factions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Faction.ListDetachments)

	huma.Register(api, huma.Operation{
		OperationID: "list-faction-stratagems",
		Method:      http.MethodGet,
		Path:        "/api/factions/{factionId}/stratagems",
		Summary:     "List stratagems for a faction",
		Tags:        []string{"Factions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Faction.ListStratagems)

	huma.Register(api, huma.Operation{
		OperationID: "list-detachment-stratagems",
		Method:      http.MethodGet,
		Path:        "/api/detachments/{detachmentId}/stratagems",
		Summary:     "List stratagems for a detachment",
		Tags:        []string{"Factions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Faction.ListDetachmentStratagems)

	// --- Player mission endpoints ---
	huma.Register(api, huma.Operation{
		OperationID: "list-mission-packs",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs",
		Summary:     "List all mission packs",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListMissionPacks)

	huma.Register(api, huma.Operation{
		OperationID: "list-missions",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs/{packId}/missions",
		Summary:     "List missions in a pack",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListMissions)

	huma.Register(api, huma.Operation{
		OperationID: "list-secondaries",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs/{packId}/secondaries",
		Summary:     "List secondary objectives in a pack",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListSecondaries)

	huma.Register(api, huma.Operation{
		OperationID: "list-gambits",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs/{packId}/gambits",
		Summary:     "List gambits in a pack",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListGambits)

	huma.Register(api, huma.Operation{
		OperationID: "list-mission-rules",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs/{packId}/rules",
		Summary:     "List mission rules in a pack",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListMissionRules)

	huma.Register(api, huma.Operation{
		OperationID: "list-challenger-cards",
		Method:      http.MethodGet,
		Path:        "/api/mission-packs/{packId}/challenger-cards",
		Summary:     "List challenger cards in a pack",
		Tags:        []string{"Missions"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Mission.ListChallengerCards)

	// --- Game endpoints ---
	huma.Register(api, huma.Operation{
		OperationID:   "create-game",
		Method:        http.MethodPost,
		Path:          "/api/games",
		Summary:       "Create a new game",
		Tags:          []string{"Games"},
		Security:      playerSecurity,
		Middlewares:   huma.Middlewares{playerMiddleware},
		DefaultStatus: 201,
	}, h.Game.CreateGame)

	huma.Register(api, huma.Operation{
		OperationID: "list-games",
		Method:      http.MethodGet,
		Path:        "/api/games",
		Summary:     "List user's games",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.ListGames)

	huma.Register(api, huma.Operation{
		OperationID: "join-game",
		Method:      http.MethodPost,
		Path:        "/api/games/join/{code}",
		Summary:     "Join a game by invite code",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.JoinGame)

	huma.Register(api, huma.Operation{
		OperationID: "get-game",
		Method:      http.MethodGet,
		Path:        "/api/games/{gameId}",
		Summary:     "Get game state",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.GetGame)

	huma.Register(api, huma.Operation{
		OperationID:   "hide-game",
		Method:        http.MethodPost,
		Path:          "/api/games/{gameId}/hide",
		Summary:       "Hide a game from the current user's game list",
		Tags:          []string{"Games"},
		Security:      playerSecurity,
		Middlewares:   huma.Middlewares{playerMiddleware},
		DefaultStatus: 204,
	}, h.Game.HideGame)

	huma.Register(api, huma.Operation{
		OperationID: "get-game-events",
		Method:      http.MethodGet,
		Path:        "/api/games/{gameId}/events",
		Summary:     "Get game event history",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.GetGameEvents)

	huma.Register(api, huma.Operation{
		OperationID: "get-history",
		Method:      http.MethodGet,
		Path:        "/api/users/me/history",
		Summary:     "Get user's completed game history",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.GetHistory)

	huma.Register(api, huma.Operation{
		OperationID: "get-stats",
		Method:      http.MethodGet,
		Path:        "/api/users/me/stats",
		Summary:     "Get user's game stats",
		Tags:        []string{"Games"},
		Security:    playerSecurity,
		Middlewares: huma.Middlewares{playerMiddleware},
	}, h.Game.GetStats)

	// --- Admin auth (huma) ---
	huma.Register(api, huma.Operation{
		OperationID: "get-admin-me",
		Method:      http.MethodGet,
		Path:        "/api/admin/me",
		Summary:     "Get current admin info",
		Tags:        []string{"Admin Auth"},
		Security:    adminSecurity,
		Middlewares: huma.Middlewares{adminMiddleware},
	}, h.AdminAuth.HandleAdminMe)

	// --- Admin CRUD ---
	registerAdminCRUD(api, adminMiddleware, adminSecurity, h.Admin)

	return api
}

func NewRouter(pool *pgxpool.Pool, hub *ws.Hub, cfg *config.Config) http.Handler {
	h := NewHandlers(pool, hub, cfg)

	hub.OnStateChange = h.Game.PersistGameState

	r := chi.NewRouter()

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

	// Register all huma operations on this router
	_ = NewAPI(r, h, cfg.JWTSecret)

	// --- Auth routes (raw chi - OAuth redirects) ---
	r.Get("/api/auth/discord", h.Auth.HandleDiscordRedirect)
	r.Get("/api/auth/discord/callback", h.Auth.HandleDiscordCallback)
	r.Get("/api/auth/github", h.AdminAuth.HandleGitHubRedirect)
	r.Get("/api/auth/github/callback", h.AdminAuth.HandleGitHubCallback)

	// Logout stays as raw chi (needs to set cookies)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/api/auth/logout", h.Auth.HandleLogout)
	})

	// --- Admin imports (raw chi - multipart file upload) ---
	r.Group(func(r chi.Router) {
		r.Use(auth.AdminMiddleware(cfg.JWTSecret))
		r.Post("/api/admin/import/factions", h.Admin.ImportFactions)
		r.Post("/api/admin/import/detachments", h.Admin.ImportDetachments)
		r.Post("/api/admin/import/stratagems", h.Admin.ImportStratagems)
		r.Post("/api/admin/import/missions", h.Admin.ImportMissions)
	})

	// --- WebSocket (raw chi - auth via query param) ---
	r.Get("/ws/game/{gameId}", h.Game.HandleWebSocket)
	// Public read-only spectator WebSocket; no auth.
	r.Get("/ws/game/{gameId}/spectate", h.Game.HandleSpectatorWebSocket)

	// Wrap with OTEL HTTP instrumentation
	return otelhttp.NewHandler(r, "tacticarium")
}

// registerAdminCRUD registers all admin CRUD operations with huma.
func registerAdminCRUD(api huma.API, mw func(huma.Context, func(huma.Context)), security []map[string][]string, h *handler.AdminHandler) {
	op := func(id, method, path, summary, tag string) huma.Operation {
		return huma.Operation{
			OperationID:   id,
			Method:        method,
			Path:          path,
			Summary:       summary,
			Tags:          []string{tag},
			Security:      security,
			Middlewares:    huma.Middlewares{mw},
			DefaultStatus: 200,
		}
	}

	// Factions
	huma.Register(api, op("admin-list-factions", http.MethodGet, "/api/admin/factions", "List all factions", "Admin Factions"), h.ListFactions)
	huma.Register(api, op("admin-get-faction", http.MethodGet, "/api/admin/factions/{id}", "Get a faction", "Admin Factions"), h.GetFaction)
	cr := op("admin-create-faction", http.MethodPost, "/api/admin/factions", "Create a faction", "Admin Factions")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateFaction)
	huma.Register(api, op("admin-update-faction", http.MethodPut, "/api/admin/factions/{id}", "Update a faction", "Admin Factions"), h.UpdateFaction)
	del := op("admin-delete-faction", http.MethodDelete, "/api/admin/factions/{id}", "Delete a faction", "Admin Factions")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteFaction)

	// Detachments
	huma.Register(api, op("admin-list-detachments", http.MethodGet, "/api/admin/detachments", "List detachments", "Admin Detachments"), h.ListDetachments)
	huma.Register(api, op("admin-get-detachment", http.MethodGet, "/api/admin/detachments/{id}", "Get a detachment", "Admin Detachments"), h.GetDetachment)
	cr = op("admin-create-detachment", http.MethodPost, "/api/admin/detachments", "Create a detachment", "Admin Detachments")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateDetachment)
	huma.Register(api, op("admin-update-detachment", http.MethodPut, "/api/admin/detachments/{id}", "Update a detachment", "Admin Detachments"), h.UpdateDetachment)
	del = op("admin-delete-detachment", http.MethodDelete, "/api/admin/detachments/{id}", "Delete a detachment", "Admin Detachments")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteDetachment)

	// Stratagems
	huma.Register(api, op("admin-list-stratagems", http.MethodGet, "/api/admin/stratagems", "List stratagems", "Admin Stratagems"), h.ListStratagems)
	huma.Register(api, op("admin-get-stratagem", http.MethodGet, "/api/admin/stratagems/{id}", "Get a stratagem", "Admin Stratagems"), h.GetStratagem)
	cr = op("admin-create-stratagem", http.MethodPost, "/api/admin/stratagems", "Create a stratagem", "Admin Stratagems")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateStratagem)
	huma.Register(api, op("admin-update-stratagem", http.MethodPut, "/api/admin/stratagems/{id}", "Update a stratagem", "Admin Stratagems"), h.UpdateStratagem)
	del = op("admin-delete-stratagem", http.MethodDelete, "/api/admin/stratagems/{id}", "Delete a stratagem", "Admin Stratagems")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteStratagem)

	// Mission Packs
	huma.Register(api, op("admin-list-mission-packs", http.MethodGet, "/api/admin/mission-packs", "List mission packs", "Admin Mission Packs"), h.ListMissionPacks)
	cr = op("admin-create-mission-pack", http.MethodPost, "/api/admin/mission-packs", "Create a mission pack", "Admin Mission Packs")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateMissionPack)
	huma.Register(api, op("admin-update-mission-pack", http.MethodPut, "/api/admin/mission-packs/{id}", "Update a mission pack", "Admin Mission Packs"), h.UpdateMissionPack)
	del = op("admin-delete-mission-pack", http.MethodDelete, "/api/admin/mission-packs/{id}", "Delete a mission pack", "Admin Mission Packs")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteMissionPack)

	// Missions
	huma.Register(api, op("admin-list-missions", http.MethodGet, "/api/admin/missions", "List missions", "Admin Missions"), h.ListMissions)
	huma.Register(api, op("admin-get-mission", http.MethodGet, "/api/admin/missions/{id}", "Get a mission", "Admin Missions"), h.GetMission)
	cr = op("admin-create-mission", http.MethodPost, "/api/admin/missions", "Create a mission", "Admin Missions")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateMission)
	huma.Register(api, op("admin-update-mission", http.MethodPut, "/api/admin/missions/{id}", "Update a mission", "Admin Missions"), h.UpdateMission)
	del = op("admin-delete-mission", http.MethodDelete, "/api/admin/missions/{id}", "Delete a mission", "Admin Missions")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteMission)

	// Secondaries
	huma.Register(api, op("admin-list-secondaries", http.MethodGet, "/api/admin/secondaries", "List secondaries", "Admin Secondaries"), h.ListSecondaries)
	huma.Register(api, op("admin-get-secondary", http.MethodGet, "/api/admin/secondaries/{id}", "Get a secondary", "Admin Secondaries"), h.GetSecondary)
	cr = op("admin-create-secondary", http.MethodPost, "/api/admin/secondaries", "Create a secondary", "Admin Secondaries")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateSecondary)
	huma.Register(api, op("admin-update-secondary", http.MethodPut, "/api/admin/secondaries/{id}", "Update a secondary", "Admin Secondaries"), h.UpdateSecondary)
	del = op("admin-delete-secondary", http.MethodDelete, "/api/admin/secondaries/{id}", "Delete a secondary", "Admin Secondaries")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteSecondary)

	// Gambits
	huma.Register(api, op("admin-list-gambits", http.MethodGet, "/api/admin/gambits", "List gambits", "Admin Gambits"), h.ListGambits)
	huma.Register(api, op("admin-get-gambit", http.MethodGet, "/api/admin/gambits/{id}", "Get a gambit", "Admin Gambits"), h.GetGambit)
	cr = op("admin-create-gambit", http.MethodPost, "/api/admin/gambits", "Create a gambit", "Admin Gambits")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateGambit)
	huma.Register(api, op("admin-update-gambit", http.MethodPut, "/api/admin/gambits/{id}", "Update a gambit", "Admin Gambits"), h.UpdateGambit)
	del = op("admin-delete-gambit", http.MethodDelete, "/api/admin/gambits/{id}", "Delete a gambit", "Admin Gambits")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteGambit)

	// Challenger Cards
	huma.Register(api, op("admin-list-challenger-cards", http.MethodGet, "/api/admin/challenger-cards", "List challenger cards", "Admin Challenger Cards"), h.ListChallengerCards)
	huma.Register(api, op("admin-get-challenger-card", http.MethodGet, "/api/admin/challenger-cards/{id}", "Get a challenger card", "Admin Challenger Cards"), h.GetChallengerCard)
	cr = op("admin-create-challenger-card", http.MethodPost, "/api/admin/challenger-cards", "Create a challenger card", "Admin Challenger Cards")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateChallengerCard)
	huma.Register(api, op("admin-update-challenger-card", http.MethodPut, "/api/admin/challenger-cards/{id}", "Update a challenger card", "Admin Challenger Cards"), h.UpdateChallengerCard)
	del = op("admin-delete-challenger-card", http.MethodDelete, "/api/admin/challenger-cards/{id}", "Delete a challenger card", "Admin Challenger Cards")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteChallengerCard)

	// Mission Rules
	huma.Register(api, op("admin-list-mission-rules", http.MethodGet, "/api/admin/mission-rules", "List mission rules", "Admin Mission Rules"), h.ListMissionRules)
	huma.Register(api, op("admin-get-mission-rule", http.MethodGet, "/api/admin/mission-rules/{id}", "Get a mission rule", "Admin Mission Rules"), h.GetMissionRule)
	cr = op("admin-create-mission-rule", http.MethodPost, "/api/admin/mission-rules", "Create a mission rule", "Admin Mission Rules")
	cr.DefaultStatus = 201
	huma.Register(api, cr, h.CreateMissionRule)
	huma.Register(api, op("admin-update-mission-rule", http.MethodPut, "/api/admin/mission-rules/{id}", "Update a mission rule", "Admin Mission Rules"), h.UpdateMissionRule)
	del = op("admin-delete-mission-rule", http.MethodDelete, "/api/admin/mission-rules/{id}", "Delete a mission rule", "Admin Mission Rules")
	del.DefaultStatus = 204
	huma.Register(api, del, h.DeleteMissionRule)
}
