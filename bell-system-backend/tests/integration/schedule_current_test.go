//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleCurrent_NoSession(t *testing.T) {
	cleanAndSeed(t)

	// Delete all sessions so no current session exists
	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult struct {
		Total int `json:"total"`
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	readJSON(t, resp, &listResult)

	for _, s := range listResult.Items {
		resp = doRequest(t, http.MethodDelete, "/api/v1/sessions/"+s.ID, nil, "")
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

	soundID := createTestAudioFile(t)
	sessionID := getSessionID(t)

	// Create a schedule item for today
	now := "12:00" // use a time that works — items returned regardless of past/future
	dayOfWeek := getDayOfWeek()

	createBody := fmt.Sprintf(
		`{"sessionId":"%s","name":"Test Current Bell","time":"%s","soundId":"%s","days":[%d]}`,
		sessionID, now, soundID, dayOfWeek,
	)
	resp := doRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(createBody), "")
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
