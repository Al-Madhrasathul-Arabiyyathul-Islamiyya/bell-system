package models

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleItem represents an item in the bell schedule
type ScheduleItem struct {
	ID        uuid.UUID        `json:"id" db:"Id"`
	SessionID *uuid.UUID       `json:"sessionId,omitempty" db:"SessionId"`
	Name      string           `json:"name" db:"Name"`
	Time      time.Time        `json:"time" db:"Time"`
	SoundID   uuid.UUID        `json:"soundId" db:"SoundId"`
	CreatedAt time.Time        `json:"createdAt" db:"CreatedAt"`
	UpdatedAt time.Time        `json:"updatedAt,omitempty" db:"UpdatedAt"`
	Days      []int            `json:"days" db:"-"`              // Days of week, populated separately
	Sound     *SystemAudioFile `json:"sound,omitempty" db:"-"`   // Related sound file
	Session   *Session         `json:"session,omitempty" db:"-"` // Related session
}
