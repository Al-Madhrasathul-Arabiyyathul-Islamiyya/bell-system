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

func uploadAudioFile(t *testing.T, name, fileType, filename, content, token string) (int, jsonapi.Document) {
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

	status := resp.StatusCode
	var doc jsonapi.Document
	json.NewDecoder(resp.Body).Decode(&doc)
	resp.Body.Close()
	return status, doc
}

func TestAudioFileUpload(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	status, doc := uploadAudioFile(t, "bell.mp3", "bell", "bell.mp3", "fake mp3 data", token)
	require.Equal(t, http.StatusCreated, status)

	require.NotNil(t, doc.Data)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.NotEmpty(t, doc.Data.ID)
	assert.Equal(t, "bell.mp3", attrs["name"])
	assert.Equal(t, "bell", attrs["fileType"])
	assert.NotEmpty(t, attrs["checksum"])
}

func TestAudioFileCRUD(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Upload
	status, doc := uploadAudioFile(t, "anthem.mp3", "anthem", "anthem.mp3", "anthem data", token)
	require.Equal(t, http.StatusCreated, status)

	audioID := doc.Data.ID

	// List
	resp := doRequest(t, http.MethodGet, "/api/v1/audio", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.GreaterOrEqual(t, col.Meta["total"].(float64), float64(1))

	// Get by ID (PUBLIC)
	resp = doRequest(t, http.MethodGet, "/api/v1/audio/"+audioID, nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Update metadata
	updateBody := `{"data":{"type":"audio-files","attributes":{"name":"updated-anthem.mp3","fileType":"other"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/audio/"+audioID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	updateDoc := readDocument(t, resp)
	updatedAttrs := updateDoc.Data.Attributes.(map[string]any)
	assert.Equal(t, "updated-anthem.mp3", updatedAttrs["name"])
	assert.Equal(t, "other", updatedAttrs["fileType"])

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
	status, doc := uploadAudioFile(t, "song.mp3", "school_song", "song.mp3", "song data", token)
	require.Equal(t, http.StatusCreated, status)

	attrs := doc.Data.Attributes.(map[string]any)
	expectedChecksum := attrs["checksum"].(string)

	// Get checksums
	resp := doRequest(t, http.MethodGet, "/api/v1/audio/checksums", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	require.NotEmpty(t, col.Data)

	// Find our uploaded file's checksum
	found := false
	for _, item := range col.Data {
		itemAttrs := item.Attributes.(map[string]any)
		if itemAttrs["checksum"] == expectedChecksum {
			found = true
			assert.Equal(t, "school_song", itemAttrs["fileType"])
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

func TestAudioFileList_FilterByType(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Upload files of different types
	uploadAudioFile(t, "bell1.mp3", "bell", "bell1.mp3", "data1", token)
	uploadAudioFile(t, "anthem1.mp3", "anthem", "anthem1.mp3", "data2", token)

	resp := doRequest(t, http.MethodGet, "/api/v1/audio?filter[fileType]=bell", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	for _, item := range col.Data {
		attrs := item.Attributes.(map[string]any)
		assert.Equal(t, "bell", attrs["fileType"])
	}
}

func TestAudioFileList_Pagination(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Upload several files
	for i := 0; i < 3; i++ {
		uploadAudioFile(t, fmt.Sprintf("page%d.mp3", i), "bell", fmt.Sprintf("page%d.mp3", i), fmt.Sprintf("data%d", i), token)
	}

	resp := doRequest(t, http.MethodGet, "/api/v1/audio?page[number]=1&page[size]=2", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.Len(t, col.Data, 2)
	assert.GreaterOrEqual(t, col.Meta["total"].(float64), float64(3))
	assert.Equal(t, float64(2), col.Meta["pageSize"])
}
