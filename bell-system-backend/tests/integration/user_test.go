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

func TestUserList_Pagination(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Create extra users to ensure we have enough for pagination
	for i := 0; i < 3; i++ {
		createTestUser(t, fmt.Sprintf("pageuser%d", i), "securepass123", "admin")
	}

	// Request page 1 with size 2
	resp := doRequest(t, http.MethodGet, "/api/v1/users?page[number]=1&page[size]=2", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	assert.Len(t, col.Data, 2)
	assert.GreaterOrEqual(t, col.Meta["total"].(float64), float64(4))
	assert.Equal(t, float64(2), col.Meta["pageSize"])
	assert.Equal(t, float64(1), col.Meta["page"])
}

func TestUserList_FilterByRole(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	// Create users with different roles
	createTestUser(t, "filteradmin", "securepass123", "admin")
	createTestUser(t, "filtermorning", "securepass123", "morning_user")

	resp := doRequest(t, http.MethodGet, "/api/v1/users?filter[role]=morning_user", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	for _, item := range col.Data {
		attrs := item.Attributes.(map[string]any)
		assert.Equal(t, "morning_user", attrs["role"])
	}
}

func TestUserList_SortDescending(t *testing.T) {
	cleanAndSeed(t)
	token := adminToken(t)

	resp := doRequest(t, http.MethodGet, "/api/v1/users?sort=-username", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	col := readCollection(t, resp)
	require.GreaterOrEqual(t, len(col.Data), 2)

	// Verify descending order
	for i := 1; i < len(col.Data); i++ {
		prev := col.Data[i-1].Attributes.(map[string]any)["username"].(string)
		curr := col.Data[i].Attributes.(map[string]any)["username"].(string)
		assert.GreaterOrEqual(t, prev, curr, "expected descending username order")
	}
}
