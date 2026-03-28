package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/game"
	"github.com/peter/tacticarium/backend/internal/models"
	"github.com/peter/tacticarium/backend/internal/ws"
	"github.com/peter/tacticarium/backend/pkg/invite"
	"nhooyr.io/websocket"
)

type GameHandler struct {
	db  *pgxpool.Pool
	hub *ws.Hub
	cfg interface {
		GetJWTSecret() string
	}
	jwtSecret string
}

func NewGameHandler(db *pgxpool.Pool, hub *ws.Hub, jwtSecret string) *GameHandler {
	return &GameHandler{
		db:        db,
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	code := invite.GenerateCode(6)

	var gameID string
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO games (invite_code) VALUES ($1) RETURNING id`, code,
	).Scan(&gameID)
	if err != nil {
		log.Printf("Create game error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Add creator as player 1
	_, err = h.db.Exec(r.Context(),
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, 1)`,
		gameID, user.UserID)
	if err != nil {
		log.Printf("Add player error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":         gameID,
		"inviteCode": code,
	})
}

func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	code := chi.URLParam(r, "code")

	var gameID, status string
	err := h.db.QueryRow(r.Context(),
		`SELECT id, status FROM games WHERE invite_code = $1`, code,
	).Scan(&gameID, &status)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	if status != "setup" {
		http.Error(w, "game already started", http.StatusBadRequest)
		return
	}

	// Check if already in game
	var count int
	h.db.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1 AND user_id = $2`,
		gameID, user.UserID,
	).Scan(&count)

	if count > 0 {
		writeJSON(w, http.StatusOK, map[string]string{"id": gameID, "inviteCode": code})
		return
	}

	// Check player count
	h.db.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID,
	).Scan(&count)

	if count >= 2 {
		http.Error(w, "game is full", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(r.Context(),
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, $3)`,
		gameID, user.UserID, count+1)
	if err != nil {
		log.Printf("Join game error: %v", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": gameID, "inviteCode": code})
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")
	state, err := h.loadGameState(r.Context(), gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(r.Context(),
		`SELECT g.id, g.invite_code, g.status, COALESCE(m.name, ''), g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 LEFT JOIN missions m ON g.mission_id = m.id
		 JOIN game_players gp ON g.id = gp.game_id
		 WHERE gp.user_id = $1
		 ORDER BY g.created_at DESC
		 LIMIT 50`, user.UserID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	games := make([]models.GameSummary, 0)
	for rows.Next() {
		var g models.GameSummary
		if err := rows.Scan(&g.ID, &g.InviteCode, &g.Status, &g.MissionName, &g.CreatedAt, &g.CompletedAt, &g.WinnerID); err != nil {
			continue
		}

		// Fetch players for this game
		pRows, err := h.db.Query(r.Context(),
			`SELECT gp.user_id, u.discord_username, COALESCE(f.name, ''), gp.player_number,
			        gp.vp_primary + gp.vp_secondary + gp.vp_gambit + gp.vp_paint
			 FROM game_players gp
			 JOIN users u ON gp.user_id = u.id
			 LEFT JOIN factions f ON gp.faction_id = f.id
			 WHERE gp.game_id = $1
			 ORDER BY gp.player_number`, g.ID)
		if err == nil {
			for pRows.Next() {
				var p models.GamePlayerSummary
				pRows.Scan(&p.UserID, &p.Username, &p.FactionName, &p.PlayerNumber, &p.TotalVP)
				g.Players = append(g.Players, p)
			}
			pRows.Close()
		}

		games = append(games, g)
	}

	writeJSON(w, http.StatusOK, games)
}

func (h *GameHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(r.Context(),
		`SELECT g.id, g.invite_code, g.status, COALESCE(m.name, ''), g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 LEFT JOIN missions m ON g.mission_id = m.id
		 JOIN game_players gp ON g.id = gp.game_id
		 WHERE gp.user_id = $1 AND g.status = 'completed'
		 ORDER BY g.completed_at DESC
		 LIMIT 50`, user.UserID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	games := make([]models.GameSummary, 0)
	for rows.Next() {
		var g models.GameSummary
		if err := rows.Scan(&g.ID, &g.InviteCode, &g.Status, &g.MissionName, &g.CreatedAt, &g.CompletedAt, &g.WinnerID); err != nil {
			continue
		}

		pRows, err := h.db.Query(r.Context(),
			`SELECT gp.user_id, u.discord_username, COALESCE(f.name, ''), gp.player_number,
			        gp.vp_primary + gp.vp_secondary + gp.vp_gambit + gp.vp_paint
			 FROM game_players gp
			 JOIN users u ON gp.user_id = u.id
			 LEFT JOIN factions f ON gp.faction_id = f.id
			 WHERE gp.game_id = $1
			 ORDER BY gp.player_number`, g.ID)
		if err == nil {
			for pRows.Next() {
				var p models.GamePlayerSummary
				pRows.Scan(&p.UserID, &p.Username, &p.FactionName, &p.PlayerNumber, &p.TotalVP)
				g.Players = append(g.Players, p)
			}
			pRows.Close()
		}

		games = append(games, g)
	}

	writeJSON(w, http.StatusOK, games)
}

func (h *GameHandler) GetGameEvents(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

	rows, err := h.db.Query(r.Context(),
		`SELECT id, player_number, event_type, event_data, round, phase, created_at
		 FROM game_events WHERE game_id = $1 ORDER BY id`, gameID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]any
	for rows.Next() {
		var id int64
		var playerNum *int
		var eventType string
		var eventData json.RawMessage
		var round *int
		var phase *string
		var createdAt time.Time

		if err := rows.Scan(&id, &playerNum, &eventType, &eventData, &round, &phase, &createdAt); err != nil {
			continue
		}

		events = append(events, map[string]any{
			"id":           id,
			"playerNumber": playerNum,
			"eventType":    eventType,
			"eventData":    json.RawMessage(eventData),
			"round":        round,
			"phase":        phase,
			"createdAt":    createdAt,
		})
	}

	writeJSON(w, http.StatusOK, events)
}

// WebSocket upgrade handler
func (h *GameHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

	// Authenticate via query param
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(h.jwtSecret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Verify player is in this game
	var playerNumber int
	err = h.db.QueryRow(r.Context(),
		`SELECT gp.player_number FROM game_players gp
		 JOIN users u ON gp.user_id = u.id
		 WHERE gp.game_id = $1 AND u.id = $2`,
		gameID, claims.UserID,
	).Scan(&playerNumber)
	if err != nil {
		http.Error(w, "not a player in this game", http.StatusForbidden)
		return
	}

	// Load or create game engine
	state, err := h.loadGameState(r.Context(), gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	engine := game.NewEngine(state)
	room := h.hub.GetOrCreateRoom(gameID, engine)

	// Upgrade to WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}

	client := ws.NewClient(conn, room, claims.UserID, claims.Username, playerNumber)
	room.Register(client)

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)
}

func (h *GameHandler) loadGameState(ctx context.Context, gameID string) (*game.GameState, error) {
	var state game.GameState
	var missionPackID, missionID, missionName *string
	var firstTurnPlayer *int
	var winnerID *string
	var completedAt *time.Time

	err := h.db.QueryRow(ctx,
		`SELECT g.id, g.invite_code, g.status, g.current_round, g.current_phase,
		        g.active_player, g.first_turn_player, g.mission_pack_id, g.mission_id,
		        m.name, g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 LEFT JOIN missions m ON g.mission_id = m.id
		 WHERE g.id = $1`, gameID,
	).Scan(&state.GameID, &state.InviteCode, &state.Status, &state.CurrentRound,
		&state.CurrentPhase, &state.ActivePlayer, &firstTurnPlayer,
		&missionPackID, &missionID, &missionName,
		&state.CreatedAt, &completedAt, &winnerID)
	if err != nil {
		return nil, err
	}

	if firstTurnPlayer != nil {
		state.FirstTurnPlayer = *firstTurnPlayer
	}
	if missionPackID != nil {
		state.MissionPackID = *missionPackID
	}
	if missionID != nil {
		state.MissionID = *missionID
	}
	if missionName != nil {
		state.MissionName = *missionName
	}
	if completedAt != nil {
		state.CompletedAt = completedAt
	}
	if winnerID != nil {
		state.WinnerID = *winnerID
	}

	// Load players
	rows, err := h.db.Query(ctx,
		`SELECT gp.user_id, u.discord_username, gp.player_number,
		        COALESCE(gp.faction_id, ''), COALESCE(f.name, ''),
		        COALESCE(gp.detachment_id, ''), COALESCE(d.name, ''),
		        gp.cp, gp.vp_primary, gp.vp_secondary, gp.vp_gambit, gp.vp_paint,
		        gp.is_ready
		 FROM game_players gp
		 JOIN users u ON gp.user_id = u.id
		 LEFT JOIN factions f ON gp.faction_id = f.id
		 LEFT JOIN detachments d ON gp.detachment_id = d.id
		 WHERE gp.game_id = $1
		 ORDER BY gp.player_number`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p game.PlayerState
		if err := rows.Scan(&p.UserID, &p.Username, &p.PlayerNumber,
			&p.FactionID, &p.FactionName,
			&p.DetachmentID, &p.DetachmentName,
			&p.CP, &p.VPPrimary, &p.VPSecondary, &p.VPGambit, &p.VPPaint,
			&p.Ready); err != nil {
			continue
		}
		if p.PlayerNumber == 1 || p.PlayerNumber == 2 {
			state.Players[p.PlayerNumber-1] = &p
		}
	}

	return &state, nil
}

// PersistGameState saves the current game state back to the database
func (h *GameHandler) PersistGameState(state game.GameState, events []game.GameEvent) {
	ctx := context.Background()

	_, err := h.db.Exec(ctx,
		`UPDATE games SET status = $1, current_round = $2, current_phase = $3,
		 active_player = $4, first_turn_player = $5, mission_pack_id = NULLIF($6, ''),
		 mission_id = NULLIF($7::text, '')::uuid, completed_at = $8, winner_id = NULLIF($9::text, '')::uuid
		 WHERE id = $10`,
		state.Status, state.CurrentRound, state.CurrentPhase,
		state.ActivePlayer, state.FirstTurnPlayer, state.MissionPackID,
		state.MissionID, state.CompletedAt, state.WinnerID, state.GameID)
	if err != nil {
		log.Printf("Persist game state error: %v", err)
	}

	// Update players
	for _, p := range state.Players {
		if p == nil {
			continue
		}
		_, err := h.db.Exec(ctx,
			`UPDATE game_players SET faction_id = NULLIF($1, ''), detachment_id = NULLIF($2, ''),
			 cp = $3, vp_primary = $4, vp_secondary = $5, vp_gambit = $6, vp_paint = $7, is_ready = $8
			 WHERE game_id = $9 AND player_number = $10`,
			p.FactionID, p.DetachmentID, p.CP,
			p.VPPrimary, p.VPSecondary, p.VPGambit, p.VPPaint, p.Ready,
			state.GameID, p.PlayerNumber)
		if err != nil {
			log.Printf("Persist player state error: %v", err)
		}
	}

	// Persist events
	for _, e := range events {
		eventData, _ := json.Marshal(e.Data)
		_, err := h.db.Exec(ctx,
			`INSERT INTO game_events (game_id, player_number, event_type, event_data, round, phase)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			state.GameID, e.PlayerNumber, e.Type, eventData, e.Round, e.Phase)
		if err != nil {
			log.Printf("Persist event error: %v", err)
		}
	}
}
