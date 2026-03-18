//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wsURL returns the WebSocket URL for the test server.
func wsURL(token, clientType, clientName string) string {
	base := strings.Replace(testServer.URL, "http://", "ws://", 1)
	url := fmt.Sprintf("%s/ws?token=%s&client_type=%s", base, token, clientType)
	if clientName != "" {
		url += "&client_name=" + clientName
	}
	return url
}

// dialWS connects to the WebSocket endpoint and reads the connection_acknowledged message.
func dialWS(t *testing.T, token, clientType, clientName string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(token, clientType, clientName), nil)
	require.NoError(t, err)

	// Read connection_acknowledged
	var ack models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &ack)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", ack.Type)

	return conn
}

// readWSMessage reads the next WebSocket message with a timeout.
func readWSMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) *models.WebSocketMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var msg models.WebSocketMessage
	err := wsjson.Read(ctx, conn, &msg)
	if err != nil {
		if ctx.Err() != nil {
			return nil // timeout — no message available
		}
		t.Fatalf("unexpected WebSocket read error: %v", err)
	}
	return &msg
}

func TestWebSocket_ConnectionAcknowledged(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsadmin", "wspass123", "admin")
	token := loginAs(t, "wsadmin", "wspass123")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL(token, "admin", "TestAdmin"), nil)
	require.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	var msg models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", msg.Type)

	payloadBytes, _ := json.Marshal(msg.Payload)
	var ack models.ConnectionAcknowledgedPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &ack))
	assert.Equal(t, "admin", ack.ClientType)
	assert.Equal(t, "Connection successful", ack.Message)
	assert.NotEmpty(t, ack.ConnectionID)
}

func TestWebSocket_MissingToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL("", "client", ""), nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocket_InvalidToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL("invalid-token", "client", ""), nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocket_InvalidClientType(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsadmin2", "wspass123", "admin")
	token := loginAs(t, "wsadmin2", "wspass123")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL(token, "invalid", ""), nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWebSocket_AdminTypeRequiresAdminRole(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsmorning", "wspass123", "morning_user")
	token := loginAs(t, "wsmorning", "wspass123")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, wsURL(token, "admin", ""), nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestWebSocket_AdminReceivesConnectedClients(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsadmin3", "wspass123", "admin")
	createTestUser(t, "wsmorning2", "wspass123", "morning_user")

	adminToken := loginAs(t, "wsadmin3", "wspass123")

	adminConn := dialWS(t, adminToken, "admin", "Admin")
	defer adminConn.Close(websocket.StatusNormalClosure, "")

	// Admin gets connected_clients for their own registration
	msg := readWSMessage(t, adminConn, 2*time.Second)
	require.NotNil(t, msg, "admin should receive connected_clients")
	assert.Equal(t, "connected_clients", msg.Type)

	// Connect a second client
	clientToken := loginAs(t, "wsmorning2", "wspass123")
	clientConn := dialWS(t, clientToken, "client", "Display")
	defer clientConn.Close(websocket.StatusNormalClosure, "")

	// Admin should receive updated connected_clients
	msg = readWSMessage(t, adminConn, 2*time.Second)
	require.NotNil(t, msg, "admin should receive connected_clients on new client connect")
	assert.Equal(t, "connected_clients", msg.Type)

	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload models.ConnectedClientsPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))
	assert.GreaterOrEqual(t, len(payload.Clients), 2, "should show at least 2 connected clients")
}

