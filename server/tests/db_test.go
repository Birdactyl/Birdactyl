package tests

import (
	"fmt"
	"testing"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
)

func TestDBUserCRUD(t *testing.T) {
	requireDB(t)

	uid := uuid.New().String()[:8]
	email := fmt.Sprintf("test_crud_%s@test.com", uid)
	username := fmt.Sprintf("test_crud_%s", uid)

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}

	t.Run("Create User", func(t *testing.T) {
		err := database.DB.Create(user).Error
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if user.ID == uuid.Nil {
			t.Errorf("Expected user ID to be populated")
		}
	})

	t.Run("Read User", func(t *testing.T) {
		var found models.User
		err := database.DB.Where("email = ?", email).First(&found).Error
		if err != nil {
			t.Fatalf("Failed to read user: %v", err)
		}
		if found.Username != username {
			t.Errorf("Expected username %s, got %s", username, found.Username)
		}
	})

	t.Run("Update User", func(t *testing.T) {
		newUsername := username + "_updated"
		err := database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("username", newUsername).Error
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		var found models.User
		database.DB.Where("id = ?", user.ID).First(&found)
		if found.Username != newUsername {
			t.Errorf("Expected updated username %s, got %s", newUsername, found.Username)
		}
	})

	t.Run("Delete User", func(t *testing.T) {
		err := database.DB.Where("id = ?", user.ID).Delete(&models.User{}).Error
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		var count int64
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Count(&count)
		if count > 0 {
			t.Errorf("Expected user to be deleted, but still found")
		}
	})
}

func TestDBSessionCRUD(t *testing.T) {
	requireDB(t)
    
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("test_sess_%s@test.com", uid),
		Username:     fmt.Sprintf("test_sess_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	defer database.DB.Where("id = ?", user.ID).Delete(&models.User{})

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: fmt.Sprintf("rt_test_%s", uuid.New().String()),
		UserAgent:    "TestAgent/1.0",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	t.Run("Create Session", func(t *testing.T) {
		err := database.DB.Create(session).Error
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	})

	t.Run("Delete Session", func(t *testing.T) {
		err := database.DB.Where("id = ?", session.ID).Delete(&models.Session{}).Error
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}
	})
}
