package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGameWithTwoPlayers creates a game with two players and returns their IDs, tokens, and the game ID.
func setupGameWithTwoPlayers(t *testing.T) (user1ID, user2ID, token1, token2, gameID string) {
	t.Helper()
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID = testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID = testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ = testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)

	token1 = testutil.GenerateToken(t, user1ID, "player1")
	token2 = testutil.GenerateToken(t, user2ID, "player2")
	return user1ID, user2ID, token1, token2, gameID
}

// setupActiveGame creates a game in active state with both players having factions.
func setupActiveGame(t *testing.T) (user1ID, user2ID, token1, token2, gameID string) {
	t.Helper()
	env := testutil.SharedEnv
	user1ID, user2ID, token1, token2, gameID = setupGameWithTwoPlayers(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "NEC", "Necrons")
	testutil.SeedDetachment(t, env.Pool, "det-sm", "SM", "Gladius")
	testutil.SeedDetachment(t, env.Pool, "det-nec", "NEC", "Awakened Dynasty")

	_, err := env.Pool.Exec(context.Background(),
		`UPDATE games SET status = 'active', current_round = 1, current_phase = 'command', active_player = 1 WHERE id = $1`,
		gameID)
	require.NoError(t, err)
	_, err = env.Pool.Exec(context.Background(),
		`UPDATE game_players SET faction_id = 'SM', detachment_id = 'det-sm', is_ready = true WHERE game_id = $1 AND player_number = 1`,
		gameID)
	require.NoError(t, err)
	_, err = env.Pool.Exec(context.Background(),
		`UPDATE game_players SET faction_id = 'NEC', detachment_id = 'det-nec', is_ready = true WHERE game_id = $1 AND player_number = 2`,
		gameID)
	require.NoError(t, err)
	return user1ID, user2ID, token1, token2, gameID
}

func TestWSConnect_Success(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupGameWithTwoPlayers(t)

	conn := testutil.DialWS(t, env, gameID, token1)
	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "state_update", msg["type"])

	data, ok := msg["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, gameID, data["gameId"])
}

func TestWSConnect_NoToken(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupGameWithTwoPlayers(t)

	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, fmt.Sprintf("%s/ws/game/%s", wsURL, gameID), nil)
	require.Error(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestWSConnect_InvalidToken(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupGameWithTwoPlayers(t)

	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, fmt.Sprintf("%s/ws/game/%s?token=garbage", wsURL, gameID), nil)
	require.Error(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestWSConnect_NotAPlayer(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	outsider := testutil.CreateTestUser(t, env.Pool, "discord-3", "outsider")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)

	token := testutil.GenerateToken(t, outsider, "outsider")

	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, fmt.Sprintf("%s/ws/game/%s?token=%s", wsURL, gameID, token), nil)
	require.Error(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	}
}

func TestWSBothPlayersReceiveState(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	msg1 := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	assert.Equal(t, "state_update", msg1["type"])

	conn2 := testutil.DialWS(t, env, gameID, token2)
	msg2 := testutil.ReadWSMessage(t, conn2, 5*time.Second)
	assert.Equal(t, "state_update", msg2["type"])

	// Player 1 should receive player_connected for player 2
	p1Msg := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	assert.Equal(t, "player_connected", p1Msg["type"])
}

func TestWSJoinerLearnsAboutExistingPeers(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	// Player 1 connects first; reads their own initial state.
	conn1 := testutil.DialWS(t, env, gameID, token1)
	msg1 := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	require.Equal(t, "state_update", msg1["type"])

	// Player 2 joins after player 1 is already in the room. Player 2 should
	// receive a state_update followed by a player_connected for player 1, so
	// the joiner's UI knows the opponent is already present.
	conn2 := testutil.DialWS(t, env, gameID, token2)
	msg2 := testutil.ReadWSMessage(t, conn2, 5*time.Second)
	assert.Equal(t, "state_update", msg2["type"])

	peerMsg := testutil.ReadWSMessage(t, conn2, 5*time.Second)
	assert.Equal(t, "player_connected", peerMsg["type"])
	data, ok := peerMsg["data"].(map[string]interface{})
	require.True(t, ok)
	assert.EqualValues(t, 1, data["playerNumber"])
}

func TestWSPingPong(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupGameWithTwoPlayers(t)

	conn := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn, 5*time.Second) // initial state

	testutil.SendWSMessage(t, conn, map[string]string{"type": "ping"})

	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "pong", msg["type"])
}

func TestWSSyncRequest(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupGameWithTwoPlayers(t)

	conn := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn, 5*time.Second) // initial state

	testutil.SendWSMessage(t, conn, map[string]string{"type": "sync_request"})

	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "state_update", msg["type"])
}

