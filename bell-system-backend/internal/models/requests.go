package models

import (
	"github.com/google/uuid"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserCreateRequest represents a request to create a new user
type UserCreateRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Role     Role   `json:"role" validate:"required,oneof=admin morning_user afternoon_user"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
	Role     Role   `json:"role,omitempty" validate:"omitempty,oneof=admin morning_user afternoon_user"`
}

// ScheduleItemCreateRequest represents a request to create a schedule item
type ScheduleItemCreateRequest struct {
	SessionID *uuid.UUID `json:"sessionId"`
	Name      string     `json:"name" validate:"required"`
	Time      string     `json:"time" validate:"required"`
	SoundID   uuid.UUID  `json:"soundId" validate:"required"`
	Days      []int      `json:"days" validate:"required,dive,min=1,max=7"`
}

// ScheduleItemUpdateRequest represents a request to update a schedule item
type ScheduleItemUpdateRequest struct {
	SessionID *uuid.UUID `json:"sessionId,omitempty"`
	Name      string     `json:"name,omitempty"`
	Time      string     `json:"time,omitempty"`
	SoundID   uuid.UUID  `json:"soundId,omitempty"`
	Days      []int      `json:"days,omitempty" validate:"omitempty,dive,min=1,max=7"`
}

// SystemStateRequest represents a request to change the system state
type SystemStateRequest struct {
	State string `json:"state" validate:"required,oneof=active paused"`
}
