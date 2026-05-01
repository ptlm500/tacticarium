package handler_test

import (
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func adminToken(t *testing.T) string {
	t.Helper()
	return testutil.GenerateAdminToken(t, "12345", "admin-user")
}

// --- Auth ---

func TestAdminMe(t *testing.T) {
	env := testutil.SharedEnv
	token := adminToken(t)

	resp := testutil.DoRequest(t, env, "GET", "/api/admin/me", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	testutil.ReadJSON(t, resp, &body)
	assert.Equal(t, "12345", body["githubId"])
	assert.Equal(t, "admin-user", body["githubUser"])
}

func TestAdminEndpoints_NoToken(t *testing.T) {
	env := testutil.SharedEnv

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/me"},
		{"GET", "/api/admin/factions"},
		{"POST", "/api/admin/factions"},
		{"GET", "/api/admin/missions"},
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, ep.method, ep.path, nil, nil)
		testutil.AssertProblemDetails(t, resp, http.StatusUnauthorized)
	}
}

func TestAdminEndpoints_PlayerTokenForbidden(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	playerToken := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/admin/factions", nil, testutil.AuthHeader(playerToken))
	testutil.AssertProblemDetails(t, resp, http.StatusForbidden)
}

// --- Factions CRUD ---

func TestAdminFactions_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/factions", map[string]interface{}{
		"id": "SM", "name": "Space Marines", "wahapediaLink": "https://wahapedia.ru/sm",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]interface{}
	testutil.ReadJSON(t, resp, &created)
	assert.Equal(t, "SM", created["id"])
	assert.Equal(t, "Space Marines", created["name"])

	// List
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/factions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var factions []map[string]interface{}
	testutil.ReadJSON(t, resp, &factions)
	require.Len(t, factions, 1)
	assert.Equal(t, "Space Marines", factions[0]["name"])

	// Get
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/factions/SM", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]interface{}
	testutil.ReadJSON(t, resp, &got)
	assert.Equal(t, "Space Marines", got["name"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/factions/SM", map[string]interface{}{
		"name": "Adeptus Astartes", "wahapediaLink": "",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]interface{}
	testutil.ReadJSON(t, resp, &updated)
	assert.Equal(t, "Adeptus Astartes", updated["name"])

	// Get not found
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/factions/NOPE", nil, testutil.AuthHeader(token))
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/factions/SM", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete not found
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/factions/SM", nil, testutil.AuthHeader(token))
	testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
}

func TestAdminFactions_CreateValidation(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	resp := testutil.DoRequest(t, env, "POST", "/api/admin/factions", map[string]interface{}{
		"id": "", "name": "",
	}, testutil.AuthHeader(token))
	testutil.AssertProblemDetails(t, resp, http.StatusBadRequest)
}

// --- Detachments CRUD ---

