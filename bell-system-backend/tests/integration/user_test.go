//go:build integration

package integration_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCRUD(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Create
	createBody := `{"data":{"type":"users","attributes":{"username":"testuser","password":"securepass123","role":"admin"}}}`
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/users", bytes.NewBufferString(createBody), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	doc := readDocument(t, resp)
	userID := doc.Data.ID
	require.NotEmpty(t, userID)
	attrs := doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "testuser", attrs["username"])
	assert.Equal(t, "admin", attrs["role"])

	// Get by ID
	resp = doRequest(t, http.MethodGet, "/api/v1/users/"+userID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "testuser", attrs["username"])

	// List
	resp = doRequest(t, http.MethodGet, "/api/v1/users", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	total := col.Meta["total"].(float64)
	assert.GreaterOrEqual(t, total, float64(4)) // 3 seeded + 1 created

	// Update
	updateBody := `{"data":{"type":"users","attributes":{"username":"updateduser"}}}`
	resp = doJSONAPIRequest(t, http.MethodPut, "/api/v1/users/"+userID, bytes.NewBufferString(updateBody), token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	doc = readDocument(t, resp)
	attrs = doc.Data.Attributes.(map[string]any)
	assert.Equal(t, "updateduser", attrs["username"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/v1/users/"+userID, nil, token)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/v1/users/"+userID, nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// "admin" already exists from seed data
	body := `{"data":{"type":"users","attributes":{"username":"admin","password":"password123","role":"admin"}}}`
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/users", bytes.NewBufferString(body), token)
	defer resp.Body.Close()

	// Should fail with conflict or server error
	assert.True(t, resp.StatusCode >= 400,
		fmt.Sprintf("expected error status, got %d", resp.StatusCode))
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	body := `{"data":{"type":"users","attributes":{"username":"shortpw","password":"short","role":"admin"}}}`
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/users", bytes.NewBufferString(body), token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateUser_InvalidRole(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	body := `{"data":{"type":"users","attributes":{"username":"badrole","password":"password123","role":"superadmin"}}}`
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/users", bytes.NewBufferString(body), token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
