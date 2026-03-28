package handler_test

import (
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMissionPacks(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "leviathan", "Leviathan")
	testutil.SeedMissionPack(t, env.Pool, "pariah-nexus", "Pariah Nexus")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var packs []map[string]interface{}
	testutil.ReadJSON(t, resp, &packs)
	assert.Len(t, packs, 2)
}

func TestListMissions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "leviathan", "Leviathan")
	testutil.SeedMission(t, env.Pool, "m1", "leviathan", "Sweep and Clear")
	testutil.SeedMission(t, env.Pool, "m2", "leviathan", "The Ritual")
	testutil.SeedMission(t, env.Pool, "m3", "leviathan", "Supply Drop")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/leviathan/missions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var missions []map[string]interface{}
	testutil.ReadJSON(t, resp, &missions)
	assert.Len(t, missions, 3)
}

func TestListSecondaries(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "leviathan", "Leviathan")
	testutil.SeedSecondary(t, env.Pool, "sec-1", "leviathan", "Assassination", true)
	testutil.SeedSecondary(t, env.Pool, "sec-2", "leviathan", "Behind Enemy Lines", false)

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/leviathan/secondaries", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var secondaries []map[string]interface{}
	testutil.ReadJSON(t, resp, &secondaries)
	assert.Len(t, secondaries, 2)

	// Verify isFixed field is present
	for _, s := range secondaries {
		if s["name"] == "Assassination" {
			assert.Equal(t, true, s["isFixed"])
		}
		if s["name"] == "Behind Enemy Lines" {
			assert.Equal(t, false, s["isFixed"])
		}
	}
}

func TestListGambits(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "leviathan", "Leviathan")
	testutil.SeedGambit(t, env.Pool, "g1", "leviathan", "Prepared Positions")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/leviathan/gambits", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var gambits []map[string]interface{}
	testutil.ReadJSON(t, resp, &gambits)
	assert.Len(t, gambits, 1)
	assert.Equal(t, "Prepared Positions", gambits[0]["name"])
}

func TestListMissionRules(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025-26")
	testutil.SeedMissionRule(t, env.Pool, "rule-1", "ca2025", "Adapt or Die")
	testutil.SeedMissionRule(t, env.Pool, "rule-2", "ca2025", "Hidden Supplies")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/ca2025/rules", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var rules []map[string]interface{}
	testutil.ReadJSON(t, resp, &rules)
	assert.Len(t, rules, 2)
}

func TestListChallengerCards(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025-26")
	testutil.SeedChallengerCard(t, env.Pool, "cc-1", "ca2025", "Challenge A")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/ca2025/challenger-cards", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var cards []map[string]interface{}
	testutil.ReadJSON(t, resp, &cards)
	assert.Len(t, cards, 1)
	assert.Equal(t, "Challenge A", cards[0]["name"])
}

func TestMissionEndpoints_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	endpoints := []string{
		"/api/mission-packs",
		"/api/mission-packs/leviathan/missions",
		"/api/mission-packs/leviathan/secondaries",
		"/api/mission-packs/leviathan/gambits",
		"/api/mission-packs/leviathan/rules",
		"/api/mission-packs/leviathan/challenger-cards",
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep, nil, nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "endpoint %s should require auth", ep)
		resp.Body.Close()
	}
}

func TestMissionEndpoints_EmptyPack(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "empty-pack", "Empty Pack")

	endpoints := []struct {
		path string
	}{
		{"/api/mission-packs/empty-pack/missions"},
		{"/api/mission-packs/empty-pack/secondaries"},
		{"/api/mission-packs/empty-pack/rules"},
		{"/api/mission-packs/empty-pack/challenger-cards"},
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep.path, nil, testutil.AuthHeader(token))
		require.Equal(t, http.StatusOK, resp.StatusCode, "endpoint %s", ep.path)

		var items []map[string]interface{}
		testutil.ReadJSON(t, resp, &items)
		assert.Len(t, items, 0, "endpoint %s should return empty array", ep.path)
	}
}
