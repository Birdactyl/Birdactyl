package tests

import (
	"testing"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/google/uuid"
)

func TestSessions(t *testing.T) {
	requireDB(t)

	user := models.User{
		ID:       uuid.New(),
		Username: "test_sessions_user",
		Email:    "test_sessions_user@test.com",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	defer database.DB.Where("id = ?", user.ID).Delete(&models.User{})
	defer database.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})

	session1 := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: "refresh1",
		UserAgent:    "TestAgent1",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	session2 := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: "refresh2",
		UserAgent:    "TestAgent2",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // expiration jumpscare
	}
	database.DB.Create(&session1)
	database.DB.Create(&session2)

	t.Run("GetUserSessions", func(t *testing.T) {
		sessions := services.GetUserSessions(user.ID, session1.ID)
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(sessions))
		}
		
		var foundCurrent bool
		for _, s := range sessions {
			if s.IsCurrent && s.ID == session1.ID {
				foundCurrent = true
			}
		}
		if !foundCurrent {
			t.Errorf("Expected session1 to be marked as current")
		}
	})

	t.Run("CleanExpiredSessions", func(t *testing.T) {
		services.CleanExpiredSessions()
		
		services.Cache.Delete("sessions_" + user.ID.String())
		sessions := services.GetUserSessions(user.ID, session1.ID)
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session after cleanup, got %d", len(sessions))
		}
	})

	t.Run("RevokeSession", func(t *testing.T) {
		err := services.RevokeSession(user.ID, session1.ID.String())
		if err != nil {
			t.Errorf("Failed to revoke session: %v", err)
		}

		services.Cache.Delete("sessions_" + user.ID.String())
		sessions := services.GetUserSessions(user.ID, session1.ID)
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after revoke, got %d", len(sessions))
		}
	})
	
	session3 := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: "refresh3",
		UserAgent:    "TestAgent3",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	session4 := models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: "refresh4",
		UserAgent:    "TestAgent4",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	database.DB.Create(&session3)
	database.DB.Create(&session4)

	t.Run("RevokeOtherSessions", func(t *testing.T) {
		services.RevokeOtherSessions(user.ID, session3.ID)
		
		services.Cache.Delete("sessions_" + user.ID.String())
		sessions := services.GetUserSessions(user.ID, session3.ID)
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session after revoking others, got %d", len(sessions))
		}
		if sessions[0].ID != session3.ID {
			t.Errorf("Expected remaining session to be the current one")
		}
	})

	t.Run("LogoutAll", func(t *testing.T) {
		err := services.LogoutAll(user.ID)
		if err != nil {
			t.Errorf("Failed to logout all: %v", err)
		}

		services.Cache.Delete("sessions_" + user.ID.String())
		sessions := services.GetUserSessions(user.ID, uuid.Nil)
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after logout all, got %d", len(sessions))
		}
	})
}
