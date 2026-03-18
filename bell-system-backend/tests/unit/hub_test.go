package unit_test

import (
	"encoding/json"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.New("test")
	require.NoError(t, err)
	return log
}

func startHub(t *testing.T) (*ws.Hub, chan struct{}) {
	t.Helper()
	log := newTestLogger(t)
	hub := ws.NewHub(log)
	done := make(chan struct{})
	go hub.Run(done)
	t.Cleanup(func() {
		close(done)
		<-hub.Done()
	})
	return hub, done
}

func TestHub_ClientCount_Empty(t *testing.T) {
	hub, _ := startHub(t)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_ConnectedClients_Empty(t *testing.T) {
	hub, _ := startHub(t)
	clients := hub.ConnectedClients()
	assert.Empty(t, clients)
}

func TestHub_RegisterClient(t *testing.T) {
	hub, _ := startHub(t)

	client := ws.NewTestClient(hub, "admin", "Admin Panel")
	hub.RegisterClient(client)

	// Give the hub goroutine time to process
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, hub.ClientCount())
	clients := hub.ConnectedClients()
	require.Len(t, clients, 1)
	assert.Equal(t, "admin", clients[0].ClientType)
	assert.Equal(t, "Admin Panel", clients[0].ClientName)
}

func TestHub_UnregisterClient(t *testing.T) {
	hub, _ := startHub(t)

	client := ws.NewTestClient(hub, "client", "Display 1")
	hub.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	hub.UnregisterClient(client)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_Broadcast_AllClients(t *testing.T) {
	hub, _ := startHub(t)

	client1 := ws.NewTestClient(hub, "admin", "Admin")
	client2 := ws.NewTestClient(hub, "client", "Display")
	hub.RegisterClient(client1)
	hub.RegisterClient(client2)
	time.Sleep(50 * time.Millisecond)

	// Drain connected_clients messages from registration
	client1.DrainSend()
	client2.DrainSend()

	msg := models.WebSocketMessage{
		Type:      "schedules_updated",
		Timestamp: time.Now().UTC(),
	}
	hub.Broadcast(msg)
	time.Sleep(50 * time.Millisecond)

	data1 := client1.ReadSend()
	data2 := client2.ReadSend()
	require.NotNil(t, data1, "admin should receive broadcast")
	require.NotNil(t, data2, "client should receive broadcast")

	var received1 models.WebSocketMessage
	require.NoError(t, json.Unmarshal(data1, &received1))
	assert.Equal(t, "schedules_updated", received1.Type)
}

func TestHub_BroadcastToAdmins_OnlyAdmins(t *testing.T) {
	hub, _ := startHub(t)

	admin := ws.NewTestClient(hub, "admin", "Admin")
	client := ws.NewTestClient(hub, "client", "Display")
	hub.RegisterClient(admin)
	hub.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)

	// Drain the connected_clients messages sent on registration
	admin.DrainSend()
	client.DrainSend()

	msg := models.WebSocketMessage{
		Type:      "system_log",
		Timestamp: time.Now().UTC(),
		Payload:   models.SystemLogPayload{Level: "info", Message: "test", Source: "hub"},
	}
	hub.BroadcastToAdmins(msg)
	time.Sleep(50 * time.Millisecond)

	data := admin.ReadSend()
	require.NotNil(t, data, "admin should receive admin-only broadcast")

	clientData := client.ReadSend()
	assert.Nil(t, clientData, "non-admin should NOT receive admin-only broadcast")
}

func TestHub_ConnectedClients_BroadcastOnRegister(t *testing.T) {
	hub, _ := startHub(t)

	admin := ws.NewTestClient(hub, "admin", "Admin")
	hub.RegisterClient(admin)
	time.Sleep(50 * time.Millisecond)

	// Admin should receive connected_clients on their own registration
	data := admin.ReadSend()
	require.NotNil(t, data)

	var msg models.WebSocketMessage
	require.NoError(t, json.Unmarshal(data, &msg))
	assert.Equal(t, "connected_clients", msg.Type)
}

func TestHub_GracefulShutdown(t *testing.T) {
	log := newTestLogger(t)
	hub := ws.NewHub(log)
	done := make(chan struct{})
	go hub.Run(done)

	client := ws.NewTestClient(hub, "client", "Display")
	hub.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)

	close(done)
	select {
	case <-hub.Done():
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("hub did not shut down in time")
	}

	assert.Equal(t, 0, hub.ClientCount())
}
