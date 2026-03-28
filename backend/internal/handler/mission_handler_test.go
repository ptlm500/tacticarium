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
	testutil.SeedMission(t, env.Pool, "leviathan", "Sweep and Clear")
	testutil.SeedMission(t, env.Pool, "leviathan", "The Ritual")
	testutil.SeedMission(t, env.Pool, "leviathan", "Supply Drop")

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
	testutil.SeedSecondary(t, env.Pool, "leviathan", "Assassination", "fixed")
	testutil.SeedSecondary(t, env.Pool, "leviathan", "Behind Enemy Lines", "tactical")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/leviathan/secondaries", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var secondaries []map[string]interface{}
	testutil.ReadJSON(t, resp, &secondaries)
	assert.Len(t, secondaries, 2)
}

func TestListGambits(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedMissionPack(t, env.Pool, "leviathan", "Leviathan")
	testutil.SeedGambit(t, env.Pool, "leviathan", "Prepared Positions")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-packs/leviathan/gambits", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var gambits []map[string]interface{}
	testutil.ReadJSON(t, resp, &gambits)
	assert.Len(t, gambits, 1)
	assert.Equal(t, "Prepared Positions", gambits[0]["name"])
}

func TestMissionEndpoints_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	endpoints := []string{
		"/api/mission-packs",
		"/api/mission-packs/leviathan/missions",
		"/api/mission-packs/leviathan/secondaries",
		"/api/mission-packs/leviathan/gambits",
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep, nil, nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "endpoint %s should require auth", ep)
		resp.Body.Close()
	}
}
