// internal/models/websocket.go
package models

import (
	"time"
)

// WebSocketMessage represents a message exchanged over WebSocket
type WebSocketMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload,omitempty"`
}

// ClientInfo represents information about a connected client
type ClientInfo struct {
	ID             string    `json:"id"`
	IP             string    `json:"ip"`
	ClientType     string    `json:"client_type"`
	ClientName     string    `json:"client_name,omitempty"`
	ConnectedSince time.Time `json:"connected_since"`
}

// RegistrationPayload represents the payload for client registration
type RegistrationPayload struct {
	ClientType string `json:"client_type"`
	ClientName string `json:"client_name,omitempty"`
}

// BellTriggeredPayload represents the payload for a bell triggered event
type BellTriggeredPayload struct {
	ScheduleItemID string `json:"scheduleItemId"`
	SoundID        string `json:"soundId"`
	Name           string `json:"name"`
}

// SystemStatePayload represents the payload for a system state change event
type SystemStatePayload struct {
	State string `json:"state"`
}

// SystemLogPayload represents the payload for a system log event
type SystemLogPayload struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Source  string `json:"source"`
}
