//go:build integration

package integration_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestAudioFile uploads a test audio file and returns its ID.
func createTestAudioFile(t *testing.T) string {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("name", "test-bell.mp3")
	w.WriteField("type", "bell")
	part, err := w.CreateFormFile("file", "test-bell.mp3")
	require.NoError(t, err)
	part.Write([]byte("fake audio data"))
	w.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/audio-files", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	readJSON(t, resp, &result)
	id, ok := result["id"].(string)
	require.True(t, ok)
	return id
}

// getSessionID returns the ID of the first session from the list.
func getSessionID(t *testing.T) string {
	t.Helper()

	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	readJSON(t, resp, &result)
	items := result["items"].([]any)
	require.NotEmpty(t, items)
	return items[0].(map[string]any)["id"].(string)
}

func TestScheduleItemCRUD(t *testing.T) {
	cleanAndSeed(t)

	soundID := createTestAudioFile(t)
	sessionID := getSessionID(t)

	// Create
	createBody := fmt.Sprintf(
		`{"sessionId":"%s","name":"First Bell","time":"07:00","soundId":"%s","days":[1,2,3,4,5]}`,
		sessionID, soundID,
	)
	resp := doRequest(t, http.MethodPost, "/api/v1/schedule-items", bytes.NewBufferString(createBody), "")
	if resp.StatusCode != http.StatusCreated {
		var errBody map[string]any
		readJSON(t, resp, &errBody)
		t.Fatalf("schedule item create failed with status %d: %v", resp.StatusCode, errBody)
	}

	var created map[string]any
	readJSON(t, resp, &created)

	itemID, ok := created["id"].(string)
	require.True(t, ok)
	assert.Equal(t, "First Bell", created["name"])

	// Get by ID — should include relations
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule-items/"+itemID, nil, "")
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		readJSON(t, resp, &errBody)
		t.Fatalf("schedule item get by ID failed with status %d: %v", resp.StatusCode, errBody)
	}

	var fetched map[string]any
	readJSON(t, resp, &fetched)
	assert.Equal(t, "First Bell", fetched["name"])

	// Verify days are present
	days, ok := fetched["days"].([]any)
	require.True(t, ok)
	assert.Len(t, days, 5)

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule-items", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult map[string]any
	readJSON(t, resp, &listResult)
	assert.GreaterOrEqual(t, listResult["total"].(float64), float64(1))

	// Update
	updateBody := `{"name":"Updated Bell"}`
	resp = doRequest(t, http.MethodPut, "/api/v1/schedule-items/"+itemID, bytes.NewBufferString(updateBody), "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated map[string]any
	readJSON(t, resp, &updated)
	assert.Equal(t, "Updated Bell", updated["name"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/schedule-items/"+itemID, nil, "")
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule-items/"+itemID, nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestScheduleItem_InvalidDays(t *testing.T) {
	cleanAndSeed(t)

	soundID := createTestAudioFile(t)

	body := fmt.Sprintf(
		`{"name":"Bad Days","time":"08:00","soundId":"%s","days":[0,8]}`, soundID,
	)
	resp := doRequest(t, http.MethodPost, "/api/v1/schedule-items", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestScheduleItem_InvalidTime(t *testing.T) {
	cleanAndSeed(t)

	soundID := createTestAudioFile(t)

	body := fmt.Sprintf(
		`{"name":"Bad Time","time":"not-a-time","soundId":"%s","days":[1]}`, soundID,
	)
	resp := doRequest(t, http.MethodPost, "/api/v1/schedule-items", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
