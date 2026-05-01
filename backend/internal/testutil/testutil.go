package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peter/tacticarium/backend/internal/auth"
	"github.com/peter/tacticarium/backend/internal/config"
	"github.com/peter/tacticarium/backend/internal/db"
	"github.com/peter/tacticarium/backend/internal/server"
	"github.com/peter/tacticarium/backend/internal/ws"
	"github.com/peter/tacticarium/backend/pkg/invite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"nhooyr.io/websocket"
)

const TestJWTSecret = "test-jwt-secret"

// TestEnv holds the shared test infrastructure.
type TestEnv struct {
	Pool      *pgxpool.Pool
	Server    *httptest.Server
	Hub       *ws.Hub
	container testcontainers.Container
}

// SharedEnv is the package-level shared test environment.
// Initialized by MustSetupTestEnv in TestMain, used by all tests.
var SharedEnv *TestEnv

// MustSetupTestEnv starts a Postgres container and creates the test server.
// Call this from TestMain. Panics on failure. Call Teardown() after m.Run().
func MustSetupTestEnv() *TestEnv {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start Postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to get connection string: %v", err)
	}

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Pass connStr instead of ctx and pool
	if err := db.RunMigrations(connStr); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	hub := ws.NewHub()

	cfg := &config.Config{
		DatabaseURL:      connStr,
		JWTSecret:        TestJWTSecret,
		FrontendURL:      "http://localhost:3000",
		AdminFrontendURL: "http://localhost:5174",
		AdminGitHubIDs:   "12345",
		Port:             "0",
	}

	r := server.NewRouter(pool, hub, cfg)
	srv := httptest.NewServer(r)

	env := &TestEnv{
		Pool:      pool,
		Server:    srv,
		Hub:       hub,
		container: pgContainer,
	}

	SharedEnv = env
	return env
}

// Teardown cleans up the shared test environment.
func (env *TestEnv) Teardown() {
	env.Server.Close()
	env.Pool.Close()
	if env.container != nil {
		env.container.Terminate(context.Background())
	}
}

// CleanDatabase truncates all user-created data tables, preserving reference data.
func CleanDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		TRUNCATE stratagem_usage, game_events, game_players, games, users CASCADE
	`)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}
}

// CleanAllTables truncates all tables including reference data.
func CleanAllTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		TRUNCATE stratagem_usage, game_events, game_players, games,
		         stratagems, detachments, factions, gambits, secondaries, missions,
		         mission_rules, challenger_cards, mission_packs, users CASCADE
	`)
	if err != nil {
		t.Fatalf("Failed to clean all tables: %v", err)
	}
}

// GenerateToken creates a valid JWT for testing.
func GenerateToken(t *testing.T, userID, username string) string {
	t.Helper()
	token, err := auth.GenerateToken(TestJWTSecret, userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	return token
}

// GenerateAdminToken creates a valid admin JWT for testing.
func GenerateAdminToken(t *testing.T, githubID, username string) string {
	t.Helper()
	token, err := auth.GenerateTokenWithRole(TestJWTSecret, githubID, username, "admin")
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}
	return token
}

// CreateTestUser inserts a user and returns its UUID.
func CreateTestUser(t *testing.T, pool *pgxpool.Pool, discordID, username string) string {
	t.Helper()
	var userID string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (discord_id, discord_username) VALUES ($1, $2) RETURNING id`,
		discordID, username,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// CreateTestGame creates a game and adds the creator as player 1. Returns (gameID, inviteCode).
func CreateTestGame(t *testing.T, pool *pgxpool.Pool, creatorUserID string) (string, string) {
	t.Helper()
	code := invite.GenerateCode(6)
	var gameID string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO games (invite_code) VALUES ($1) RETURNING id`, code,
	).Scan(&gameID)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}
	_, err = pool.Exec(context.Background(),
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, 1)`,
		gameID, creatorUserID)
	if err != nil {
		t.Fatalf("Failed to add creator as player 1: %v", err)
	}
	return gameID, code
}

// JoinTestGame adds a user as player 2 to an existing game.
func JoinTestGame(t *testing.T, pool *pgxpool.Pool, gameID, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO game_players (game_id, user_id, player_number) VALUES ($1, $2, 2)`,
		gameID, userID)
	if err != nil {
		t.Fatalf("Failed to join test game: %v", err)
	}
}

