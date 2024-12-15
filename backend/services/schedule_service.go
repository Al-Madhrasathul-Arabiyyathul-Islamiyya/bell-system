package services

import (
	"backend/db"
	"backend/models"
)

// FindAllSchedules retrieves all schedules from the database
func FindAllSchedules() ([]models.Schedule, error) {
	var schedules []models.Schedule
	if err := db.DB.Preload("Weekdays").Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

// FindScheduleByID fetches a single schedule by ID
func FindScheduleByID(id uint) (*models.Schedule, error) {
	var schedule models.Schedule
	if err := db.DB.Preload("Weekdays").First(&schedule, id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

// CreateSchedule adds a new schedule
func CreateSchedule(schedule *models.Schedule) error {
	return db.DB.Create(schedule).Error
}

// UpdateSchedule updates an existing schedule
func UpdateSchedule(id uint, updatedData *models.Schedule) error {
	var schedule models.Schedule
	if err := db.DB.First(&schedule, id).Error; err != nil {
		return err
	}
	return db.DB.Model(&schedule).Updates(updatedData).Error
}

// DeleteSchedule removes a schedule from the database
func DeleteSchedule(id uint) error {
	var schedule models.Schedule
	if err := db.DB.First(&schedule, id).Error; err != nil {
		return err
	}
	return db.DB.Delete(&schedule).Error
}
