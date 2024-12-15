package services

import (
	"backend/db"
	"backend/models"
)

// FindAllBells fetches all bell records
func FindAllBells() ([]models.BellSoundFile, error) {
	var bells []models.BellSoundFile
	if err := db.DB.Find(&bells).Error; err != nil {
		return nil, err
	}
	return bells, nil
}

// FindBellByID fetches a single bell record by ID
func FindBellByID(id uint) (*models.BellSoundFile, error) {
	var bell models.BellSoundFile
	if err := db.DB.First(&bell, id).Error; err != nil {
		return nil, err
	}
	return &bell, nil
}

// CreateBell adds a new bell record
func CreateBell(bell *models.BellSoundFile) error {
	return db.DB.Create(bell).Error
}

// UpdateBell updates an existing bell record
func UpdateBell(id uint, updatedData *models.BellSoundFile) error {
	var bell models.BellSoundFile
	if err := db.DB.First(&bell, id).Error; err != nil {
		return err
	}
	return db.DB.Model(&bell).Updates(updatedData).Error
}

// DeleteBell removes a bell record
func DeleteBell(id uint) error {
	var bell models.BellSoundFile
	if err := db.DB.First(&bell, id).Error; err != nil {
		return err
	}
	return db.DB.Delete(&bell).Error
}
