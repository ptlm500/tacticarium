package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/game"
	"github.com/peter/tacticarium/backend/internal/game/scoring"
	"github.com/peter/tacticarium/backend/internal/models"
	"github.com/peter/tacticarium/backend/internal/ws"
	"github.com/peter/tacticarium/backend/pkg/invite"
)

type GameHandler struct {
	db        *pgxpool.Pool
	hub       *ws.Hub
	jwtSecret string
}

func NewGameHandler(db *pgxpool.Pool, hub *ws.Hub, jwtSecret string) *GameHandler {
	return &GameHandler{
		db:        db,
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

func (h *GameHandler) CreateGame(ctx context.Context, input *struct{}) (*CreateGameOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	code := invite.GenerateCode(6)

	var gameID string
	err := h.db.QueryRow(ctx,
		`INSERT INTO games (invite_code) VALUES ($1) RETURNING id`, code,
	).Scan(&gameID)
	if err != nil {
		slog.ErrorContext(ctx, "Create game error", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	_, err = h.db.Exec(ctx,
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, 1)`,
		gameID, user.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "Add player error", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	out := &CreateGameOutput{}
	out.Body.ID = gameID
	out.Body.InviteCode = code
	return out, nil
}

func (h *GameHandler) JoinGame(ctx context.Context, input *JoinGameInput) (*JoinGameOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	code := input.Code

	var gameID, status string
	err := h.db.QueryRow(ctx,
		`SELECT id, status FROM games WHERE invite_code = $1`, code,
	).Scan(&gameID, &status)
	if err != nil {
		return nil, huma.Error404NotFound("game not found")
	}

	if status != "setup" {
		return nil, huma.Error400BadRequest("game already started")
	}

	// Check if already in game
	var count int
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1 AND user_id = $2`,
		gameID, user.UserID,
	).Scan(&count); err != nil {
		slog.ErrorContext(ctx, "Join game membership check failed", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	if count > 0 {
		out := &JoinGameOutput{}
		out.Body.ID = gameID
		out.Body.InviteCode = code
		return out, nil
	}

	// Check player count
	if err := h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID,
	).Scan(&count); err != nil {
		slog.ErrorContext(ctx, "Join game player count check failed", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	if count >= 2 {
		return nil, huma.Error400BadRequest("game is full")
	}

	_, err = h.db.Exec(ctx,
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, $3)`,
		gameID, user.UserID, count+1)
	if err != nil {
		slog.ErrorContext(ctx, "Join game error", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	out := &JoinGameOutput{}
	out.Body.ID = gameID
	out.Body.InviteCode = code
	return out, nil
}

func (h *GameHandler) GetGame(ctx context.Context, input *GameIDParam) (*GameStateOutput, error) {
	state, err := h.loadGameState(ctx, input.GameID)
	if err != nil {
		return nil, huma.Error404NotFound("game not found")
	}
	return &GameStateOutput{Body: state}, nil
}

func (h *GameHandler) ListGames(ctx context.Context, input *struct{}) (*GameListOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	rows, err := h.db.Query(ctx,
		`SELECT g.id, g.invite_code, g.status, '', g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 JOIN game_players gp ON g.id = gp.game_id
		 WHERE gp.user_id = $1 AND gp.hidden_at IS NULL
		 ORDER BY g.created_at DESC
		 LIMIT 50`, user.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	games := make([]models.GameSummary, 0)
	for rows.Next() {
		var g models.GameSummary
		if err := rows.Scan(&g.ID, &g.InviteCode, &g.Status, &g.MissionName, &g.CreatedAt, &g.CompletedAt, &g.WinnerID); err != nil {
			continue
		}

		pRows, err := h.db.Query(ctx,
			`SELECT gp.user_id, u.discord_username, u.discord_id, u.discord_avatar,
			        COALESCE(f.name, ''), gp.player_number,
			        gp.vp_primary + gp.vp_secondary + gp.vp_paint
			 FROM game_players gp
			 JOIN users u ON gp.user_id = u.id
			 LEFT JOIN factions f ON gp.faction_id = f.id
			 WHERE gp.game_id = $1
			 ORDER BY gp.player_number`, g.ID)
		if err == nil {
			for pRows.Next() {
				var p models.GamePlayerSummary
				var discordID string
				var discordAvatar *string
				_ = pRows.Scan(&p.UserID, &p.Username, &discordID, &discordAvatar,
					&p.FactionName, &p.PlayerNumber, &p.TotalVP)
				p.AvatarURL = auth.AvatarURL(discordID, discordAvatar)
				g.Players = append(g.Players, p)
			}
			pRows.Close()
		}

		games = append(games, g)
	}

	return &GameListOutput{Body: games}, nil
}

func (h *GameHandler) GetHistory(ctx context.Context, input *HistoryInput) (*GameListOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	query := `SELECT g.id, g.invite_code, g.status, '', g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 JOIN game_players gp ON g.id = gp.game_id
		 WHERE gp.user_id = $1 AND gp.hidden_at IS NULL AND g.status IN ('completed', 'abandoned')`
	args := []any{user.UserID}
	paramN := 2

	if input.MyFaction != "" {
		query += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM game_players mp
			JOIN factions mf ON mp.faction_id = mf.id
			WHERE mp.game_id = g.id AND mp.user_id = $1 AND mf.name = $%d
		)`, paramN)
		args = append(args, input.MyFaction)
		paramN++
	}

	if input.OpponentFaction != "" {
		query += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM game_players op
			JOIN factions of2 ON op.faction_id = of2.id
			WHERE op.game_id = g.id AND op.user_id != $1 AND of2.name = $%d
		)`, paramN)
		args = append(args, input.OpponentFaction)
	}

	query += ` ORDER BY g.completed_at DESC LIMIT 50`

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	games := make([]models.GameSummary, 0)
	for rows.Next() {
		var g models.GameSummary
		if err := rows.Scan(&g.ID, &g.InviteCode, &g.Status, &g.MissionName, &g.CreatedAt, &g.CompletedAt, &g.WinnerID); err != nil {
			continue
		}

		pRows, err := h.db.Query(ctx,
			`SELECT gp.user_id, u.discord_username, u.discord_id, u.discord_avatar,
			        COALESCE(f.name, ''), gp.player_number,
			        gp.vp_primary + gp.vp_secondary + gp.vp_paint
			 FROM game_players gp
			 JOIN users u ON gp.user_id = u.id
			 LEFT JOIN factions f ON gp.faction_id = f.id
			 WHERE gp.game_id = $1
			 ORDER BY gp.player_number`, g.ID)
		if err == nil {
			for pRows.Next() {
				var p models.GamePlayerSummary
				var discordID string
				var discordAvatar *string
				_ = pRows.Scan(&p.UserID, &p.Username, &discordID, &discordAvatar,
					&p.FactionName, &p.PlayerNumber, &p.TotalVP)
				p.AvatarURL = auth.AvatarURL(discordID, discordAvatar)
				g.Players = append(g.Players, p)
			}
			pRows.Close()
		}

		games = append(games, g)
	}

	return &GameListOutput{Body: games}, nil
}

func (h *GameHandler) GetStats(ctx context.Context, input *struct{}) (*StatsOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	stats := UserStats{
		FactionStats: make([]FactionStat, 0),
	}

	// Win/loss/draw/abandoned counts + average VP
	rows, err := h.db.Query(ctx,
		`SELECT g.status, g.winner_id,
		        gp.vp_primary + gp.vp_secondary + gp.vp_paint AS total_vp
		 FROM games g
		 JOIN game_players gp ON g.id = gp.game_id AND gp.user_id = $1
		 WHERE g.status IN ('completed', 'abandoned')`, user.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	var totalVP int
	var gameCount int
	for rows.Next() {
		var status string
		var winnerID *string
		var vp int
		if err := rows.Scan(&status, &winnerID, &vp); err != nil {
			continue
		}
		gameCount++
		totalVP += vp

		if status == "abandoned" {
			stats.Abandoned++
		} else if winnerID == nil {
			stats.Draws++
		} else if *winnerID == user.UserID {
			stats.Wins++
		} else {
			stats.Losses++
		}
	}

	if gameCount > 0 {
		stats.AverageVP = float64(totalVP) / float64(gameCount)
	}

	// Faction stats
	fRows, err := h.db.Query(ctx,
		`SELECT COALESCE(f.name, 'Unknown') AS faction_name,
		        COUNT(*) AS games_played,
		        COUNT(*) FILTER (WHERE g.winner_id = $1) AS wins
		 FROM game_players gp
		 JOIN games g ON gp.game_id = g.id
		 LEFT JOIN factions f ON gp.faction_id = f.id
		 WHERE gp.user_id = $1 AND g.status IN ('completed', 'abandoned')
		   AND gp.faction_id IS NOT NULL
		 GROUP BY f.name
		 ORDER BY games_played DESC`, user.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer fRows.Close()

	for fRows.Next() {
		var fs FactionStat
		if err := fRows.Scan(&fs.FactionName, &fs.GamesPlayed, &fs.Wins); err != nil {
			continue
		}
		stats.FactionStats = append(stats.FactionStats, fs)
	}

	return &StatsOutput{Body: stats}, nil
}

func (h *GameHandler) HideGame(ctx context.Context, input *GameIDParam) (*struct{}, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	result, err := h.db.Exec(ctx,
		`UPDATE game_players SET hidden_at = NOW()
		 WHERE game_id = $1 AND user_id = $2 AND hidden_at IS NULL`,
		input.GameID, user.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "Hide game error", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	if result.RowsAffected() == 0 {
		return nil, huma.Error404NotFound("game not found or already hidden")
	}

	return nil, nil
}

func (h *GameHandler) GetGameEvents(ctx context.Context, input *GameIDParam) (*GameEventsOutput, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, player_number, event_type, event_data, round, phase, created_at
		 FROM game_events WHERE game_id = $1 ORDER BY id`, input.GameID)
	if err != nil {
		return nil, huma.Error500InternalServerError("database error")
	}
	defer rows.Close()

	events := make([]GameEvent, 0)
	for rows.Next() {
		var ev GameEvent
		var eventData json.RawMessage

		if err := rows.Scan(&ev.ID, &ev.PlayerNumber, &ev.EventType, &eventData, &ev.Round, &ev.Phase, &ev.CreatedAt); err != nil {
			continue
		}

		var data any
		_ = json.Unmarshal(eventData, &data)
		ev.EventData = data

		events = append(events, ev)
	}

	return &GameEventsOutput{Body: events}, nil
}

// HandleWebSocket stays as a raw chi handler (WebSocket upgrade).
func (h *GameHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

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

	state, err := h.loadGameState(r.Context(), gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	avatarURL := ""
	if p := state.Players[playerNumber-1]; p != nil {
		avatarURL = p.AvatarURL
	}

	engine := game.NewEngine(state)
	engine.SetStratagemLookup(h.lookupStratagem)
	engine.SetMissionResolver(h.resolveMission)
	engine.SetBoardBuilder(h.buildBoard)
	engine.SetSecondaryDeckBuilder(h.buildSecondaryDeck)
	room := h.hub.GetOrCreateRoom(gameID, engine)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "WebSocket accept error", "error", err)
		return
	}

	client := ws.NewClient(conn, room, claims.UserID, claims.Username, avatarURL, playerNumber)
	room.Register(client)

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)
}

// HandleSpectatorWebSocket accepts a public, read-only WebSocket connection.
// Spectators do not authenticate and cannot send actions; they receive state
// updates and events for the duration of an active game.
func (h *GameHandler) HandleSpectatorWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameId")

	state, err := h.loadGameState(r.Context(), gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	if state.Status != "active" {
		http.Error(w, "game is not active", http.StatusForbidden)
		return
	}

	engine := game.NewEngine(state)
	engine.SetStratagemLookup(h.lookupStratagem)
	engine.SetMissionResolver(h.resolveMission)
	engine.SetBoardBuilder(h.buildBoard)
	engine.SetSecondaryDeckBuilder(h.buildSecondaryDeck)
	room := h.hub.GetOrCreateRoom(gameID, engine)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Spectator WebSocket accept error", "error", err)
		return
	}

	client := ws.NewSpectatorClient(conn, room)
	room.Register(client)

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)
}

