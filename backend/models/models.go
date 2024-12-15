package models

import (
	"time"

	"gorm.io/gorm"
)

func CheckAdminExists(db *gorm.DB) (bool, error) {
	var count int64
	if err := db.Model(&User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// User model
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;type:enum('admin', 'morning', 'afternoon')"`
	CreatedAt    time.Time
}

// UserDTO is used for returning user information without sensitive data
type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Session model
type Session struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	StartTime string `gorm:"not null"`
	EndTime   string `gorm:"not null"`
	CreatedAt time.Time
}

// Schedule model
type Schedule struct {
	ID          uint   `gorm:"primaryKey"`
	SessionID   uint   `gorm:"not null"`
	Description string `gorm:"type:text"`
	StartTime   string `gorm:"not null"`
	EndTime     string `gorm:"not null"`
	SoundAlert  string `gorm:"not null"`
	CreatedAt   time.Time
	Weekdays    []Weekday `gorm:"many2many:schedule_weekdays"`
}

// Weekday model
type Weekday struct {
	ID        uint       `gorm:"primaryKey"`
	Name      string     `gorm:"unique;not null"` // e.g., Monday, Tuesday, etc.
	Schedules []Schedule `gorm:"many2many:schedule_weekdays"`
}

// BellSoundFile model
type BellSoundFile struct {
	ID        uint   `gorm:"primaryKey"`
	Filename  string `gorm:"not null"`
	Filepath  string `gorm:"not null"`
	CreatedAt time.Time
}

// AudioFile model
type AudioFile struct {
	ID        uint   `gorm:"primaryKey"`
	Filename  string `gorm:"not null"`
	Filepath  string `gorm:"not null"`
	CreatedAt time.Time
}
