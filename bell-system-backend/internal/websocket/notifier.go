package websocket

import (
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// Notifier implements handlers.EventNotifier by broadcasting via the Hub.
type Notifier struct {
	hub *Hub
}

var _ handlers.EventNotifier = (*Notifier)(nil)

// NewNotifier creates a new Notifier backed by the given Hub.
func NewNotifier(hub *Hub) *Notifier {
	return &Notifier{hub: hub}
}

func (n *Notifier) NotifySchedulesUpdated() {
	n.hub.Broadcast(models.WebSocketMessage{
		Type:      "schedules_updated",
		Timestamp: time.Now().UTC(),
	})
}

func (n *Notifier) NotifyAudioFilesUpdated() {
	n.hub.Broadcast(models.WebSocketMessage{
		Type:      "audio_files_updated",
		Timestamp: time.Now().UTC(),
	})
}

func (n *Notifier) NotifyBellTriggered(payload models.BellTriggeredPayload) {
	n.hub.Broadcast(models.WebSocketMessage{
		Type:      "bell_triggered",
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	})
}

func (n *Notifier) NotifyBellCancelled(scheduleItemID string) {
	n.hub.Broadcast(models.WebSocketMessage{
		Type:      "bell_cancelled",
		Timestamp: time.Now().UTC(),
		Payload: models.BellCancelledPayload{
			ScheduleItemID: scheduleItemID,
		},
	})
}

func (n *Notifier) NotifySystemStateChanged(state string) {
	n.hub.Broadcast(models.WebSocketMessage{
		Type:      "system_state_changed",
		Timestamp: time.Now().UTC(),
		Payload: models.SystemStatePayload{
			State: state,
		},
	})
}

func (n *Notifier) NotifySystemLog(level, message, source string) {
	n.hub.BroadcastToAdmins(models.WebSocketMessage{
		Type:      "system_log",
		Timestamp: time.Now().UTC(),
		Payload: models.SystemLogPayload{
			Level:   level,
			Message: message,
			Source:  source,
		},
	})
}
