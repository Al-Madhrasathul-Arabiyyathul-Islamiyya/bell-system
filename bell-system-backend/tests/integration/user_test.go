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

	// Create
	createBody := `{"username":"testuser","password":"securepass123","role":"admin"}`
	resp := doRequest(t, http.MethodPost, "/api/users", bytes.NewBufferString(createBody), "")
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	readJSON(t, resp, &created)

	userID, ok := created["id"].(string)
	require.True(t, ok)
	assert.Equal(t, "testuser", created["username"])
	assert.Equal(t, "admin", created["role"])

	// Get by ID
	resp = doRequest(t, http.MethodGet, "/api/users/"+userID, nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var fetched map[string]any
	readJSON(t, resp, &fetched)
	assert.Equal(t, "testuser", fetched["username"])

	// List
	resp = doRequest(t, http.MethodGet, "/api/users", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult map[string]any
	readJSON(t, resp, &listResult)
	total := listResult["total"].(float64)
	assert.GreaterOrEqual(t, total, float64(4)) // 3 seeded + 1 created

	// Update
	updateBody := `{"username":"updateduser"}`
	resp = doRequest(t, http.MethodPut, "/api/users/"+userID, bytes.NewBufferString(updateBody), "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated map[string]any
	readJSON(t, resp, &updated)
	assert.Equal(t, "updateduser", updated["username"])

	// Delete
	resp = doRequest(t, http.MethodDelete, "/api/users/"+userID, nil, "")
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, http.MethodGet, "/api/users/"+userID, nil, "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	cleanAndSeed(t)

	// "admin" already exists from seed data
	body := `{"username":"admin","password":"password123","role":"admin"}`
	resp := doRequest(t, http.MethodPost, "/api/users", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	// Should fail with conflict or server error
	assert.True(t, resp.StatusCode >= 400,
		fmt.Sprintf("expected error status, got %d", resp.StatusCode))
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	cleanAndSeed(t)

	body := `{"username":"shortpw","password":"short","role":"admin"}`
	resp := doRequest(t, http.MethodPost, "/api/users", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateUser_InvalidRole(t *testing.T) {
	cleanAndSeed(t)

	body := `{"username":"badrole","password":"password123","role":"superadmin"}`
	resp := doRequest(t, http.MethodPost, "/api/users", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