func (h *GameHandler) loadGameState(ctx context.Context, gameID string) (*game.GameState, error) {
	var state game.GameState
	var firstTurnPlayer *int
	var winnerID *string
	var completedAt *time.Time
	var boardJSON, startCtrlJSON []byte

	err := h.db.QueryRow(ctx,
		`SELECT g.id, g.invite_code, g.status, g.current_round, g.current_turn, g.current_phase,
		        g.active_player, g.first_turn_player, g.board, g.vp_per_game_cap, g.vp_per_round_cap,
		        g.start_of_turn_control, g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 WHERE g.id = $1`, gameID,
	).Scan(&state.GameID, &state.InviteCode, &state.Status, &state.CurrentRound, &state.CurrentTurn,
		&state.CurrentPhase, &state.ActivePlayer, &firstTurnPlayer,
		&boardJSON, &state.VPPerGameCap, &state.VPPerRoundCap, &startCtrlJSON,
		&state.CreatedAt, &completedAt, &winnerID)
	if err != nil {
		return nil, err
	}

	if firstTurnPlayer != nil {
		state.FirstTurnPlayer = *firstTurnPlayer
	}
	if completedAt != nil {
		state.CompletedAt = completedAt
	}
	if winnerID != nil {
		state.WinnerID = *winnerID
	}
	_ = json.Unmarshal(boardJSON, &state.Board)
	_ = json.Unmarshal(startCtrlJSON, &state.StartOfTurnControl)

	rows, err := h.db.Query(ctx,
		`SELECT gp.user_id, u.discord_username, u.discord_id, u.discord_avatar, gp.player_number,
		        COALESCE(gp.faction_id, ''), COALESCE(f.name, ''),
		        COALESCE(gp.detachment_id, ''), COALESCE(d.name, ''),
		        COALESCE(gp.side, ''), COALESCE(gp.force_disposition, ''), COALESCE(gp.force_disposition_name, ''),
		        COALESCE(gp.mission_id, ''), COALESCE(gp.mission_name, ''),
		        gp.cp, gp.vp_primary, gp.vp_secondary, gp.vp_paint, gp.is_ready, gp.secondary_mode,
		        gp.primary_card, gp.secondary_deck, gp.secondary_hand, gp.secondary_scored,
		        gp.primary_scored_this_round, gp.secondary_scored_this_round, gp.pending_score_prompts
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
		var discordID, side string
		var discordAvatar *string
		var primaryCardJSON, deckJSON, handJSON, scoredJSON, promptsJSON []byte
		if err := rows.Scan(&p.UserID, &p.Username, &discordID, &discordAvatar, &p.PlayerNumber,
			&p.FactionID, &p.FactionName,
			&p.DetachmentID, &p.DetachmentName,
			&side, &p.ForceDisposition, &p.ForceDispositionName,
			&p.MissionID, &p.MissionName,
			&p.CP, &p.VPPrimary, &p.VPSecondary, &p.VPPaint, &p.Ready, &p.SecondaryMode,
			&primaryCardJSON, &deckJSON, &handJSON, &scoredJSON,
			&p.PrimaryScoredThisRound, &p.SecondaryScoredThisRound, &promptsJSON); err != nil {
			slog.Error("Scan player error", "error", err)
			continue
		}
		p.AvatarURL = auth.AvatarURL(discordID, discordAvatar)
		p.Side = scoring.Side(side)

		_ = json.Unmarshal(primaryCardJSON, &p.PrimaryCard)
		_ = json.Unmarshal(deckJSON, &p.SecondaryDeck)
		_ = json.Unmarshal(handJSON, &p.SecondaryHand)
		_ = json.Unmarshal(scoredJSON, &p.SecondaryScored)
		_ = json.Unmarshal(promptsJSON, &p.PendingScorePrompts)

		if p.SecondaryDeck == nil {
			p.SecondaryDeck = []game.SecondaryCard{}
		}
		if p.SecondaryHand == nil {
			p.SecondaryHand = []game.SecondaryCard{}
		}
		if p.SecondaryScored == nil {
			p.SecondaryScored = []game.SecondaryCard{}
		}
		if p.PendingScorePrompts == nil {
			p.PendingScorePrompts = []game.ScorePrompt{}
		}

		if p.PlayerNumber == 1 || p.PlayerNumber == 2 {
			state.Players[p.PlayerNumber-1] = &p
		}
	}

	return &state, nil
}

// PersistGameState saves the current game state back to the database.
func (h *GameHandler) PersistGameState(state game.GameState, events []game.GameEvent) {
	ctx := context.Background()

	boardJSON, _ := json.Marshal(state.Board)
	startCtrl := state.StartOfTurnControl
	if startCtrl == nil {
		startCtrl = map[int]int{}
	}
	startCtrlJSON, _ := json.Marshal(startCtrl)

	_, err := h.db.Exec(ctx,
		`UPDATE games SET status = $1, current_round = $2, current_turn = $3, current_phase = $4,
		 active_player = $5, first_turn_player = $6, board = $7, vp_per_game_cap = $8,
		 vp_per_round_cap = $9, start_of_turn_control = $10,
		 completed_at = $11, winner_id = NULLIF($12::text, '')::uuid
		 WHERE id = $13`,
		state.Status, state.CurrentRound, state.CurrentTurn, state.CurrentPhase,
		state.ActivePlayer, state.FirstTurnPlayer, boardJSON,
		capOr(state.VPPerGameCap, game.DefaultVPPerGameCap),
		capOr(state.VPPerRoundCap, game.DefaultVPPerRoundCap), startCtrlJSON,
		state.CompletedAt, state.WinnerID, state.GameID)
	if err != nil {
		slog.Error("Persist game state error", "error", err)
	}

	for _, p := range state.Players {
		if p == nil {
			continue
		}
		primaryCardJSON, _ := json.Marshal(p.PrimaryCard)
		deckJSON, _ := json.Marshal(orEmptyCards(p.SecondaryDeck))
		handJSON, _ := json.Marshal(orEmptyCards(p.SecondaryHand))
		scoredJSON, _ := json.Marshal(orEmptyCards(p.SecondaryScored))
		promptsJSON, _ := json.Marshal(orEmptyPrompts(p.PendingScorePrompts))

		_, err := h.db.Exec(ctx,
			`UPDATE game_players SET faction_id = NULLIF($1, ''), detachment_id = NULLIF($2, ''),
			 side = NULLIF($3, ''), force_disposition = NULLIF($4, ''), force_disposition_name = NULLIF($5, ''),
			 mission_id = NULLIF($6, ''), mission_name = NULLIF($7, ''),
			 cp = $8, vp_primary = $9, vp_secondary = $10, vp_paint = $11, is_ready = $12,
			 secondary_mode = $13, primary_card = $14, secondary_deck = $15, secondary_hand = $16,
			 secondary_scored = $17, primary_scored_this_round = $18, secondary_scored_this_round = $19,
			 pending_score_prompts = $20
			 WHERE game_id = $21 AND player_number = $22`,
			p.FactionID, p.DetachmentID, string(p.Side), p.ForceDisposition, p.ForceDispositionName,
			p.MissionID, p.MissionName,
			p.CP, p.VPPrimary, p.VPSecondary, p.VPPaint, p.Ready,
			p.SecondaryMode, primaryCardJSON, deckJSON, handJSON,
			scoredJSON, p.PrimaryScoredThisRound, p.SecondaryScoredThisRound,
			promptsJSON,
			state.GameID, p.PlayerNumber)
		if err != nil {
			slog.Error("Persist player state error", "error", err)
		}
	}

	for i := range events {
		e := &events[i]
		eventData, _ := json.Marshal(e.Data)
		err := h.db.QueryRow(ctx,
			`INSERT INTO game_events (game_id, player_number, event_type, event_data, round, phase)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			state.GameID, e.PlayerNumber, e.Type, eventData, e.Round, e.Phase,
		).Scan(&e.ID)
		if err != nil {
			slog.Error("Persist event error", "error", err)
		}
	}
}

