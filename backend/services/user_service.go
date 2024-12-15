package services

import (
	"backend/db"
	"backend/models"

	"gorm.io/gorm"
)

// FindAllUsers fetches all users
func FindAllUsers() ([]models.UserDTO, error) {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	// Transform users to UserDTO
	userDTOs := make([]models.UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = models.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		}
	}
	return userDTOs, nil
}

// FindUserByUsername fetches a user by their username
func FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPassword updates the password for a user by ID
func UpdateUserPassword(id uint, hashedPassword string) error {
	// Locate the user by ID and update only the password field
	result := db.DB.Model(&models.User{}).Where("id = ?", id).Update("password_hash", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindUserByID retrieves a user by their ID
func FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func CreateUser(user *models.User) error {
	return db.DB.Create(user).Error
}

// DeleteUser deletes a user by ID
func DeleteUser(userID uint) error {
	return db.DB.Delete(&models.User{}, userID).Error
}

// UpdateUser updates user data
func UpdateUser(user *models.User) error {
	return db.DB.Save(user).Error
}

// CheckAdminExists checks if an admin exists in the database
func CheckAdminExists() (bool, error) {
	var count int64
	if err := db.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
