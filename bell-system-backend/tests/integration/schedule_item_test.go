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
	assert.Equal(t, "07:00", attrs["time"])

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
	assert.Equal(t, "07:00", attrs["time"])

	// Verify days are present
	days, ok := attrs["days"].([]any)
	require.True(t, ok)
	assert.Len(t, days, 5)

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.GreaterOrEqual(t, col.Meta["total"].(float64), float64(1))
	foundCreatedItem := false
	for _, item := range col.Data {
		if item.ID != itemID {
			continue
		}

		foundCreatedItem = true
		listAttrs := item.Attributes.(map[string]any)
		assert.Equal(t, "07:00", listAttrs["time"])
		break
	}
	assert.True(t, foundCreatedItem, "expected created schedule item in collection")

	// Update
	updateBody := `{"data":{"type":"schedule-items","attributes":{"name":"Updated Bell","time":"07:15"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/schedule/"+itemID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "Updated Bell", attrs["name"])
	assert.Equal(t, "07:15", attrs["time"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/schedule/"+itemID, nil, token)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/"+itemID, nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestScheduleItem_TimePersistenceRoundTrip(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	createBody := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Time Persistence Bell","time":"13:25","days":[2,3,4]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(createBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	doc := readDocument(t, resp)
	itemID := doc.Data.ID
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "13:25", attrs["time"])

	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/"+itemID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "13:25", attrs["time"])

	resp = doRequest(t, http.MethodGet, "/api/v1/schedule?filter[sessionId]="+sessionID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	found := false
	for _, item := range col.Data {
		if item.ID != itemID {
			continue
		}

		found = true
		listAttrs := item.Attributes.(map[string]any)
		assert.Equal(t, "13:25", listAttrs["time"])
		break
	}
	assert.True(t, found, "expected time persistence item in filtered schedule list")

	updateBody := `{"data":{"type":"schedule-items","attributes":{"time":"13:40"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/schedule/"+itemID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "13:40", attrs["time"])

	resp = doRequest(t, http.MethodGet, "/api/v1/schedule/"+itemID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "13:40", attrs["time"])
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

func TestScheduleItemList_FilterBySession(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	// Create an item in the session
	body := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Filter Test","time":"09:00","days":[1]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Filter by session
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule?filter[sessionId]="+sessionID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.GreaterOrEqual(t, len(col.Data), 1)
}

func TestScheduleItemList_IncludeSession(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	// Create an item
	body := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Include Test","time":"10:00","days":[1,2]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List with include=session
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule?include=session", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	require.NotEmpty(t, col.Data)
	require.NotEmpty(t, col.Included, "expected included resources for session")

	// Verify included contains a session resource
	foundSession := false
	for _, inc := range col.Included {
		if inc.Type == "sessions" {
			foundSession = true
			break
		}
	}
	assert.True(t, foundSession, "expected sessions in included resources")
}

func TestScheduleItemList_IncludeSound(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	soundID := createTestAudioFile(t, token)
	sessionID := getSessionID(t, token)

	// Create an item
	body := fmt.Sprintf(
		`{"data":{"type":"schedule-items","attributes":{"name":"Sound Include","time":"11:00","days":[3]},"relationships":{"sound":{"data":{"type":"audio-files","id":"%s"}},"session":{"data":{"type":"sessions","id":"%s"}}}}}`,
		soundID, sessionID,
	)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List with include=sound
	resp = doRequest(t, http.MethodGet, "/api/v1/schedule?include=sound", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	require.NotEmpty(t, col.Data)
	require.NotEmpty(t, col.Included, "expected included resources for sound")

	foundSound := false
	for _, inc := range col.Included {
		if inc.Type == "audio-files" {
			foundSound = true
			break
		}
	}
	assert.True(t, foundSound, "expected audio-files in included resources")
}