func TestWSSelectFaction(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // initial state

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second) // initial state
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected for p2

	// Player 1 selects faction
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "select_faction",
			"factionId":   "SM",
			"factionName": "Space Marines",
		},
	})

	// Both should receive event + state_update
	event1 := testutil.DrainUntil(t, conn1, "event", 5*time.Second)
	assert.Equal(t, "event", event1["type"])

	state1 := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	assert.Equal(t, "state_update", state1["type"])

	event2 := testutil.DrainUntil(t, conn2, "event", 5*time.Second)
	assert.Equal(t, "event", event2["type"])

	state2 := testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)
	assert.Equal(t, "state_update", state2["type"])
}

func TestWSSetReady_GameStart(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "NEC", "Necrons")
	testutil.SeedDetachment(t, env.Pool, "det-sm", "SM", "Gladius")
	testutil.SeedDetachment(t, env.Pool, "det-nec", "NEC", "Awakened Dynasty")

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected

	// P1 select faction + detachment
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "select_faction", "factionId": "SM", "factionName": "Space Marines"},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "select_detachment", "detachmentId": "det-sm", "detachmentName": "Gladius"},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	// P2 select faction + detachment
	testutil.SendWSMessage(t, conn2, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "select_faction", "factionId": "NEC", "factionName": "Necrons"},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	testutil.SendWSMessage(t, conn2, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "select_detachment", "detachmentId": "det-nec", "detachmentName": "Awakened Dynasty"},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	// Pick who goes first (required before readying up)
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "select_first_turn_player", "firstTurnPlayer": 1},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	// Both ready
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "set_ready", "ready": true},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)

	testutil.SendWSMessage(t, conn2, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "set_ready", "ready": true},
	})

	// Final state should be active
	finalState := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	data := finalState["data"].(map[string]interface{})
	assert.Equal(t, "active", data["status"])
	assert.Equal(t, float64(1), data["currentRound"])
	assert.Equal(t, "command", data["currentPhase"])
}

func TestWSAdvancePhase(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "advance_phase"},
	})

	state := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	data := state["data"].(map[string]interface{})
	assert.Equal(t, "movement", data["currentPhase"])

	state2 := testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)
	data2 := state2["data"].(map[string]interface{})
	assert.Equal(t, "movement", data2["currentPhase"])
}

func TestWSAdvancePhase_WrongPlayer(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, token2, gameID := setupActiveGame(t)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)

	testutil.SendWSMessage(t, conn2, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "advance_phase"},
	})

	msg := testutil.ReadWSMessage(t, conn2, 5*time.Second)
	assert.Equal(t, "error", msg["type"])
}

func TestWSScoreVP(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "score_vp",
			"category":    "primary",
			"delta":       5,
			"scoringSlot": "end_of_command_phase",
		},
	})

	state := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	data := state["data"].(map[string]interface{})
	players := data["players"].([]interface{})
	p1 := players[0].(map[string]interface{})
	assert.Equal(t, float64(5), p1["vpPrimary"])
}

func TestWSUseStratagem_InsufficientCP(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":          "use_stratagem",
			"stratagemId":   "strat-1",
			"stratagemName": "Some Stratagem",
			"cpCost":        1,
		},
	})

	msg := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	assert.Equal(t, "error", msg["type"])
}

// TestWSUseStratagem_RejectsBoardingActions verifies defense-in-depth: even if
// a client submits a boarding-actions stratagem ID directly (bypassing the
// filtered list endpoints), the engine's DB lookup filters it out and the
// action is rejected.
func TestWSUseStratagem_RejectsBoardingActions(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	// Seed a boarding-actions stratagem and give player 1 enough CP to afford
	// it, so the only reason the action can fail is the game_mode filter.
	_, err := env.Pool.Exec(context.Background(),
		`INSERT INTO stratagems (id, faction_id, name, type, cp_cost, turn, phase, description, game_mode)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		"strat-ba", "SM", "Explosive Clearance",
		"Boarding Actions \u2013 Battle Tactic Stratagem", 1,
		"Your turn", "Shooting phase", "Boarding Actions stratagem.", "boarding_actions")
	require.NoError(t, err)
	_, err = env.Pool.Exec(context.Background(),
		`UPDATE game_players SET cp = 5 WHERE game_id = $1 AND player_number = 1`, gameID)
	require.NoError(t, err)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "use_stratagem",
			"stratagemId": "strat-ba",
			"cpCost":      1,
		},
	})

	msg := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	assert.Equal(t, "error", msg["type"])
}

func TestWSConcede(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "concede"},
	})

	state := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	data := state["data"].(map[string]interface{})
	assert.Equal(t, "completed", data["status"])
	assert.NotEmpty(t, data["winnerId"])

	state2 := testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)
	data2 := state2["data"].(map[string]interface{})
	assert.Equal(t, "completed", data2["status"])
}

func TestWSPersistence(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "score_vp",
			"category":    "primary",
			"delta":       10,
			"scoringSlot": "end_of_command_phase",
		},
	})
	testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)

	// Give persistence a moment to complete
	time.Sleep(200 * time.Millisecond)

	var vpPrimary int
	err := env.Pool.QueryRow(context.Background(),
		`SELECT vp_primary FROM game_players WHERE game_id = $1 AND player_number = 1`, gameID,
	).Scan(&vpPrimary)
	require.NoError(t, err)
	assert.Equal(t, 10, vpPrimary)

	var eventCount int
	require.NoError(t, env.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM game_events WHERE game_id = $1`, gameID,
	).Scan(&eventCount))
	assert.GreaterOrEqual(t, eventCount, 1)
}

