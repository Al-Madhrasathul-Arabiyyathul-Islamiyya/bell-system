package models

import (
	"time"

	"github.com/google/uuid"
)

// FileType represents the types of system audio files
type FileType string

const (
	FileTypeBell       FileType = "bell"
	FileTypeAnthem     FileType = "anthem"
	FileTypeSchoolSong FileType = "school_song"
	FileTypeOther      FileType = "other"
)

// SystemAudioFile represents a system audio file
type SystemAudioFile struct {
	ID        uuid.UUID `json:"id" db:"Id"`
	Name      string    `json:"name" db:"Name"`
	FilePath  string    `json:"filePath" db:"FilePath"`
	FileType  FileType  `json:"fileType" db:"FileType"`
	Checksum  string    `json:"checksum" db:"Checksum"`
	CreatedAt time.Time `json:"createdAt" db:"CreatedAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" db:"UpdatedAt"`
}