func capOr(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

func orEmptyCards(c []game.SecondaryCard) []game.SecondaryCard {
	if c == nil {
		return []game.SecondaryCard{}
	}
	return c
}

func orEmptyPrompts(p []game.ScorePrompt) []game.ScorePrompt {
	if p == nil {
		return []game.ScorePrompt{}
	}
	return p
}

// resolveMission maps a force-disposition matchup to the resolved primary
// mission (caps + scoring card) for a player.
func (h *GameHandler) resolveMission(disposition, opponentDisposition string) (game.ResolvedMission, bool) {
	ctx := context.Background()
	var missionID string
	if err := h.db.QueryRow(ctx,
		`SELECT mission_id FROM mission_matchups WHERE disposition = $1 AND opponent_disposition = $2`,
		disposition, opponentDisposition,
	).Scan(&missionID); err != nil {
		return game.ResolvedMission{}, false
	}

	rm := game.ResolvedMission{ID: missionID}
	_ = h.db.QueryRow(ctx,
		`SELECT name, vp_per_game_cap, vp_per_round_cap FROM primary_missions WHERE id = $1`, missionID,
	).Scan(&rm.Name, &rm.GameCap, &rm.RoundCap)

	var cardName, cardType, text string
	var awardsJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT name, card_type, text, awards FROM cards WHERE id = $1`, missionID,
	).Scan(&cardName, &cardType, &text, &awardsJSON); err == nil {
		if card, err := scoring.ParseCard(missionID, cardName, cardType, text, awardsJSON); err == nil {
			rm.PrimaryCard = card
		}
	}
	return rm, true
}

// buildBoard builds the board for the game's deployment pattern and player sides.
func (h *GameHandler) buildBoard(player1Side, player2Side scoring.Side) (scoring.Board, error) {
	ctx := context.Background()
	var id string
	var objJSON, zonesJSON, terrJSON []byte
	if err := h.db.QueryRow(ctx,
		`SELECT id, objectives, zones, territories FROM deployment_patterns ORDER BY id LIMIT 1`,
	).Scan(&id, &objJSON, &zonesJSON, &terrJSON); err != nil {
		return scoring.Board{}, err
	}
	dp, err := scoring.ParseDeploymentPattern(id, objJSON, zonesJSON, terrJSON)
	if err != nil {
		return scoring.Board{}, err
	}
	return scoring.BuildBoard(dp, player1Side, player2Side), nil
}

// buildSecondaryDeck loads the secondary card deck. (Mode-specific selection is
// refined in the setup UI work; for now both modes receive the full deck.)
func (h *GameHandler) buildSecondaryDeck(mode string) []game.SecondaryCard {
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, card_type, text, awards FROM cards WHERE card_type = 'secondary' ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var deck []game.SecondaryCard
	for rows.Next() {
		var id, name, cardType, text string
		var awardsJSON []byte
		if err := rows.Scan(&id, &name, &cardType, &text, &awardsJSON); err != nil {
			continue
		}
		card, err := scoring.ParseCard(id, name, cardType, text, awardsJSON)
		if err != nil {
			continue
		}
		deck = append(deck, game.SecondaryCard{Card: card})
	}
	return deck
}

func (h *GameHandler) lookupStratagem(id string) (*game.StratagemInfo, error) {
	var info game.StratagemInfo
	err := h.db.QueryRow(context.Background(),
		// Defense in depth: the player-facing list endpoints already hide
		// alternate game-mode content, so boarding-actions stratagem IDs should
		// never reach a client. Filter here too so a crafted action can't bypass
		// the exclusion.
		`SELECT name, cp_cost FROM stratagems WHERE id = $1 AND game_mode = 'core'`, id,
	).Scan(&info.Name, &info.CPCost)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
