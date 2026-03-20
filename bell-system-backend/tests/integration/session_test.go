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
	createBody := `{"data":{"type":"sessions","attributes":{"name":"Evening Session","startTime":"18:30","endTime":"21:00"}}}`
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(createBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	doc := readDocument(t, resp)
	sessionID := doc.Data.ID
	require.NotEmpty(t, sessionID)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "Evening Session", attrs["name"])

	// Get by ID
	resp = doRequest(t, http.MethodGet, "/api/v1/sessions/"+sessionID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "Evening Session", attrs["name"])

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	total := col.Meta["total"].(float64)
	assert.GreaterOrEqual(t, total, float64(3)) // 2 seeded + 1 created

	// Update
	updateBody := `{"data":{"type":"sessions","attributes":{"name":"Late Evening Session"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/sessions/"+sessionID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "Late Evening Session", attrs["name"])

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

	col := readCollection(t, resp)
	require.GreaterOrEqual(t, len(col.Data), 2)

	// Verify Morning comes before Afternoon (ordered by StartTime)
	firstAttrs := col.Data[0].Attributes.(map[string]any)
	secondAttrs := col.Data[1].Attributes.(map[string]any)
	assert.Equal(t, "Morning Session", firstAttrs["name"])
	assert.Equal(t, "Afternoon Session", secondAttrs["name"])
}
