package db

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"backend/models"
)

// CreateAdminUser checks and creates an admin user if none exists
func CreateAdminUser() {
	adminExists, err := models.CheckAdminExists(DB)
	if err != nil {
		log.Fatalf("Error checking admin existence: %v", err)
	}

	if adminExists {
		log.Println("Admin user already exists. Skipping admin creation.")
		return
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminUsername == "" || adminPassword == "" {
		log.Fatalf("ADMIN_USERNAME or ADMIN_PASSWORD not set in .env file")
	}

	// Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	admin := models.User{
		Username:     adminUsername,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Println("Admin user created successfully.")
}