// TestWSPersistence_PrimaryScoredSlots verifies that the per-rule scoring
// ledger survives a backend restart by round-tripping through the database.
// Without this, missions like Purge the Foe could lose their duplicate-score
// guards (and undo history) after any deploy or crash.
func TestWSPersistence_PrimaryScoredSlots(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	for _, label := range []string{"Destroyed 1+ enemy unit", "Control 1+ objective"} {
		testutil.SendWSMessage(t, conn1, map[string]interface{}{
			"type": "action",
			"data": map[string]interface{}{
				"type":             "score_vp",
				"category":         "primary",
				"delta":            4,
				"scoringSlot":      "end_of_battle_round",
				"scoringRuleLabel": label,
			},
		})
		testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	}

	time.Sleep(200 * time.Millisecond)

	var slotsJSON []byte
	err := env.Pool.QueryRow(context.Background(),
		`SELECT vp_primary_scored_slots FROM game_players WHERE game_id = $1 AND player_number = 1`, gameID,
	).Scan(&slotsJSON)
	require.NoError(t, err)

	var slots map[string]map[string]map[string]int
	require.NoError(t, json.Unmarshal(slotsJSON, &slots))

	rules := slots["1"]["end_of_battle_round"]
	assert.Equal(t, 4, rules["Destroyed 1+ enemy unit"])
	assert.Equal(t, 4, rules["Control 1+ objective"])
}

// TestWSLateJoin_Player2CanAct verifies that when player 2 connects to a room
// that was already created by player 1, player 2 is properly added to the engine
// state and can perform actions without getting "invalid player number".
func TestWSLateJoin_Player2CanAct(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	// Create game with only player 1 initially
	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)

	token1 := testutil.GenerateToken(t, user1ID, "player1")
	token2 := testutil.GenerateToken(t, user2ID, "player2")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")

	// Player 1 connects first — this creates the room with only player 1 in engine state
	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // initial state_update

	// Player 2 joins the game via REST
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)

	// Player 2 connects via WS — should be added to engine state by AddPlayer
	conn2 := testutil.DialWS(t, env, gameID, token2)
	p2State := testutil.ReadWSMessage(t, conn2, 5*time.Second) // initial state_update
	assert.Equal(t, "state_update", p2State["type"])

	// Verify player 2's state_update contains both players
	data := p2State["data"].(map[string]interface{})
	players := data["players"].([]interface{})
	assert.Equal(t, 2, len(players))
	// Both player slots should be non-nil
	assert.NotNil(t, players[0])
	assert.NotNil(t, players[1])

	// Player 1 should receive player_connected for player 2
	p1Msg := testutil.ReadWSMessage(t, conn1, 5*time.Second)
	assert.Equal(t, "player_connected", p1Msg["type"])

	// Drain the player_connected notification on conn1's state
	// Now player 2 selects a faction — this should NOT return "invalid player number"
	testutil.SendWSMessage(t, conn2, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "select_faction",
			"factionId":   "SM",
			"factionName": "Space Marines",
		},
	})

	// Player 2 should receive event + state_update, NOT an error
	event := testutil.DrainUntil(t, conn2, "event", 5*time.Second)
	assert.Equal(t, "event", event["type"])

	state := testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)
	assert.Equal(t, "state_update", state["type"])

	// Verify the faction was actually set on player 2
	stateData := state["data"].(map[string]interface{})
	statePlayers := stateData["players"].([]interface{})
	p2 := statePlayers[1].(map[string]interface{})
	assert.Equal(t, "SM", p2["factionId"])
	assert.Equal(t, "Space Marines", p2["factionName"])
}

