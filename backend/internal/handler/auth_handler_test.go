package handler_test

import (
	"net/http"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleMe_Success(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "testuser")
	token := testutil.GenerateToken(t, userID, "testuser")

	resp := testutil.DoRequest(t, env, "GET", "/api/auth/me", nil, testutil.AuthHeader(token))
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	testutil.ReadJSON(t, resp, &body)

	assert.Equal(t, userID, body["id"])
	assert.Equal(t, "testuser", body["username"])
	assert.NotEmpty(t, body["createdAt"])
}

func TestHandleMe_NoToken(t *testing.T) {
	env := testutil.SharedEnv

	resp := testutil.DoRequest(t, env, "GET", "/api/auth/me", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestHandleMe_InvalidToken(t *testing.T) {
	env := testutil.SharedEnv

	resp := testutil.DoRequest(t, env, "GET", "/api/auth/me", nil,
		testutil.AuthHeader("not-a-valid-token"))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestHandleMe_UserNotInDB(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	token := testutil.GenerateToken(t, "00000000-0000-0000-0000-000000000000", "ghost")

	resp := testutil.DoRequest(t, env, "GET", "/api/auth/me", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestHandleLogout(t *testing.T) {
	env := testutil.SharedEnv
	testutil.CleanDatabase(t, env.Pool)

	userID := testutil.CreateTestUser(t, env.Pool, "discord-1", "testuser")
	token := testutil.GenerateToken(t, userID, "testuser")

	resp := testutil.DoRequest(t, env, "POST", "/api/auth/logout", nil, testutil.AuthHeader(token))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