// SeedFaction inserts a faction for testing.
func SeedFaction(t *testing.T, pool *pgxpool.Pool, id, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO factions (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, name)
	if err != nil {
		t.Fatalf("Failed to seed faction: %v", err)
	}
}

// SeedDetachment inserts a detachment for testing.
func SeedDetachment(t *testing.T, pool *pgxpool.Pool, id, factionID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO detachments (id, faction_id, name) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		id, factionID, name)
	if err != nil {
		t.Fatalf("Failed to seed detachment: %v", err)
	}
}

// SeedStratagem inserts a stratagem for testing.
func SeedStratagem(t *testing.T, pool *pgxpool.Pool, id, factionID, detachmentID, name string) {
	t.Helper()
	var detID *string
	if detachmentID != "" {
		detID = &detachmentID
	}
	_, err := pool.Exec(context.Background(),
		`INSERT INTO stratagems (id, faction_id, detachment_id, name, type, cp_cost, turn, phase, description)
		 VALUES ($1, $2, $3, $4, 'Battle Tactic', 1, 'Your turn', 'Shooting phase', 'Test stratagem')
		 ON CONFLICT DO NOTHING`,
		id, factionID, detID, name)
	if err != nil {
		t.Fatalf("Failed to seed stratagem: %v", err)
	}
}

// SeedMissionPack inserts a mission pack.
func SeedMissionPack(t *testing.T, pool *pgxpool.Pool, id, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO mission_packs (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, name)
	if err != nil {
		t.Fatalf("Failed to seed mission pack: %v", err)
	}
}

// SeedMission inserts a mission with a TEXT PK.
func SeedMission(t *testing.T, pool *pgxpool.Pool, id, packID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO missions (id, mission_pack_id, name, description, scoring_rules) VALUES ($1, $2, $3, 'Test mission', '[]') ON CONFLICT DO NOTHING`,
		id, packID, name)
	if err != nil {
		t.Fatalf("Failed to seed mission: %v", err)
	}
}

// SeedSecondary inserts a secondary objective with a TEXT PK.
func SeedSecondary(t *testing.T, pool *pgxpool.Pool, id, packID, name string, isFixed bool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO secondaries (id, mission_pack_id, name, description, max_vp, is_fixed)
		 VALUES ($1, $2, $3, 'Test secondary', 5, $4) ON CONFLICT DO NOTHING`,
		id, packID, name, isFixed)
	if err != nil {
		t.Fatalf("Failed to seed secondary: %v", err)
	}
}

// SeedMissionRule inserts a mission rule (twist) with a TEXT PK.
func SeedMissionRule(t *testing.T, pool *pgxpool.Pool, id, packID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO mission_rules (id, mission_pack_id, name, description) VALUES ($1, $2, $3, 'Test rule') ON CONFLICT DO NOTHING`,
		id, packID, name)
	if err != nil {
		t.Fatalf("Failed to seed mission rule: %v", err)
	}
}

// SeedChallengerCard inserts a challenger card with a TEXT PK.
func SeedChallengerCard(t *testing.T, pool *pgxpool.Pool, id, packID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO challenger_cards (id, mission_pack_id, name, description) VALUES ($1, $2, $3, 'Test card') ON CONFLICT DO NOTHING`,
		id, packID, name)
	if err != nil {
		t.Fatalf("Failed to seed challenger card: %v", err)
	}
}

// SeedGambit inserts a gambit.
func SeedGambit(t *testing.T, pool *pgxpool.Pool, id, packID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO gambits (id, mission_pack_id, name, description, vp_value)
		 VALUES ($1, $2, $3, 'Test gambit', 12) ON CONFLICT DO NOTHING`,
		id, packID, name)
	if err != nil {
		t.Fatalf("Failed to seed gambit: %v", err)
	}
}

// CompleteTestGame marks a game as completed with an optional winner.
func CompleteTestGame(t *testing.T, pool *pgxpool.Pool, gameID string, winnerID *string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE games SET status = 'completed', completed_at = NOW(), winner_id = $2 WHERE id = $1`,
		gameID, winnerID)
	if err != nil {
		t.Fatalf("Failed to complete test game: %v", err)
	}
}

