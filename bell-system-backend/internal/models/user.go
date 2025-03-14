package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin         Role = "admin"
	RoleMorningUser   Role = "morning_user"
	RoleAfternoonUser Role = "afternoon_user"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"Id"`
	Username     string    `json:"username" db:"Username"`
	PasswordHash string    `json:"-" db:"PasswordHash"`
	Role         Role      `json:"role" db:"Role"`
	CreatedAt    time.Time `json:"createdAt" db:"CreatedAt"`
}
