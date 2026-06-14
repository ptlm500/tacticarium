package handler_test

import (
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListForceDispositions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedForceDisposition(t, env.Pool, "attrition", "Attrition")
	testutil.SeedForceDisposition(t, env.Pool, "recon", "Recon")

	resp := testutil.DoRequest(t, env, "GET", "/api/force-dispositions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var dispositions []map[string]interface{}
	testutil.ReadJSON(t, resp, &dispositions)
	assert.Len(t, dispositions, 2)
	// Ordered by name.
	assert.Equal(t, "Attrition", dispositions[0]["name"])
	assert.Equal(t, "Recon", dispositions[1]["name"])
}

func TestListMissions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedPrimaryMission(t, env.Pool, "m1", "Take and Hold")
	testutil.SeedPrimaryMission(t, env.Pool, "m2", "The Ritual")
	testutil.SeedPrimaryMission(t, env.Pool, "m3", "Supply Drop")

	resp := testutil.DoRequest(t, env, "GET", "/api/missions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var missions []map[string]interface{}
	testutil.ReadJSON(t, resp, &missions)
	assert.Len(t, missions, 3)
	// 11e caps are exposed per mission.
	for _, m := range missions {
		if m["name"] == "Take and Hold" {
			assert.Equal(t, float64(45), m["vpPerGameCap"])
			assert.Equal(t, float64(15), m["vpPerRoundCap"])
		}
	}
}

func TestListMissionMatchups(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	// Matchups reference dispositions and a primary mission.
	testutil.SeedForceDisposition(t, env.Pool, "attrition", "Attrition")
	testutil.SeedForceDisposition(t, env.Pool, "recon", "Recon")
	testutil.SeedPrimaryMission(t, env.Pool, "m1", "Take and Hold")
	testutil.SeedMissionMatchup(t, env.Pool, "mm1", "attrition", "recon", "m1")

	resp := testutil.DoRequest(t, env, "GET", "/api/mission-matchups", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var matchups []map[string]interface{}
	testutil.ReadJSON(t, resp, &matchups)
	require.Len(t, matchups, 1)
	assert.Equal(t, "attrition", matchups[0]["disposition"])
	assert.Equal(t, "recon", matchups[0]["opponentDisposition"])
	assert.Equal(t, "m1", matchups[0]["missionId"])
}

func TestListSecondaryCards(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedSecondaryCard(t, env.Pool, "sec-1", "Assassination")
	testutil.SeedSecondaryCard(t, env.Pool, "sec-2", "Behind Enemy Lines")

	resp := testutil.DoRequest(t, env, "GET", "/api/secondary-cards", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var cards []map[string]interface{}
	testutil.ReadJSON(t, resp, &cards)
	assert.Len(t, cards, 2)
	for _, c := range cards {
		assert.Equal(t, "secondary", c["cardType"])
	}
}

func TestListDeploymentPatterns(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedDeploymentPattern(t, env.Pool, "dp1", "Crucible of Battle")
	testutil.SeedDeploymentPattern(t, env.Pool, "dp2", "Hammer and Anvil")

	resp := testutil.DoRequest(t, env, "GET", "/api/deployment-patterns", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var patterns []map[string]interface{}
	testutil.ReadJSON(t, resp, &patterns)
	assert.Len(t, patterns, 2)
	assert.Equal(t, "Crucible of Battle", patterns[0]["name"])
	assert.Equal(t, "Hammer and Anvil", patterns[1]["name"])
}

func TestMissionEndpoints_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	endpoints := []string{
		"/api/force-dispositions",
		"/api/missions",
		"/api/mission-matchups",
		"/api/secondary-cards",
		"/api/deployment-patterns",
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep, nil, nil)
		testutil.AssertProblemDetails(t, resp, http.StatusUnauthorized)
	}
}

func TestMissionEndpoints_Empty(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	endpoints := []string{
		"/api/force-dispositions",
		"/api/missions",
		"/api/mission-matchups",
		"/api/secondary-cards",
		"/api/deployment-patterns",
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep, nil, testutil.AuthHeader(token))
		require.Equal(t, http.StatusOK, resp.StatusCode, "endpoint %s", ep)

		var items []map[string]interface{}
		testutil.ReadJSON(t, resp, &items)
		assert.Len(t, items, 0, "endpoint %s should return empty array", ep)
	}
}
