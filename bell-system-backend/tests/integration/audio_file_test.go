//go:build integration

package integration_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uploadAudioFile(t *testing.T, name, fileType, filename, content, token string) (int, map[string]any) {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("name", name)
	w.WriteField("type", fileType)
	part, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)
	part.Write([]byte(content))
	w.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/audio", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var result map[string]any
	readJSON(t, resp, &result)
	return resp.StatusCode, result
}

func TestAudioFileUpload(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	status, result := uploadAudioFile(t, "bell.mp3", "bell", "bell.mp3", "fake mp3 data", token)
	require.Equal(t, http.StatusCreated, status)

	assert.NotEmpty(t, result["id"])
	assert.Equal(t, "bell.mp3", result["name"])
	assert.Equal(t, "bell", result["fileType"])
	assert.NotEmpty(t, result["checksum"])
}

func TestAudioFileCRUD(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Upload
	status, created := uploadAudioFile(t, "anthem.mp3", "anthem", "anthem.mp3", "anthem data", token)
	require.Equal(t, http.StatusCreated, status)

	audioID := created["id"].(string)

	// List
	resp := doRequest(t, http.MethodGet, "/api/v1/audio", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult map[string]any
	readJSON(t, resp, &listResult)
	assert.GreaterOrEqual(t, listResult["total"].(float64), float64(1))

	// Get by ID (PUBLIC)
	resp = doRequest(t, http.MethodGet, "/api/v1/audio/"+audioID, nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Update metadata
	updateBody := `{"name":"updated-anthem.mp3","fileType":"other"}`
	resp = doRequest(t, http.MethodPut, "/api/v1/audio/"+audioID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated map[string]any
	readJSON(t, resp, &updated)
	assert.Equal(t, "updated-anthem.mp3", updated["name"])
	assert.Equal(t, "other", updated["fileType"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/audio/"+audioID, nil, token)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted (GET by ID is PUBLIC)
	resp = doRequest(t, http.MethodGet, "/api/v1/audio/"+audioID, nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestAudioFileChecksums(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Upload a file
	status, created := uploadAudioFile(t, "song.mp3", "school_song", "song.mp3", "song data", token)
	require.Equal(t, http.StatusCreated, status)

	expectedChecksum := created["checksum"].(string)

	// Get checksums
	resp := doRequest(t, http.MethodGet, "/api/v1/audio/checksums", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var checksums []map[string]any
	readJSON(t, resp, &checksums)

	require.NotEmpty(t, checksums)

	// Find our uploaded file's checksum
	found := false
	for _, entry := range checksums {
		if entry["checksum"] == expectedChecksum {
			found = true
			assert.Equal(t, "school_song", entry["type"])
			break
		}
	}
	assert.True(t, found, "expected checksum not found in checksums list")
}

func TestAudioFileUpload_InvalidType(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	status, _ := uploadAudioFile(t, "bad.mp3", "invalid_type", "bad.mp3", "data", token)
	assert.Equal(t, http.StatusBadRequest, status)
}
