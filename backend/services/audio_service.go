package services

import (
	"backend/db"
	"backend/models"
)

// FindAllAudios fetches all non-bell audio files
func FindAllAudios() ([]models.AudioFile, error) {
	var audios []models.AudioFile
	if err := db.DB.Where("is_bell = ?", false).Find(&audios).Error; err != nil {
		return nil, err
	}
	return audios, nil
}

// FindAllBellAudios fetches all bell sound files
func FindAllBellAudios() ([]models.AudioFile, error) {
	var audios []models.AudioFile
	if err := db.DB.Where("is_bell = ?", true).Find(&audios).Error; err != nil {
		return nil, err
	}
	return audios, nil
}

// FindAudioByID fetches an audio file by ID
func FindAudioByID(id uint) (*models.AudioFile, error) {
	var audio models.AudioFile
	if err := db.DB.First(&audio, id).Error; err != nil {
		return nil, err
	}
	return &audio, nil
}

// CreateAudio adds a new audio record
func CreateAudio(audio *models.AudioFile) error {
	return db.DB.Create(audio).Error
}

// DeleteAudio removes an audio record by ID
func DeleteAudio(id uint) error {
	var audio models.AudioFile
	if err := db.DB.First(&audio, id).Error; err != nil {
		return err
	}
	return db.DB.Delete(&audio).Error
}
