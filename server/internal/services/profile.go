package services

import (
	"errors"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func AdminCreateUser(email, username, password string) (*models.User, error) {
	cfg := config.Get()

	var existing models.User
	if err := database.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	}
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Auth.BcryptCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		RegisterIP:   "admin-created",
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateProfile(userID uuid.UUID, username, email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	if username != "" && username != user.Username {
		var existing models.User
		if err := database.DB.Where("username = ? AND id != ?", username, userID).First(&existing).Error; err == nil {
			return nil, ErrUsernameTaken
		}
		user.Username = username
	}

	if email != "" && email != user.Email {
		var existing models.User
		if err := database.DB.Where("email = ? AND id != ?", email, userID).First(&existing).Error; err == nil {
			return nil, ErrEmailTaken
		}
		user.Email = email
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdatePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	cfg := config.Get()

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), cfg.Auth.BcryptCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ForcePasswordReset = false
	return database.DB.Save(&user).Error
}