// AbandonTestGame marks a game as abandoned.
func AbandonTestGame(t *testing.T, pool *pgxpool.Pool, gameID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE games SET status = 'abandoned', completed_at = NOW() WHERE id = $1`, gameID)
	if err != nil {
		t.Fatalf("Failed to abandon test game: %v", err)
	}
}

// SetPlayerFaction sets the faction for a player in a game.
func SetPlayerFaction(t *testing.T, pool *pgxpool.Pool, gameID, userID, factionID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE game_players SET faction_id = $3 WHERE game_id = $1 AND user_id = $2`,
		gameID, userID, factionID)
	if err != nil {
		t.Fatalf("Failed to set player faction: %v", err)
	}
}

// SetPlayerVP sets VP columns for a player in a game.
func SetPlayerVP(t *testing.T, pool *pgxpool.Pool, gameID, userID string, primary, secondary, gambit, paint int) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE game_players SET vp_primary = $3, vp_secondary = $4, vp_gambit = $5, vp_paint = $6
		 WHERE game_id = $1 AND user_id = $2`,
		gameID, userID, primary, secondary, gambit, paint)
	if err != nil {
		t.Fatalf("Failed to set player VP: %v", err)
	}
}

// AuthHeader returns headers with a Bearer token.
func AuthHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// DoRequest performs an HTTP request against the test server.
func DoRequest(t *testing.T, env *TestEnv, method, path string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, env.Server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// ReadJSON reads and unmarshals the response body.
func ReadJSON(t *testing.T, resp *http.Response, out interface{}) {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v\nBody: %s", err, string(data))
	}
}

// ProblemDetails represents an RFC 9457 problem details response from huma.
type ProblemDetails struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// AssertProblemDetails reads the response body and asserts it is a valid
// RFC 9457 problem details JSON object with the expected status code.
// Returns the parsed ProblemDetails for further assertions.
func AssertProblemDetails(t *testing.T, resp *http.Response, expectedStatus int) ProblemDetails {
	t.Helper()
	defer resp.Body.Close()

	assert.Equal(t, expectedStatus, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	var pd ProblemDetails
	require.NoError(t, json.Unmarshal(data, &pd), "response is not valid JSON: %s", string(data))

	assert.Equal(t, expectedStatus, pd.Status, "problem details status should match HTTP status")
	assert.NotEmpty(t, pd.Title, "problem details should have a title")

	return pd
}

// DialWS connects to the WebSocket endpoint for a game.
func DialWS(t *testing.T, env *TestEnv, gameID, token string) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	url := fmt.Sprintf("%s/ws/game/%s?token=%s", wsURL, gameID, token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}

	t.Cleanup(func() {
		conn.Close(websocket.StatusNormalClosure, "test done")
	})

	return conn
}

// DialSpectatorWS connects to the public spectator WebSocket endpoint for a game.
func DialSpectatorWS(t *testing.T, env *TestEnv, gameID string) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(env.Server.URL, "http://", "ws://", 1)
	url := fmt.Sprintf("%s/ws/game/%s/spectate", wsURL, gameID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("Failed to dial spectator WebSocket: %v", err)
	}

	t.Cleanup(func() {
		conn.Close(websocket.StatusNormalClosure, "test done")
	})

	return conn
}

// ReadWSMessage reads a single WebSocket message with a timeout.
func ReadWSMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read WebSocket message: %v", err)
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("Failed to unmarshal WS message: %v", err)
	}
	return msg
}

// SendWSMessage sends a JSON message over WebSocket.
func SendWSMessage(t *testing.T, conn *websocket.Conn, msg interface{}) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal WS message: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatalf("Failed to write WS message: %v", err)
	}
}

// DrainUntil reads and discards messages until the specified type is found or timeout.
func DrainUntil(t *testing.T, conn *websocket.Conn, msgType string, timeout time.Duration) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		msg := ReadWSMessage(t, conn, remaining)
		if msg["type"] == msgType {
			return msg
		}
	}
	t.Fatalf("Timed out waiting for message type %q", msgType)
	return nil
}
