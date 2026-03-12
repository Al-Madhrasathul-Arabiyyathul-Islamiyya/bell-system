package unit_test

import (
	"encoding/json"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNotifier(t *testing.T) (*ws.Notifier, *ws.Client, *ws.Client) {
	t.Helper()
	hub, _ := startHub(t)

	admin := ws.NewTestClient(hub, "admin", "Admin")
	client := ws.NewTestClient(hub, "client", "Display")
	hub.RegisterClient(admin)
	hub.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)
	admin.DrainSend()
	client.DrainSend()

	notifier := ws.NewNotifier(hub)
	return notifier, admin, client
}

func TestNotifier_SchedulesUpdated(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	notifier.NotifySchedulesUpdated()
	time.Sleep(50 * time.Millisecond)

	for _, c := range []*ws.Client{admin, client} {
		data := c.ReadSend()
		require.NotNil(t, data)
		var msg models.WebSocketMessage
		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Equal(t, "schedules_updated", msg.Type)
	}
}

func TestNotifier_AudioFilesUpdated(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	notifier.NotifyAudioFilesUpdated()
	time.Sleep(50 * time.Millisecond)

	for _, c := range []*ws.Client{admin, client} {
		data := c.ReadSend()
		require.NotNil(t, data)
		var msg models.WebSocketMessage
		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Equal(t, "audio_files_updated", msg.Type)
	}
}

func TestNotifier_BellTriggered(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	payload := models.BellTriggeredPayload{
		ScheduleItemID: "item-1",
		SoundID:        "sound-1",
		Name:           "First Bell",
	}
	notifier.NotifyBellTriggered(payload)
	time.Sleep(50 * time.Millisecond)

	for _, c := range []*ws.Client{admin, client} {
		data := c.ReadSend()
		require.NotNil(t, data)
		var msg models.WebSocketMessage
		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Equal(t, "bell_triggered", msg.Type)
	}
}

func TestNotifier_BellCancelled(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	notifier.NotifyBellCancelled("item-1")
	time.Sleep(50 * time.Millisecond)

	for _, c := range []*ws.Client{admin, client} {
		data := c.ReadSend()
		require.NotNil(t, data)
		var msg models.WebSocketMessage
		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Equal(t, "bell_cancelled", msg.Type)
	}
}

func TestNotifier_SystemStateChanged(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	notifier.NotifySystemStateChanged("paused")
	time.Sleep(50 * time.Millisecond)

	for _, c := range []*ws.Client{admin, client} {
		data := c.ReadSend()
		require.NotNil(t, data)
		var msg models.WebSocketMessage
		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Equal(t, "system_state_changed", msg.Type)
	}
}

func TestNotifier_SystemLog_AdminOnly(t *testing.T) {
	notifier, admin, client := setupNotifier(t)
	notifier.NotifySystemLog("info", "test message", "scheduler")
	time.Sleep(50 * time.Millisecond)

	data := admin.ReadSend()
	require.NotNil(t, data, "admin should receive system_log")
	var msg models.WebSocketMessage
	require.NoError(t, json.Unmarshal(data, &msg))
	assert.Equal(t, "system_log", msg.Type)

	clientData := client.ReadSend()
	assert.Nil(t, clientData, "non-admin should NOT receive system_log")
}
