//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestAudioFile uploads a test audio file and returns its ID.
func createTestAudioFile(t *testing.T, token string) string {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("name", "test-bell.mp3")
	w.WriteField("type", "bell")
	part, err := w.CreateFormFile("file", "test-bell.mp3")
	require.NoError(t, err)
	part.Write([]byte("fake audio data"))
	w.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/audio", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var doc jsonapi.Document
	json.NewDecoder(resp.Body).Decode(&doc)
	resp.Body.Close()
	require.NotNil(t, doc.Data)
	return doc.Data.ID
}

// getSessionID returns the ID of the first session from the list.
func getSessionID(t *testing.T, token string) string {
	t.Helper()

	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	require.NotEmpty(t, col.Data)
	return col.Data[0].ID
}

func TestScheduleItemCRUD(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	// Create
	createBody := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"First Bell","time":"07:00","days":[1,2,3,4,5]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(createBody), token)
	if resp.StatusCode != http.StatusCreated {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		resp.Body.Close()
		t.Fatalf("schedule item create failed with status %d: %v", resp.StatusCode, errBody)
	}

	doc := readDocument(t, resp)
	itemID := doc.Data.ID
	require.NotEmpty(t, itemID)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "First Bell", attrs["name"])

	// Get by ID — should include relations
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/"+itemID, nil, token)
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		resp.Body.Close()
		t.Fatalf("schedule item get by ID failed with status %d: %v", resp.StatusCode, errBody)
	}

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "First Bell", attrs["name"])

	// Verify days are present
	days, ok := attrs["days"].([]any)
	require.True(t, ok)
	assert.Len(t, days, 5)

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.GreaterOrEqual(t, col.Meta["total"].(float64), float64(1))

	// Update
	updateBody := `{"data":{"type":"schedule-items","attributes":{"name":"Updated Bell"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/schedule/"+itemID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "Updated Bell", attrs["name"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/schedule/"+itemID, nil, token)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/"+itemID, nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestScheduleItem_InvalidDays(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	body := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Bad Days","time":"08:00","days":[0,8]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestScheduleItem_InvalidTime(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	body := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Bad Time","time":"not-a-time","days":[1]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