func TestAdminDetachments_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/detachments", map[string]interface{}{
		"id": "gladius", "factionId": "SM", "name": "Gladius Task Force",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List with filter
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/detachments?faction_id=SM", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var detachments []map[string]interface{}
	testutil.ReadJSON(t, resp, &detachments)
	require.Len(t, detachments, 1)
	assert.Equal(t, "Gladius Task Force", detachments[0]["name"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/detachments/gladius", map[string]interface{}{
		"factionId": "SM", "name": "Gladius TF Updated",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/detachments/gladius", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Stratagems CRUD ---

func TestAdminStratagems_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/stratagems", map[string]interface{}{
		"id": "strat-1", "factionId": "SM", "name": "Armor of Contempt",
		"type": "Battle Tactic", "cpCost": 1, "turn": "your", "phase": "shooting",
		"description": "Reduce AP by 1",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List with filter
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/stratagems?faction_id=SM", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var stratagems []map[string]interface{}
	testutil.ReadJSON(t, resp, &stratagems)
	require.Len(t, stratagems, 1)
	assert.Equal(t, "Armor of Contempt", stratagems[0]["name"])
	assert.Equal(t, float64(1), stratagems[0]["cpCost"])

	// Get
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/stratagems/strat-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]interface{}
	testutil.ReadJSON(t, resp, &got)
	assert.Equal(t, "Reduce AP by 1", got["description"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/stratagems/strat-1", map[string]interface{}{
		"factionId": "SM", "name": "Armor of Contempt", "type": "Battle Tactic",
		"cpCost": 2, "turn": "your", "phase": "shooting", "description": "Updated desc",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/stratagems/strat-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Mission Packs CRUD ---

func TestAdminMissionPacks_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/mission-packs", map[string]interface{}{
		"id": "ca2025", "name": "Chapter Approved 2025", "description": "2025 season",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/mission-packs", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var packs []map[string]interface{}
	testutil.ReadJSON(t, resp, &packs)
	require.Len(t, packs, 1)
	assert.Equal(t, "Chapter Approved 2025", packs[0]["name"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/mission-packs/ca2025", map[string]interface{}{
		"name": "CA 2025-26", "description": "Updated",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/mission-packs/ca2025", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Missions CRUD ---

func TestAdminMissions_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	// Create with scoring rules
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/missions", map[string]interface{}{
		"id": "mission-1", "missionPackId": "ca2025", "name": "Take and Hold",
		"lore": "Control the field", "description": "Hold objectives",
		"scoringTiming": "end_of_command_phase",
		"scoringRules": []map[string]interface{}{
			{"label": "Hold 1", "vp": 2, "minRound": 0, "description": "Hold one objective"},
			{"label": "Hold 2+", "vp": 3, "minRound": 2, "description": "Hold two or more"},
		},
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// Get and check scoring rules
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/missions/mission-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var mission map[string]interface{}
	testutil.ReadJSON(t, resp, &mission)
	assert.Equal(t, "Take and Hold", mission["name"])
	assert.Equal(t, "end_of_command_phase", mission["scoringTiming"])
	rules := mission["scoringRules"].([]interface{})
	require.Len(t, rules, 2)
	assert.Equal(t, "Hold 1", rules[0].(map[string]interface{})["label"])

	// List with filter
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/missions?pack_id=ca2025", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var missions []map[string]interface{}
	testutil.ReadJSON(t, resp, &missions)
	require.Len(t, missions, 1)

	// Update scoring rules
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/missions/mission-1", map[string]interface{}{
		"missionPackId": "ca2025", "name": "Take and Hold v2",
		"description": "Updated", "scoringTiming": "end_of_turn",
		"scoringRules": []map[string]interface{}{
			{"label": "Hold any", "vp": 5},
		},
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]interface{}
	testutil.ReadJSON(t, resp, &updated)
	assert.Equal(t, "Take and Hold v2", updated["name"])

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/missions/mission-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Secondaries CRUD ---

func TestAdminSecondaries_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	// Create with scoring options
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/secondaries", map[string]interface{}{
		"id": "sec-1", "missionPackId": "ca2025", "name": "Assassination",
		"description": "Kill characters", "maxVp": 15, "isFixed": true,
		"scoringOptions": []map[string]interface{}{
			{"label": "Kill character", "vp": 5, "mode": "fixed"},
			{"label": "Kill warlord", "vp": 5, "mode": ""},
		},
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// Get and verify
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/secondaries/sec-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var sec map[string]interface{}
	testutil.ReadJSON(t, resp, &sec)
	assert.Equal(t, "Assassination", sec["name"])
	assert.Equal(t, float64(15), sec["maxVp"])
	assert.Equal(t, true, sec["isFixed"])
	opts := sec["scoringOptions"].([]interface{})
	require.Len(t, opts, 2)

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/secondaries/sec-1", map[string]interface{}{
		"missionPackId": "ca2025", "name": "Assassination v2",
		"description": "Updated", "maxVp": 12, "isFixed": false,
		"scoringOptions": []map[string]interface{}{},
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/secondaries/sec-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Gambits CRUD ---

func TestAdminGambits_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/gambits", map[string]interface{}{
		"id": "gambit-1", "missionPackId": "ca2025", "name": "Proceed as Planned",
		"description": "Score extra VP", "vpValue": 12,
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/gambits?pack_id=ca2025", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var gambits []map[string]interface{}
	testutil.ReadJSON(t, resp, &gambits)
	require.Len(t, gambits, 1)
	assert.Equal(t, float64(12), gambits[0]["vpValue"])

	// Get
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/gambits/gambit-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/gambits/gambit-1", map[string]interface{}{
		"missionPackId": "ca2025", "name": "Updated Gambit",
		"description": "Changed", "vpValue": 8,
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/gambits/gambit-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Challenger Cards CRUD ---

func TestAdminChallengerCards_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/challenger-cards", map[string]interface{}{
		"id": "card-1", "missionPackId": "ca2025", "name": "Test Card",
		"lore": "Some lore", "description": "Card rules",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/challenger-cards?pack_id=ca2025", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var cards []map[string]interface{}
	testutil.ReadJSON(t, resp, &cards)
	require.Len(t, cards, 1)

	// Get
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/challenger-cards/card-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var card map[string]interface{}
	testutil.ReadJSON(t, resp, &card)
	assert.Equal(t, "Test Card", card["name"])
	assert.Equal(t, "Some lore", card["lore"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/challenger-cards/card-1", map[string]interface{}{
		"missionPackId": "ca2025", "name": "Updated Card",
		"lore": "", "description": "New rules",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/challenger-cards/card-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Mission Rules CRUD ---

func TestAdminMissionRules_CRUD(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	// Create
	resp := testutil.DoRequest(t, env, "POST", "/api/admin/mission-rules", map[string]interface{}{
		"id": "rule-1", "missionPackId": "ca2025", "name": "Chilling Rain",
		"lore": "Dark clouds", "description": "-1 to hit beyond 12\"",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// List
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/mission-rules?pack_id=ca2025", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var rules []map[string]interface{}
	testutil.ReadJSON(t, resp, &rules)
	require.Len(t, rules, 1)

	// Get
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/mission-rules/rule-1", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var rule map[string]interface{}
	testutil.ReadJSON(t, resp, &rule)
	assert.Equal(t, "Chilling Rain", rule["name"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/mission-rules/rule-1", map[string]interface{}{
		"missionPackId": "ca2025", "name": "Chilling Rain v2",
		"lore": "Updated lore", "description": "Updated rules",
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/mission-rules/rule-1", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Update/Delete not found ---

func TestAdminUpdate_NotFound(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedMissionPack(t, env.Pool, "ca2025", "Chapter Approved 2025")

	cases := []struct {
		path string
		body map[string]interface{}
	}{
		{"/api/admin/factions/NOPE", map[string]interface{}{"name": "x"}},
		{"/api/admin/detachments/NOPE", map[string]interface{}{"factionId": "x", "name": "x"}},
		{"/api/admin/stratagems/NOPE", map[string]interface{}{"factionId": "x", "name": "x", "type": "x", "cpCost": 0, "turn": "x", "phase": "x", "description": "x"}},
		{"/api/admin/mission-packs/NOPE", map[string]interface{}{"name": "x"}},
		{"/api/admin/missions/NOPE", map[string]interface{}{"missionPackId": "ca2025", "name": "x", "description": "x", "scoringRules": []interface{}{}, "scoringTiming": "x"}},
		{"/api/admin/secondaries/NOPE", map[string]interface{}{"missionPackId": "ca2025", "name": "x", "description": "x", "maxVp": 0, "isFixed": false, "scoringOptions": []interface{}{}}},
		{"/api/admin/gambits/NOPE", map[string]interface{}{"missionPackId": "ca2025", "name": "x", "description": "x", "vpValue": 0}},
		{"/api/admin/challenger-cards/NOPE", map[string]interface{}{"missionPackId": "ca2025", "name": "x", "description": "x"}},
		{"/api/admin/mission-rules/NOPE", map[string]interface{}{"missionPackId": "ca2025", "name": "x", "description": "x"}},
	}

	for _, tc := range cases {
		resp := testutil.DoRequest(t, env, "PUT", tc.path, tc.body, testutil.AuthHeader(token))
		testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
	}
}
