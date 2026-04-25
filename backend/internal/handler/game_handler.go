package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
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
	h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1 AND user_id = $2`,
		gameID, user.UserID,
	).Scan(&count)

	if count > 0 {
		out := &JoinGameOutput{}
		out.Body.ID = gameID
		out.Body.InviteCode = code
		return out, nil
	}

	// Check player count
	h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID,
	).Scan(&count)

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
		`SELECT g.id, g.invite_code, g.status, COALESCE(g.mission_name, ''), g.created_at, g.completed_at, g.winner_id
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

	return &GameListOutput{Body: games}, nil
}

func (h *GameHandler) GetHistory(ctx context.Context, input *HistoryInput) (*GameListOutput, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	query := `SELECT g.id, g.invite_code, g.status, COALESCE(g.mission_name, ''), g.created_at, g.completed_at, g.winner_id
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
		paramN++
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
		        gp.vp_primary + gp.vp_secondary + gp.vp_gambit + gp.vp_paint AS total_vp
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
		json.Unmarshal(eventData, &data)
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

	engine := game.NewEngine(state)
	engine.SetStratagemLookup(h.lookupStratagem)
	room := h.hub.GetOrCreateRoom(gameID, engine)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "WebSocket accept error", "error", err)
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
	var missionPackID, missionID, missionName, twistID, twistName *string
	var firstTurnPlayer *int
	var winnerID *string
	var completedAt *time.Time

	err := h.db.QueryRow(ctx,
		`SELECT g.id, g.invite_code, g.status, g.current_round, g.current_turn, g.current_phase,
		        g.active_player, g.first_turn_player, g.mission_pack_id, g.mission_id,
		        g.mission_name, g.twist_id, g.twist_name,
		        g.created_at, g.completed_at, g.winner_id
		 FROM games g
		 WHERE g.id = $1`, gameID,
	).Scan(&state.GameID, &state.InviteCode, &state.Status, &state.CurrentRound, &state.CurrentTurn,
		&state.CurrentPhase, &state.ActivePlayer, &firstTurnPlayer,
		&missionPackID, &missionID, &missionName, &twistID, &twistName,
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
	if twistID != nil {
		state.TwistID = *twistID
	}
	if twistName != nil {
		state.TwistName = *twistName
	}
	if completedAt != nil {
		state.CompletedAt = completedAt
	}
	if winnerID != nil {
		state.WinnerID = *winnerID
	}

	rows, err := h.db.Query(ctx,
		`SELECT gp.user_id, u.discord_username, gp.player_number,
		        COALESCE(gp.faction_id, ''), COALESCE(f.name, ''),
		        COALESCE(gp.detachment_id, ''), COALESCE(d.name, ''),
		        gp.cp, gp.vp_primary, gp.vp_secondary, gp.vp_gambit, gp.vp_paint,
		        gp.is_ready, gp.secondary_mode,
		        gp.tactical_deck, gp.active_secondaries, gp.achieved_secondaries, gp.discarded_secondaries,
		        gp.is_challenger, COALESCE(gp.challenger_card_id, ''), gp.adapt_or_die_uses
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
		var tacticalDeckJSON, activeSecJSON, achievedSecJSON, discardedSecJSON []byte
		if err := rows.Scan(&p.UserID, &p.Username, &p.PlayerNumber,
			&p.FactionID, &p.FactionName,
			&p.DetachmentID, &p.DetachmentName,
			&p.CP, &p.VPPrimary, &p.VPSecondary, &p.VPGambit, &p.VPPaint,
			&p.Ready, &p.SecondaryMode,
			&tacticalDeckJSON, &activeSecJSON, &achievedSecJSON, &discardedSecJSON,
			&p.IsChallenger, &p.ChallengerCardID, &p.AdaptOrDieUses); err != nil {
			slog.Error("Scan player error", "error", err)
			continue
		}

		json.Unmarshal(tacticalDeckJSON, &p.TacticalDeck)
		json.Unmarshal(activeSecJSON, &p.ActiveSecondaries)
		json.Unmarshal(achievedSecJSON, &p.AchievedSecondaries)
		json.Unmarshal(discardedSecJSON, &p.DiscardedSecondaries)

		if p.TacticalDeck == nil {
			p.TacticalDeck = []game.ActiveSecondary{}
		}
		if p.ActiveSecondaries == nil {
			p.ActiveSecondaries = []game.ActiveSecondary{}
		}
		if p.AchievedSecondaries == nil {
			p.AchievedSecondaries = []game.ActiveSecondary{}
		}
		if p.DiscardedSecondaries == nil {
			p.DiscardedSecondaries = []game.ActiveSecondary{}
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

	_, err := h.db.Exec(ctx,
		`UPDATE games SET status = $1, current_round = $2, current_turn = $3, current_phase = $4,
		 active_player = $5, first_turn_player = $6, mission_pack_id = NULLIF($7, ''),
		 mission_id = NULLIF($8, ''), mission_name = NULLIF($9, ''),
		 twist_id = NULLIF($10, ''), twist_name = NULLIF($11, ''),
		 completed_at = $12, winner_id = NULLIF($13::text, '')::uuid
		 WHERE id = $14`,
		state.Status, state.CurrentRound, state.CurrentTurn, state.CurrentPhase,
		state.ActivePlayer, state.FirstTurnPlayer, state.MissionPackID,
		state.MissionID, state.MissionName, state.TwistID, state.TwistName,
		state.CompletedAt, state.WinnerID, state.GameID)
	if err != nil {
		slog.Error("Persist game state error", "error", err)
	}

	for _, p := range state.Players {
		if p == nil {
			continue
		}
		tacticalDeckJSON, _ := json.Marshal(p.TacticalDeck)
		activeSecJSON, _ := json.Marshal(p.ActiveSecondaries)
		achievedSecJSON, _ := json.Marshal(p.AchievedSecondaries)
		discardedSecJSON, _ := json.Marshal(p.DiscardedSecondaries)

		_, err := h.db.Exec(ctx,
			`UPDATE game_players SET faction_id = NULLIF($1, ''), detachment_id = NULLIF($2, ''),
			 cp = $3, vp_primary = $4, vp_secondary = $5, vp_gambit = $6, vp_paint = $7, is_ready = $8,
			 secondary_mode = $9, tactical_deck = $10, active_secondaries = $11,
			 achieved_secondaries = $12, discarded_secondaries = $13,
			 is_challenger = $14, challenger_card_id = NULLIF($15, ''), adapt_or_die_uses = $16
			 WHERE game_id = $17 AND player_number = $18`,
			p.FactionID, p.DetachmentID, p.CP,
			p.VPPrimary, p.VPSecondary, p.VPGambit, p.VPPaint, p.Ready,
			p.SecondaryMode, tacticalDeckJSON, activeSecJSON,
			achievedSecJSON, discardedSecJSON,
			p.IsChallenger, p.ChallengerCardID, p.AdaptOrDieUses,
			state.GameID, p.PlayerNumber)
		if err != nil {
			slog.Error("Persist player state error", "error", err)
		}
	}

	for _, e := range events {
		eventData, _ := json.Marshal(e.Data)
		_, err := h.db.Exec(ctx,
			`INSERT INTO game_events (game_id, player_number, event_type, event_data, round, phase)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			state.GameID, e.PlayerNumber, e.Type, eventData, e.Round, e.Phase)
		if err != nil {
			slog.Error("Persist event error", "error", err)
		}
	}
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
