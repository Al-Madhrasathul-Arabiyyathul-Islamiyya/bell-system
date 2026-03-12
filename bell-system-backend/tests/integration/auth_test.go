//go:build integration

package integration_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin_Success(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"username":"admin","password":"admin123"}`), "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	readJSON(t, resp, &result)

	assert.NotEmpty(t, result["token"])

	user, ok := result["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "admin", user["username"])
	assert.Equal(t, "admin", user["role"])
	assert.NotEmpty(t, user["id"])
}

func TestLogin_WrongPassword(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"username":"admin","password":"wrongpassword"}`), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_NonexistentUser(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"username":"nonexistent","password":"password"}`), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_MissingFields(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{}`), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestChangePassword_Success(t *testing.T) {
	cleanAndSeed(t)

	token := loginAs(t, "admin", "admin123")

	// Change password
	resp := doRequest(t, http.MethodPost, "/api/auth/change-password",
		bytes.NewBufferString(`{"oldPassword":"admin123","newPassword":"newpass123"}`), token)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Login with new password should succeed
	resp2 := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"username":"admin","password":"newpass123"}`), "")
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Login with old password should fail
	resp3 := doRequest(t, http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"username":"admin","password":"admin123"}`), "")
	defer resp3.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp3.StatusCode)
}

func TestLogout(t *testing.T) {
	cleanAndSeed(t)

	token := loginAs(t, "admin", "admin123")

	resp := doRequest(t, http.MethodPost, "/api/auth/logout", nil, token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProtectedEndpoint_NoToken(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/change-password",
		bytes.NewBufferString(`{"oldPassword":"x","newPassword":"y"}`), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodPost, "/api/auth/change-password",
		bytes.NewBufferString(`{"oldPassword":"x","newPassword":"y"}`), "garbage-token")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
