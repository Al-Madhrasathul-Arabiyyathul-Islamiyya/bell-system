//go:build integration

package integration_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCRUD(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Create
	createBody := `{"name":"Evening Session","startTime":"18:30","endTime":"21:00"}`
	resp := doRequest(t, http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(createBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	readJSON(t, resp, &created)

	sessionID, ok := created["id"].(string)
	require.True(t, ok)
	assert.Equal(t, "Evening Session", created["name"])

	// Get by ID
	resp = doRequest(t, http.MethodGet, "/api/v1/sessions/"+sessionID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var fetched map[string]any
	readJSON(t, resp, &fetched)
	assert.Equal(t, "Evening Session", fetched["name"])

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult map[string]any
	readJSON(t, resp, &listResult)
	total := listResult["total"].(float64)
	assert.GreaterOrEqual(t, total, float64(3)) // 2 seeded + 1 created

	// Update
	updateBody := `{"name":"Late Evening Session"}`
	resp = doRequest(t, http.MethodPut, "/api/v1/sessions/"+sessionID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated map[string]any
	readJSON(t, resp, &updated)
	assert.Equal(t, "Late Evening Session", updated["name"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/sessions/"+sessionID, nil, token)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/v1/sessions/"+sessionID, nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestGetCurrentSession(t *testing.T) {
	cleanAndSeed(t)

	resp := doRequest(t, http.MethodGet, "/api/v1/sessions/current", nil, "")
	defer resp.Body.Close()

	// Could be 200 (if current time falls in a session) or 404
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode,
		"expected 200 or 404 for current session")
}

func TestSessionList_Ordered(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult map[string]any
	readJSON(t, resp, &listResult)

	items, ok := listResult["items"].([]any)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(items), 2)

	// Verify Morning comes before Afternoon (ordered by StartTime)
	first := items[0].(map[string]any)
	second := items[1].(map[string]any)
	assert.Equal(t, "Morning Session", first["name"])
	assert.Equal(t, "Afternoon Session", second["name"])
}
