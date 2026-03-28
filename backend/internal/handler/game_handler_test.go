package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "POST", "/api/games", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var body map[string]string
	testutil.ReadJSON(t, resp, &body)

	assert.NotEmpty(t, body["id"])
	assert.Len(t, body["inviteCode"], 6)

	var count int
	env.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1 AND user_id = $2 AND player_number = 1`,
		body["id"], userID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestCreateGame_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	resp := testutil.DoRequest(t, env, "POST", "/api/games", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestJoinGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, inviteCode := testutil.CreateTestGame(t, env.Pool, user1ID)

	token2 := testutil.GenerateToken(t, user2ID, "player2")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/"+inviteCode, nil, testutil.AuthHeader(token2))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	testutil.ReadJSON(t, resp, &body)
	assert.Equal(t, gameID, body["id"])

	var count int
	env.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID).Scan(&count)
	assert.Equal(t, 2, count)
}

func TestJoinGame_AlreadyInGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	_, inviteCode := testutil.CreateTestGame(t, env.Pool, userID)

	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/"+inviteCode, nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestJoinGame_GameFull(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	user3ID := testutil.CreateTestUser(t, env.Pool, "discord-3", "player3")

	gameID, inviteCode := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)

	token3 := testutil.GenerateToken(t, user3ID, "player3")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/"+inviteCode, nil, testutil.AuthHeader(token3))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestJoinGame_InvalidCode(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/XXXXXX", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestJoinGame_GameAlreadyStarted(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, inviteCode := testutil.CreateTestGame(t, env.Pool, user1ID)

	env.Pool.Exec(context.Background(), `UPDATE games SET status = 'active' WHERE id = $1`, gameID)

	token2 := testutil.GenerateToken(t, user2ID, "player2")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/"+inviteCode, nil, testutil.AuthHeader(token2))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestGetGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)

	token := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/games/"+gameID, nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	testutil.ReadJSON(t, resp, &body)

	assert.Equal(t, gameID, body["gameId"])
	assert.Equal(t, "setup", body["status"])
	players := body["players"].([]interface{})
	assert.Len(t, players, 2)
}

func TestGetGame_NotFound(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/games/00000000-0000-0000-0000-000000000000", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestListGames(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")

	testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.CreateTestGame(t, env.Pool, user2ID)

	token1 := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/games", nil, testutil.AuthHeader(token1))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 2)
}

func TestListGames_Empty(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/games", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 0)
}

func TestGetHistory_Empty(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/history", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 0)
}

func TestGetGameEvents(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, userID)
	token := testutil.GenerateToken(t, userID, "player1")

	eventData, _ := json.Marshal(map[string]string{"test": "data"})
	env.Pool.Exec(context.Background(),
		`INSERT INTO game_events (game_id, player_number, event_type, event_data, round, phase)
		 VALUES ($1, 1, 'test_event', $2, 1, 'command')`,
		gameID, eventData)
	env.Pool.Exec(context.Background(),
		`INSERT INTO game_events (game_id, player_number, event_type, event_data, round, phase)
		 VALUES ($1, 2, 'test_event_2', $2, 1, 'movement')`,
		gameID, eventData)

	resp := testutil.DoRequest(t, env, "GET", "/api/games/"+gameID+"/events", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events []map[string]interface{}
	testutil.ReadJSON(t, resp, &events)
	assert.Len(t, events, 2)
	assert.Equal(t, "test_event", events[0]["eventType"])
	assert.Equal(t, "test_event_2", events[1]["eventType"])
}
