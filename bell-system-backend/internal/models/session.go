package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a school session (morning or afternoon)
type Session struct {
	ID        uuid.UUID `json:"id" db:"Id"`
	Name      string    `json:"name" db:"Name"`
	StartTime time.Time `json:"startTime" db:"StartTime"`
	EndTime   time.Time `json:"endTime" db:"EndTime"`
}
