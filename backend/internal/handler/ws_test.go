package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"
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
	return
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

	env.Pool.Exec(context.Background(),
		`UPDATE games SET status = 'active', current_round = 1, current_phase = 'command', active_player = 1 WHERE id = $1`,
		gameID)
	env.Pool.Exec(context.Background(),
		`UPDATE game_players SET faction_id = 'SM', detachment_id = 'det-sm', is_ready = true WHERE game_id = $1 AND player_number = 1`,
		gameID)
	env.Pool.Exec(context.Background(),
		`UPDATE game_players SET faction_id = 'NEC', detachment_id = 'det-nec', is_ready = true WHERE game_id = $1 AND player_number = 2`,
		gameID)
	return
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
			"type":     "score_vp",
			"category": "primary",
			"delta":    5,
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
			"type":     "score_vp",
			"category": "primary",
			"delta":    10,
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
	env.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM game_events WHERE game_id = $1`, gameID,
	).Scan(&eventCount)
	assert.GreaterOrEqual(t, eventCount, 1)
}

func TestWSPlayerDisconnect(t *testing.T) {
	env := testutil.SharedEnv
	_, _, token1, token2, gameID := setupGameWithTwoPlayers(t)

	conn1 := testutil.DialWS(t, env, gameID, token1)
	testutil.ReadWSMessage(t, conn1, 5*time.Second)

	conn2 := testutil.DialWS(t, env, gameID, token2)
	testutil.ReadWSMessage(t, conn2, 5*time.Second)
	testutil.ReadWSMessage(t, conn1, 5*time.Second) // player_connected for p2

	conn1.Close(websocket.StatusNormalClosure, "leaving")

	msg := testutil.DrainUntil(t, conn2, "player_disconnected", 5*time.Second)
	assert.Equal(t, "player_disconnected", msg["type"])
}
