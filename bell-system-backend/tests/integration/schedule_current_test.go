//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleCurrent_NoSession(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Delete all sessions so no current session exists
	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	for _, s := range col.Data {
		resp = doRequest(t, http.MethodDelete, "/api/v1/sessions/"+s.ID, nil, token)
		resp.Body.Close()
	}

	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/current", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result models.CurrentScheduleResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Nil(t, result.Session)
	assert.Empty(t, result.Items)
}

func TestScheduleCurrent_WithItems(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Create a session that covers the current time (so GetCurrentSession finds it)
	now := time.Now()
	startTime := fmt.Sprintf("%02d:%02d", now.Hour(), 0)
	endHour := now.Hour() + 1
	if endHour > 23 {
		endHour = 23
	}
	endTime := fmt.Sprintf("%02d:%02d", endHour, 59)

	sessionBody := fmt.Sprintf(`{"data":{"type":"sessions","attributes":{"name":"Test Now Session","startTime":"%s","endTime":"%s"}}}`, startTime, endTime)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(sessionBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	sessionDoc := readDocument(t, resp)
	sessionID := sessionDoc.Data.ID

	soundID := createTestAudioFile(t, token)
	dayOfWeek := getDayOfWeek()

	// Create a schedule item for today within the session
	bellTime := fmt.Sprintf("%02d:30", now.Hour())
	createBody := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Test Current Bell","time":"%s","days":[%d]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		bellTime, dayOfWeek, soundID, sessionID,
	)
	resp = doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(createBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Fetch current schedule
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/current", nil, "")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result models.CurrentScheduleResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotNil(t, result.Session)
	assert.Equal(t, "Test Now Session", result.Session.Name)
	require.NotEmpty(t, result.Items)

	// Find our item
	found := false
	for _, item := range result.Items {
		if item.Name == "Test Current Bell" {
			found = true
			assert.Contains(t, []string{"pending", "current", "completed"}, item.Status)
			break
		}
	}
	assert.True(t, found, "expected to find 'Test Current Bell' in current schedule")
}
