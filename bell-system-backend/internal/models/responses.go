package models

import (
	"github.com/google/uuid"
)

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
