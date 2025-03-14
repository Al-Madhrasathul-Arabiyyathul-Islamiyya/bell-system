package models

import (
	"time"

	"github.com/google/uuid"
)

// LoginResponse represents the response to a successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// CurrentScheduleResponse represents the current schedule for today
type CurrentScheduleResponse struct {
	Session *Session              `json:"session"`
	Items   []CurrentScheduleItem `json:"items"`
}

// CurrentScheduleItem represents a schedule item with status for today
type CurrentScheduleItem struct {
	ID     uuid.UUID        `json:"id"`
	Name   string           `json:"name"`
	Time   string           `json:"time"`
	Sound  *SystemAudioFile `json:"sound"`
	Status string           `json:"status"` // pending, current, completed
}

// SystemStateResponse represents the current system state
type SystemStateResponse struct {
	State       string    `json:"state"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// AudioChecksumsResponse represents a list of audio file checksums
type AudioChecksumsResponse struct {
	Files []struct {
		ID       uuid.UUID `json:"id"`
		Type     string    `json:"type"`
		Checksum string    `json:"checksum"`
	} `json:"files"`
}

// ListResponse is a generic structure for list responses
type ListResponse struct {
	Total int `json:"total"`
	Items any `json:"items"`
}
