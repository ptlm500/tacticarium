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
		{"GET", "/api/admin/stratagems"},
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
		"id": "SM", "name": "Space Marines",
		"parentFactionId": "IMP", "factionRuleId": "oath",
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
	assert.Equal(t, "IMP", got["parentFactionId"])
	assert.Equal(t, "oath", got["factionRuleId"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/factions/SM", map[string]interface{}{
		"name": "Adeptus Astartes", "parentFactionId": "", "factionRuleId": "",
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
		"detachmentPoints": 0, "forceDispositions": []string{"attrition", "recon"},
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
	assert.ElementsMatch(t, []interface{}{"attrition", "recon"}, detachments[0]["forceDispositions"])

	// Update
	resp = testutil.DoRequest(t, env, "PUT", "/api/admin/detachments/gladius", map[string]interface{}{
		"factionId": "SM", "name": "Gladius TF Updated", "forceDispositions": []string{"attrition"},
	}, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete
	resp = testutil.DoRequest(t, env, "DELETE", "/api/admin/detachments/gladius", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()
}

// --- Stratagems (read-only list) ---
//
// Stratagem write CRUD was removed in 11e; stratagems are bulk-seeded from
// 40kdc-data into stratagems_11e and only listed via the admin API.
func TestAdminStratagems_List(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedDetachment(t, env.Pool, "gladius", "SM", "Gladius Task Force")
	testutil.SeedStratagem(t, env.Pool, "armour-of-contempt", "SM", "", "Armour of Contempt")
	testutil.SeedStratagem(t, env.Pool, "only-in-death", "SM", "gladius", "Only in Death Does Duty End")

	// List all.
	resp := testutil.DoRequest(t, env, "GET", "/api/admin/stratagems", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var all []map[string]interface{}
	testutil.ReadJSON(t, resp, &all)
	assert.Len(t, all, 2)

	// Filter by faction.
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/stratagems?faction_id=SM", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var byFaction []map[string]interface{}
	testutil.ReadJSON(t, resp, &byFaction)
	assert.Len(t, byFaction, 2)

	// Filter by detachment.
	resp = testutil.DoRequest(t, env, "GET", "/api/admin/stratagems?detachment_id=gladius", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var byDetachment []map[string]interface{}
	testutil.ReadJSON(t, resp, &byDetachment)
	require.Len(t, byDetachment, 1)
	assert.Equal(t, "Only in Death Does Duty End", byDetachment[0]["name"])
	assert.Equal(t, float64(1), byDetachment[0]["cpCost"])
}

// --- Update not found (carried-forward CRUD entities only) ---

func TestAdminUpdate_NotFound(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)
	token := adminToken(t)

	cases := []struct {
		path string
		body map[string]interface{}
	}{
		{"/api/admin/factions/NOPE", map[string]interface{}{"name": "x"}},
		{"/api/admin/detachments/NOPE", map[string]interface{}{"factionId": "x", "name": "x", "forceDispositions": []string{}}},
	}

	for _, tc := range cases {
		resp := testutil.DoRequest(t, env, "PUT", tc.path, tc.body, testutil.AuthHeader(token))
		testutil.AssertProblemDetails(t, resp, http.StatusNotFound)
	}
}
