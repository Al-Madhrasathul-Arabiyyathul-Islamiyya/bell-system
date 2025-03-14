package models

import (
	"github.com/google/uuid"
)

// ScheduleDay represents a day of the week a schedule item applies to
type ScheduleDay struct {
	ScheduleItemID uuid.UUID `json:"scheduleItemId" db:"ScheduleItemId"`
	DayOfWeek      int       `json:"dayOfWeek" db:"DayOfWeek"` // 1-7, where 1 is Monday
}
