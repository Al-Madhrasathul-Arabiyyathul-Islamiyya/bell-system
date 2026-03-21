//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func systemStateBody(state string) *bytes.Buffer {
	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"type":       "system-state",
			"attributes": map[string]any{"state": state},
		},
	})
	return bytes.NewBuffer(body)
}

// --- GET /api/v1/system/state ---

func TestSystemState_GetState(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodGet, "/api/v1/system/state", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var doc jsonapi.Document
	readJSON(t, resp, &doc)
	assert.Equal(t, "system-state", doc.Data.Type)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "active", attrs["state"])
}

// --- POST /api/v1/system/state ---

func TestSystemState_SetStatePaused(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", systemStateBody("paused"), token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var doc jsonapi.Document
	readJSON(t, resp, &doc)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "paused", attrs["state"])
}

func TestSystemState_SetStateActive(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// First pause
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", systemStateBody("paused"), token)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Then activate
	resp = doRequest(t, http.MethodPost, "/api/v1/system/state", systemStateBody("active"), token)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var doc jsonapi.Document
	readJSON(t, resp, &doc)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "active", attrs["state"])
}

func TestSystemState_SetStateInvalid(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", systemStateBody("invalid"), token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSystemState_SetStatePersists(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Set to paused
	resp := doRequest(t, http.MethodPost, "/api/v1/system/state", systemStateBody("paused"), token)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read back
	resp = doRequest(t, http.MethodGet, "/api/v1/system/state", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var doc jsonapi.Document
	readJSON(t, resp, &doc)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "paused", attrs["state"])
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