func TestWebSocket_ScheduleUpdateNotifiesClients(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsadmin4", "wspass123", "admin")
	createTestUser(t, "wsmorning3", "wspass123", "morning_user")

	adminToken := loginAs(t, "wsadmin4", "wspass123")

	// Connect a WebSocket client
	clientToken := loginAs(t, "wsmorning3", "wspass123")
	clientConn := dialWS(t, clientToken, "client", "Display")
	defer clientConn.Close(websocket.StatusNormalClosure, "")

	// Give hub time to register the client
	time.Sleep(100 * time.Millisecond)

	// Create an audio file first (needed for schedule item) — this triggers audio_files_updated
	audioID := createTestAudioFileForWS(t, adminToken)

	// Drain the audio_files_updated notification
	msg := readWSMessage(t, clientConn, 2*time.Second)
	require.NotNil(t, msg, "should receive audio_files_updated from audio file creation")
	assert.Equal(t, "audio_files_updated", msg.Type)

	// Create a schedule item via REST — should trigger schedules_updated
	sessionID := getSessionIDForWS(t, adminToken)
	body := fmt.Sprintf(`{"name":"WS Test Bell","time":"08:00","soundId":"%s","sessionId":"%s","days":[2,3,4,5,6]}`, audioID, sessionID)
	resp := doRequest(t, http.MethodPost, "/api/v1/schedule", bytes.NewBufferString(body), adminToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Client should receive schedules_updated
	msg = readWSMessage(t, clientConn, 2*time.Second)
	require.NotNil(t, msg, "client should receive schedules_updated notification")
	assert.Equal(t, "schedules_updated", msg.Type)
}

func TestWebSocket_AudioUpdateNotifiesClients(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsmorning4", "wspass123", "morning_user")
	token := adminToken(t)

	// Connect a WebSocket client
	clientToken := loginAs(t, "wsmorning4", "wspass123")
	clientConn := dialWS(t, clientToken, "client", "Display")
	defer clientConn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(100 * time.Millisecond)

	// Upload an audio file via REST — should trigger audio_files_updated
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("name", "ws-test-bell.mp3")
	w.WriteField("type", "bell")
	part, err := w.CreateFormFile("file", "ws-test-bell.mp3")
	require.NoError(t, err)
	part.Write([]byte("fake audio data for ws test"))
	w.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/audio", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Client should receive audio_files_updated
	msg := readWSMessage(t, clientConn, 2*time.Second)
	require.NotNil(t, msg, "client should receive audio_files_updated notification")
	assert.Equal(t, "audio_files_updated", msg.Type)
}

func TestWebSocket_Heartbeat(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsmorning5", "wspass123", "morning_user")
	token := loginAs(t, "wsmorning5", "wspass123")

	conn := dialWS(t, token, "client", "Display")
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send heartbeat
	heartbeat := models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
	err := wsjson.Write(ctx, conn, heartbeat)
	require.NoError(t, err)

	// Connection should still be alive — send another heartbeat
	time.Sleep(100 * time.Millisecond)
	err = wsjson.Write(ctx, conn, heartbeat)
	require.NoError(t, err, "connection should still be alive after heartbeat")
}

func TestWebSocket_ClientTypeClient(t *testing.T) {
	cleanAndSeed(t)
	createTestUser(t, "wsmorning6", "wspass123", "morning_user")
	token := loginAs(t, "wsmorning6", "wspass123")

	conn := dialWS(t, token, "client", "MorningDisplay")
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Verify connected — connection_acknowledged was already consumed by dialWS
	// Send heartbeat to verify connection works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := wsjson.Write(ctx, conn, models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	})
	require.NoError(t, err)
}

// createTestAudioFileForWS is a helper that creates an audio file via REST API.
func createTestAudioFileForWS(t *testing.T, token string) string {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("name", "ws-schedule-bell.mp3")
	w.WriteField("type", "bell")
	part, err := w.CreateFormFile("file", "ws-schedule-bell.mp3")
	require.NoError(t, err)
	part.Write([]byte("fake audio"))
	w.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/audio", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	readJSON(t, resp, &result)
	return result["id"].(string)
}

// getSessionIDForWS returns the first session ID.
func getSessionIDForWS(t *testing.T, token string) string {
	t.Helper()

	resp := doRequest(t, http.MethodGet, "/api/v1/sessions", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	readJSON(t, resp, &result)
	items := result["items"].([]any)
	require.NotEmpty(t, items)
	return items[0].(map[string]any)["id"].(string)
}
