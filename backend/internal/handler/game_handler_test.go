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
	testutil.AssertProblemDetails(t, resp, http.StatusUnauthorized)
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
	pd := testutil.AssertProblemDetails(t, resp, http.StatusBadRequest)
	assert.Contains(t, pd.Detail, "full")
}

func TestJoinGame_InvalidCode(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/join/XXXXXX", nil, testutil.AuthHeader(token))
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
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
	pd := testutil.AssertProblemDetails(t, resp, http.StatusBadRequest)
	assert.Contains(t, pd.Detail, "started")
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
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
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

// --- Hide Game Tests ---

func TestHideGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, userID)
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Game should no longer appear in list
	resp = testutil.DoRequest(t, env, "GET", "/api/games", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 0)
}

func TestHideGame_OnlyHidesForRequestingUser(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)

	token1 := testutil.GenerateToken(t, user1ID, "player1")
	token2 := testutil.GenerateToken(t, user2ID, "player2")

	// Player 1 hides the game
	resp := testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token1))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Player 1 should not see it
	resp = testutil.DoRequest(t, env, "GET", "/api/games", nil, testutil.AuthHeader(token1))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var games1 []map[string]interface{}
	testutil.ReadJSON(t, resp, &games1)
	assert.Len(t, games1, 0)

	// Player 2 should still see it
	resp = testutil.DoRequest(t, env, "GET", "/api/games", nil, testutil.AuthHeader(token2))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var games2 []map[string]interface{}
	testutil.ReadJSON(t, resp, &games2)
	assert.Len(t, games2, 1)
}

func TestHideGame_NotInGame(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)

	token2 := testutil.GenerateToken(t, user2ID, "player2")

	resp := testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token2))
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
}

func TestHideGame_AlreadyHidden(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	gameID, _ := testutil.CreateTestGame(t, env.Pool, userID)
	token := testutil.GenerateToken(t, userID, "player1")

	// Hide once
	resp := testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Hide again - should return 404
	resp = testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token))
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
}

func TestHideGame_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	resp := testutil.DoRequest(t, env, "POST", "/api/games/00000000-0000-0000-0000-000000000000/hide", nil, nil)
	testutil.AssertProblemDetails(t, resp, http.StatusUnauthorized)
}

func TestHideGame_HiddenFromHistory(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")

	gameID, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, gameID, user2ID)
	testutil.CompleteTestGame(t, env.Pool, gameID, &user1ID)

	token := testutil.GenerateToken(t, user1ID, "player1")

	// Hide the completed game
	resp := testutil.DoRequest(t, env, "POST", "/api/games/"+gameID+"/hide", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Should not appear in history
	resp = testutil.DoRequest(t, env, "GET", "/api/users/me/history", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 0)
}

// --- Stats Tests ---

func TestGetStats_Empty(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/stats", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stats map[string]interface{}
	testutil.ReadJSON(t, resp, &stats)
	assert.Equal(t, float64(0), stats["wins"])
	assert.Equal(t, float64(0), stats["losses"])
	assert.Equal(t, float64(0), stats["draws"])
	assert.Equal(t, float64(0), stats["abandoned"])
	assert.Equal(t, float64(0), stats["averageVp"])
	assert.Empty(t, stats["factionStats"])
}

func TestGetStats_WithGames(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	testutil.SeedFaction(t, env.Pool, "faction-sm", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "faction-ork", "Orks")

	// Game 1: user1 wins with Space Marines (VP: 10+5+3+2 = 20)
	g1, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g1, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g1, user1ID, "faction-sm")
	testutil.SetPlayerFaction(t, env.Pool, g1, user2ID, "faction-ork")
	testutil.SetPlayerVP(t, env.Pool, g1, user1ID, 10, 5, 3, 2)
	testutil.SetPlayerVP(t, env.Pool, g1, user2ID, 5, 3, 0, 1)
	testutil.CompleteTestGame(t, env.Pool, g1, &user1ID)

	// Game 2: user1 loses with Orks (VP: 3+2+0+1 = 6)
	g2, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g2, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g2, user1ID, "faction-ork")
	testutil.SetPlayerFaction(t, env.Pool, g2, user2ID, "faction-sm")
	testutil.SetPlayerVP(t, env.Pool, g2, user1ID, 3, 2, 0, 1)
	testutil.SetPlayerVP(t, env.Pool, g2, user2ID, 10, 5, 3, 2)
	testutil.CompleteTestGame(t, env.Pool, g2, &user2ID)

	// Game 3: draw (VP: 8+4+0+0 = 12)
	g3, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g3, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g3, user1ID, "faction-sm")
	testutil.SetPlayerFaction(t, env.Pool, g3, user2ID, "faction-sm")
	testutil.SetPlayerVP(t, env.Pool, g3, user1ID, 8, 4, 0, 0)
	testutil.SetPlayerVP(t, env.Pool, g3, user2ID, 8, 4, 0, 0)
	testutil.CompleteTestGame(t, env.Pool, g3, nil)

	token := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/stats", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stats map[string]interface{}
	testutil.ReadJSON(t, resp, &stats)
	assert.Equal(t, float64(1), stats["wins"])
	assert.Equal(t, float64(1), stats["losses"])
	assert.Equal(t, float64(1), stats["draws"])
	assert.Equal(t, float64(0), stats["abandoned"])

	// Average VP: (20 + 6 + 12) / 3 ≈ 12.67
	avgVP := stats["averageVp"].(float64)
	assert.InDelta(t, 12.67, avgVP, 0.01)

	// Faction stats: Space Marines (2 games, 1 win), Orks (1 game, 0 wins)
	factionStats := stats["factionStats"].([]interface{})
	assert.Len(t, factionStats, 2)
	// Sorted by games_played DESC, so Space Marines first
	sm := factionStats[0].(map[string]interface{})
	assert.Equal(t, "Space Marines", sm["factionName"])
	assert.Equal(t, float64(2), sm["gamesPlayed"])
	assert.Equal(t, float64(1), sm["wins"])

	ork := factionStats[1].(map[string]interface{})
	assert.Equal(t, "Orks", ork["factionName"])
	assert.Equal(t, float64(1), ork["gamesPlayed"])
	assert.Equal(t, float64(0), ork["wins"])
}

