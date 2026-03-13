package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWSTest creates an httptest.Server that upgrades connections, creates a Client,
// and runs ReadPump/WritePump. Returns the client-side conn, the server-side Client, and the Hub.
func setupWSTest(t *testing.T) (*websocket.Conn, *ws.Client, *ws.Hub) {
	t.Helper()

	log, err := logger.New("test")
	require.NoError(t, err)

	hub := ws.NewHub(log)
	done := make(chan struct{})
	go hub.Run(done)
	t.Cleanup(func() {
		close(done)
		<-hub.Done()
	})

	cfg := ws.TestWSConfig()

	var client *ws.Client
	clientReady := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		client = ws.NewClient(conn, hub, log, cfg, r.RemoteAddr, "client", "TestClient", [16]byte{})
		hub.RegisterClient(client)
		close(clientReady)

		ctx := r.Context()
		go client.WritePump(ctx)
		client.ReadPump(ctx)
	}))
	t.Cleanup(srv.Close)

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.CloseNow() })

	// Wait for server-side client to be ready
	select {
	case <-clientReady:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client to be created")
	}

	return conn, client, hub
}

func TestClient_ReadPump_ProcessesHeartbeat(t *testing.T) {
	conn, _, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send a heartbeat message
	msg := models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
	err := wsjson.Write(ctx, conn, msg)
	require.NoError(t, err)

	// The connection should remain alive — verify by sending another message
	time.Sleep(100 * time.Millisecond)
	err = wsjson.Write(ctx, conn, msg)
	assert.NoError(t, err)
}

func TestClient_ReadPump_ProcessesRegister(t *testing.T) {
	conn, client, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send a register message with a new client name
	msg := models.WebSocketMessage{
		Type:      "register",
		Timestamp: time.Now().UTC(),
		Payload: models.RegistrationPayload{
			ClientName: "UpdatedName",
		},
	}
	err := wsjson.Write(ctx, conn, msg)
	require.NoError(t, err)

	// Should receive connection_acknowledged back
	var ack models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &ack)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", ack.Type)

	// Verify client name was updated
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "UpdatedName", client.ClientName())
}

func TestClient_ReadPump_ClosesOnReadError(t *testing.T) {
	conn, client, hub := setupWSTest(t)

	// Close the client-side connection abruptly
	conn.CloseNow()

	// Wait for the client to be unregistered from the hub
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for client to be unregistered")
		default:
			// Check if client is still in the hub
			clients := hub.ConnectedClients()
			found := false
			for _, c := range clients {
				if c.ID == client.ID {
					found = true
					break
				}
			}
			if !found {
				return // success: client was unregistered
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func TestClient_ReadPump_IgnoresUnknownMessageTypes(t *testing.T) {
	conn, _, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send an unknown message type
	msg := models.WebSocketMessage{
		Type:      "unknown_type",
		Timestamp: time.Now().UTC(),
	}
	err := wsjson.Write(ctx, conn, msg)
	require.NoError(t, err)

	// Connection should stay alive — verify by sending a heartbeat
	time.Sleep(100 * time.Millisecond)
	heartbeat := models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
	err = wsjson.Write(ctx, conn, heartbeat)
	assert.NoError(t, err)
}

func TestClient_WritePump_DeliversQueuedMessages(t *testing.T) {
	conn, client, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Queue a message via SendMessage
	payload := models.WebSocketMessage{
		Type:      "test_event",
		Timestamp: time.Now().UTC(),
		Payload:   map[string]string{"key": "value"},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	client.SendMessage(data)

	// Client should receive it
	var received models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &received)
	require.NoError(t, err)
	assert.Equal(t, "test_event", received.Type)
}

func TestClient_WritePump_SendsPingsPeriodically(t *testing.T) {
	// With PingInterval=1s from TestWSConfig, we should see a ping within ~1.5s
	conn, _, _ := setupWSTest(t)

	// Set a read deadline to detect that the connection is still alive after ping
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Wait for at least one ping cycle
	time.Sleep(1500 * time.Millisecond)

	// The connection should still be alive — send a message to verify
	msg := models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
	err := wsjson.Write(ctx, conn, msg)
	assert.NoError(t, err)
}

func TestClient_SendMessage_DropsWhenBufferFull(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)

	hub := ws.NewHub(log)
	done := make(chan struct{})
	go hub.Run(done)
	defer func() {
		close(done)
		<-hub.Done()
	}()

	// Create a test client (no real connection) to test SendMessage buffering
	client := ws.NewTestClient(hub, "client", "BufferTest")

	// Fill the send buffer (256 capacity)
	for i := 0; i < 256; i++ {
		client.SendMessage([]byte("msg"))
	}

	// This should not block — it drops the message
	doneCh := make(chan struct{})
	go func() {
		client.SendMessage([]byte("overflow"))
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// success — SendMessage returned without blocking
	case <-time.After(1 * time.Second):
		t.Fatal("SendMessage blocked when buffer was full")
	}
}

func TestClient_HandleRegister_SendsConnectionAcknowledged(t *testing.T) {
	conn, _, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send register message
	msg := models.WebSocketMessage{
		Type:      "register",
		Timestamp: time.Now().UTC(),
		Payload: models.RegistrationPayload{
			ClientName: "NewDisplayUnit",
		},
	}
	err := wsjson.Write(ctx, conn, msg)
	require.NoError(t, err)

	// Read the ack
	var ack models.WebSocketMessage
	err = wsjson.Read(ctx, conn, &ack)
	require.NoError(t, err)
	assert.Equal(t, "connection_acknowledged", ack.Type)

	// Parse the payload
	payloadBytes, err := json.Marshal(ack.Payload)
	require.NoError(t, err)
	var ackPayload models.ConnectionAcknowledgedPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &ackPayload))
	assert.Equal(t, "client", ackPayload.ClientType)
	assert.Equal(t, "Registration updated", ackPayload.Message)
	assert.NotEmpty(t, ackPayload.ConnectionID)
}

func TestClient_ReadPump_InvalidJSON(t *testing.T) {
	conn, _, _ := setupWSTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send invalid JSON
	err := conn.Write(ctx, websocket.MessageText, []byte("not valid json"))
	require.NoError(t, err)

	// Connection should stay alive — send a valid heartbeat after
	time.Sleep(100 * time.Millisecond)
	msg := models.WebSocketMessage{
		Type:      "heartbeat",
		Timestamp: time.Now().UTC(),
	}
	err = wsjson.Write(ctx, conn, msg)
	assert.NoError(t, err)
}