// TestWSLateJoin_Player1CanActAfterPlayer2Joins verifies that player 1 can still
// perform actions after player 2 late-joins the room.
func TestWSLateJoin_Player1CanActAfterPlayer2Joins(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)

	token1 := testutil.GenerateToken(t, user1ID, "player1")
	token2 := testutil.GenerateToken(t, user2ID, "player2")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")

	// Player 1 connects first
	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // initial state

	// Player 2 joins and connects
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)
	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second) // initial state
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected

	// Player 1 selects faction — should still work fine
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "select_faction",
			"factionId":   "SM",
			"factionName": "Space Marines",
		},
	})

	event := testutil.DrainUntil(t, conn1, "event", 5*time.Second)
	assert.Equal(t, "event", event["type"])

	state := testutil.DrainUntil(t, conn1, "state_update", 5*time.Second)
	stateData := state["data"].(map[string]interface{})
	statePlayers := stateData["players"].([]interface{})
	p1 := statePlayers[0].(map[string]interface{})
	assert.Equal(t, "SM", p1["factionId"])

	// Player 2 should also receive the updates
	event2 := testutil.DrainUntil(t, conn2, "event", 5*time.Second)
	assert.Equal(t, "event", event2["type"])

	state2 := testutil.DrainUntil(t, conn2, "state_update", 5*time.Second)
	assert.Equal(t, "state_update", state2["type"])
}

func TestWSSpectator_ConnectActiveGame(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupActiveGame(t)

	conn := testutil.DialSpectatorWS(t, env, gameID)
	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "state_update", msg["type"])

	data, ok := msg["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, gameID, data["gameId"])
	assert.Equal(t, "active", data["status"])
}

func TestWSSpectator_RejectsSetupGame(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupGameWithTwoPlayers(t)

	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, fmt.Sprintf("%s/ws/game/%s/spectate", wsURL, gameID), nil)
	require.Error(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	}
}

func TestWSSpectator_UnknownGame(t *testing.T) {
	env := testutil.SharedEnv

	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, fmt.Sprintf("%s/ws/game/%s/spectate", wsURL, "00000000-0000-0000-0000-000000000000"), nil)
	require.Error(t, err)
	if resp != nil {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
}

func TestWSSpectator_ActionRejected(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupActiveGame(t)

	conn := testutil.DialSpectatorWS(t, env, gameID)
	testutil.ReadWSMessage(t, conn, 5*time.Second) // state_update

	testutil.SendWSMessage(t, conn, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{"type": "advance_phase"},
	})

	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "error", msg["type"])
	data, ok := msg["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", data["code"])
}

func TestWSSpectator_ReceivesLiveStateUpdates(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	spectator := testutil.DialSpectatorWS(t, env, gameID)
	testutil.ReadWSMessage(t, spectator, 5*time.Second) // initial state

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "score_vp",
			"category":    "primary",
			"delta":       5,
			"scoringSlot": "end_of_command_phase",
		},
	})

	state := testutil.DrainUntil(t, spectator, "state_update", 5*time.Second)
	data := state["data"].(map[string]interface{})
	players := data["players"].([]interface{})
	p1 := players[0].(map[string]interface{})
	assert.Equal(t, float64(5), p1["vpPrimary"])
}

func TestWSSpectator_DoesNotBroadcastPresenceToPlayers(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, _, gameID := setupActiveGame(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // initial state

	// Spectator joins; players must NOT receive a player_connected event for them.
	spectator := testutil.DialSpectatorWS(t, env, gameID)
	testutil.ReadWSMessage(t, spectator, 5*time.Second)

	// Trigger a real state-changing action so we have something to read.
	testutil.SendWSMessage(t, conn1, map[string]interface{}{
		"type": "action",
		"data": map[string]interface{}{
			"type":        "score_vp",
			"category":    "primary",
			"delta":       1,
			"scoringSlot": "end_of_command_phase",
		},
	})

	// The next non-pong message conn1 sees should be the event/state_update from
	// the score action — not a player_connected for the spectator.
	for i := 0; i < 5; i++ {
		msg := testutil.ReadWSMessage(t, conn1, 5*time.Second)
		require.NotEqual(t, "player_connected", msg["type"], "player must not see spectator presence")
		if msg["type"] == "state_update" {
			return
		}
	}
	t.Fatal("did not see expected state_update on player connection")
}

func TestWSSpectator_PingPong(t *testing.T) {
	env := testutil.SharedEnv
	_, _, _, _, gameID := setupActiveGame(t)

	conn := testutil.DialSpectatorWS(t, env, gameID)
	testutil.ReadWSMessage(t, conn, 5*time.Second) // state_update

	testutil.SendWSMessage(t, conn, map[string]string{"type": "ping"})
	msg := testutil.ReadWSMessage(t, conn, 5*time.Second)
	assert.Equal(t, "pong", msg["type"])
}

func TestWSPlayerDisconnect(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected for p2

	_ = conn1.Close(websocket.StatusNormalClosure, "leaving")

	msg := testutil.DrainUntil(t, conn2, "player_disconnected", 5*time.Second)
	assert.Equal(t, "player_disconnected", msg["type"])
}
