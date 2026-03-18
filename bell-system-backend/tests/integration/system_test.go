//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GET /api/v1/system/state ---

func TestSystemState_GetState(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodGet, "/api/v1/system/state", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var state models.SystemStateResponse
	readJSON(t, resp, &state)
	assert.Equal(t, "active", state.State)
	assert.False(t, state.LastUpdated.IsZero())
}

// --- POST /api/v1/system/state ---

func TestSystemState_SetStatePaused(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	body, _ := json.Marshal(models.SystemStateRequest{State: "paused"})
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", bytes.NewReader(body), token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var state models.SystemStateResponse
	readJSON(t, resp, &state)
	assert.Equal(t, "paused", state.State)
}

func TestSystemState_SetStateActive(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// First pause
	body, _ := json.Marshal(models.SystemStateRequest{State: "paused"})
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", bytes.NewReader(body), token)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Then activate
	body, _ = json.Marshal(models.SystemStateRequest{State: "active"})
	resp = doRequest(t, http.MethodPost, "/api/v1/system/state", bytes.NewReader(body), token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var state models.SystemStateResponse
	readJSON(t, resp, &state)
	assert.Equal(t, "active", state.State)
}

func TestSystemState_SetStateInvalid(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	body, _ := json.Marshal(map[string]string{"state": "invalid"})
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", bytes.NewReader(body), token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSystemState_SetStatePersists(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Set to paused
	body, _ := json.Marshal(models.SystemStateRequest{State: "paused"})
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", bytes.NewReader(body), token)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read back
	resp = doRequest(t, http.MethodGet, "/api/v1/system/state", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var state models.SystemStateResponse
	readJSON(t, resp, &state)
	assert.Equal(t, "paused", state.State)
}

// --- POST /api/v1/system/cancel-next-bell ---

func TestSystemState_CancelNextBell_NoPending(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// No scheduler running in integration tests, so no pending bells
	resp := doRequest(t, http.MethodPost, "/api/v1/system/cancel-next-bell", nil, token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