func TestGetStats_Abandoned(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")

	g1, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g1, user2ID)
	testutil.SetPlayerVP(t, env.Pool, g1, user1ID, 5, 0, 0, 0)
	testutil.AbandonTestGame(t, env.Pool, g1)

	token := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/stats", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stats map[string]interface{}
	testutil.ReadJSON(t, resp, &stats)
	assert.Equal(t, float64(0), stats["wins"])
	assert.Equal(t, float64(0), stats["losses"])
	assert.Equal(t, float64(0), stats["draws"])
	assert.Equal(t, float64(1), stats["abandoned"])
	assert.Equal(t, float64(5), stats["averageVp"])
}

// --- History Filter Tests ---

func TestGetHistory_NoFilter(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")

	g1, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g1, user2ID)
	testutil.CompleteTestGame(t, env.Pool, g1, &user1ID)

	g2, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g2, user2ID)
	testutil.AbandonTestGame(t, env.Pool, g2)

	token := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/history", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 2)
}

func TestGetHistory_FilterByMyFaction(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	testutil.SeedFaction(t, env.Pool, "faction-sm", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "faction-ork", "Orks")

	// Game 1: user1 as Space Marines
	g1, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g1, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g1, user1ID, "faction-sm")
	testutil.SetPlayerFaction(t, env.Pool, g1, user2ID, "faction-ork")
	testutil.CompleteTestGame(t, env.Pool, g1, &user1ID)

	// Game 2: user1 as Orks
	g2, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g2, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g2, user1ID, "faction-ork")
	testutil.SetPlayerFaction(t, env.Pool, g2, user2ID, "faction-sm")
	testutil.CompleteTestGame(t, env.Pool, g2, &user2ID)

	token := testutil.GenerateToken(t, user1ID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/history?myFaction=Space+Marines", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 1)
	assert.Equal(t, g1, games[0]["id"])
}

func TestGetHistory_FilterByOpponentFaction(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	user1ID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	user2ID := testutil.CreateTestUser(t, env.Pool, "discord-2", "player2")
	testutil.SeedFaction(t, env.Pool, "faction-sm", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "faction-ork", "Orks")

	// Game 1: opponent as Orks
	g1, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g1, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g1, user1ID, "faction-sm")
	testutil.SetPlayerFaction(t, env.Pool, g1, user2ID, "faction-ork")
	testutil.CompleteTestGame(t, env.Pool, g1, &user1ID)

	// Game 2: opponent as Space Marines
	g2, _ := testutil.CreateTestGame(t, env.Pool, user1ID)
	testutil.JoinTestGame(t, env.Pool, g2, user2ID)
	testutil.SetPlayerFaction(t, env.Pool, g2, user1ID, "faction-ork")
	testutil.SetPlayerFaction(t, env.Pool, g2, user2ID, "faction-sm")
	testutil.CompleteTestGame(t, env.Pool, g2, &user2ID)

	token := testutil.GenerateToken(t, user1ID, "player1")

	// Filter for games where opponent played Orks
	resp := testutil.DoRequest(t, env, "GET", "/api/users/me/history?opponentFaction=Orks", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var games []map[string]interface{}
	testutil.ReadJSON(t, resp, &games)
	assert.Len(t, games, 1)
	assert.Equal(t, g1, games[0]["id"])
}
