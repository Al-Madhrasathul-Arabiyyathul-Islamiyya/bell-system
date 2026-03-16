package models

import "time"

// SystemState represents a key-value entry in the SystemState table.
type SystemState struct {
	Key       string    `json:"key" db:"Key"`
	Value     string    `json:"value" db:"Value"`
	UpdatedAt time.Time `json:"updatedAt" db:"UpdatedAt"`
}
