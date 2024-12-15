package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

// Initialize a global database connection
var DB *gorm.DB

// ConnectDB sets up the database connection
func ConnectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("school_bell_system.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// AutoMigrate the database models
	err = db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Schedule{},
		&models.Weekday{},
		&models.BellSoundFile{},
		&models.AudioFile{},
	)
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}
