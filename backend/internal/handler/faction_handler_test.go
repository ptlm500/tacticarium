package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFactions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedFaction(t, env.Pool, "NEC", "Necrons")
	testutil.SeedFaction(t, env.Pool, "ORK", "Orks")

	resp := testutil.DoRequest(t, env, "GET", "/api/factions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var factions []map[string]interface{}
	testutil.ReadJSON(t, resp, &factions)
	assert.Len(t, factions, 3)
	assert.Equal(t, "Necrons", factions[0]["name"])
	assert.Equal(t, "Orks", factions[1]["name"])
	assert.Equal(t, "Space Marines", factions[2]["name"])
}

func TestListFactions_Empty(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	resp := testutil.DoRequest(t, env, "GET", "/api/factions", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestListDetachments(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedDetachment(t, env.Pool, "det-1", "SM", "Gladius Task Force")
	testutil.SeedDetachment(t, env.Pool, "det-2", "SM", "Ironstorm Spearhead")

	resp := testutil.DoRequest(t, env, "GET", "/api/factions/SM/detachments", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var detachments []map[string]interface{}
	testutil.ReadJSON(t, resp, &detachments)
	assert.Len(t, detachments, 2)
}

func TestListDetachments_WrongFaction(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedDetachment(t, env.Pool, "det-1", "SM", "Gladius Task Force")

	resp := testutil.DoRequest(t, env, "GET", "/api/factions/NEC/detachments", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestListStratagems(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedStratagem(t, env.Pool, "strat-1", "SM", "", "Armor of Contempt")
	testutil.SeedStratagem(t, env.Pool, "strat-2", "SM", "", "Shock Assault")

	resp := testutil.DoRequest(t, env, "GET", "/api/factions/SM/stratagems", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stratagems []map[string]interface{}
	testutil.ReadJSON(t, resp, &stratagems)
	assert.Len(t, stratagems, 2)
}

func TestListDetachmentStratagems(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	testutil.SeedDetachment(t, env.Pool, "det-1", "SM", "Gladius Task Force")
	testutil.SeedStratagem(t, env.Pool, "strat-1", "SM", "det-1", "Gladius Strat")
	testutil.SeedStratagem(t, env.Pool, "strat-2", "SM", "", "Generic Strat")

	resp := testutil.DoRequest(t, env, "GET", "/api/detachments/det-1/stratagems", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stratagems []map[string]interface{}
	testutil.ReadJSON(t, resp, &stratagems)
	assert.Len(t, stratagems, 1)
	assert.Equal(t, "Gladius Strat", stratagems[0]["name"])
}

// --- Boarding Actions game-mode filter ---

func TestListDetachments_ExcludesBoardingActions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	// Core detachment (default game_mode = 'core')
	testutil.SeedDetachment(t, env.Pool, "det-core", "SM", "Gladius Task Force")
	// Boarding Actions detachment — insert directly so we can set game_mode.
	_, err := env.Pool.Exec(context.Background(),
		`INSERT INTO detachments (id, faction_id, name, game_mode) VALUES ($1, $2, $3, $4)`,
		"det-ba", "SM", "Boarding Butchers", "boarding_actions")
	require.NoError(t, err)

	resp := testutil.DoRequest(t, env, "GET", "/api/factions/SM/detachments", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var detachments []map[string]interface{}
	testutil.ReadJSON(t, resp, &detachments)
	assert.Len(t, detachments, 1)
	assert.Equal(t, "Gladius Task Force", detachments[0]["name"])
}

func TestListStratagems_ExcludesBoardingActions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	// Core stratagem
	testutil.SeedStratagem(t, env.Pool, "s-core", "SM", "", "Armor of Contempt")
	// Boarding Actions stratagem — inserted directly so we can set game_mode.
	_, err := env.Pool.Exec(context.Background(),
		`INSERT INTO stratagems (id, faction_id, name, type, cp_cost, turn, phase, description, game_mode)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		"s-ba", "SM", "Explosive Clearance",
		"Boarding Actions \u2013 Battle Tactic Stratagem", 1,
		"Your turn", "Shooting phase", "Boarding Actions stratagem.", "boarding_actions")
	require.NoError(t, err)

	resp := testutil.DoRequest(t, env, "GET", "/api/factions/SM/stratagems", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stratagems []map[string]interface{}
	testutil.ReadJSON(t, resp, &stratagems)
	assert.Len(t, stratagems, 1)
	assert.Equal(t, "Armor of Contempt", stratagems[0]["name"])
}

func TestListDetachmentStratagems_ExcludesBoardingActions(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanAllTables(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "player1")
	token := testutil.GenerateToken(t, userID, "player1")

	testutil.SeedFaction(t, env.Pool, "SM", "Space Marines")
	// Boarding Actions detachment with a single boarding-actions stratagem.
	_, err := env.Pool.Exec(context.Background(),
		`INSERT INTO detachments (id, faction_id, name, game_mode) VALUES ($1, $2, $3, $4)`,
		"det-ba", "SM", "Boarding Butchers", "boarding_actions")
	require.NoError(t, err)
	_, err = env.Pool.Exec(context.Background(),
		`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, turn, phase, description, game_mode)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		"s-ba", "SM", "det-ba", "Savage Resilience",
		"Boarding Butchers \u2013 Battle Tactic Stratagem", 1,
		"Your turn", "Fight phase", "Detachment strat.", "boarding_actions")
	require.NoError(t, err)

	resp := testutil.DoRequest(t, env, "GET", "/api/detachments/det-ba/stratagems", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stratagems []map[string]interface{}
	testutil.ReadJSON(t, resp, &stratagems)
	assert.Len(t, stratagems, 0)
}

func TestFactionEndpoints_Unauthorized(t *testing.T) {
	env := testutil.SharedEnv

	endpoints := []string{
		"/api/factions",
		"/api/factions/SM/detachments",
		"/api/factions/SM/stratagems",
		"/api/detachments/det-1/stratagems",
	}

	for _, ep := range endpoints {
		resp := testutil.DoRequest(t, env, "GET", ep, nil, nil)
		testutil.AssertProblemDetails(t, resp, http.StatusUnauthorized)
	}
}
