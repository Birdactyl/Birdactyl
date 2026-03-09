package tests

import (
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"
)

func TestSettings(t *testing.T) {
	requireDB(t)

	// Settings to clean up so that we dont mess up anything idk
	keys := []string{
		"test_setting",
		"registration_enabled",
		"server_creation_enabled",
		"email_verification_enabled",
	}

	defer func() {
		for _, k := range keys {
			database.DB.Where("key = ?", k).Delete(&models.Setting{})
		}
	}()

	t.Run("Set and Get Setting", func(t *testing.T) {
		err := services.SetSetting("test_setting", "test_value")
		if err != nil {
			t.Fatalf("Failed to set setting: %v", err)
		}

		val := services.GetSetting("test_setting")
		if val != "test_value" {
			t.Errorf("Expected test_value, got '%s'", val)
		}
	})

	t.Run("Registration Status", func(t *testing.T) {
		err := services.SetRegistrationEnabled(true)
		if err != nil {
			t.Fatalf("Failed to set registration to true: %v", err)
		}
		if !services.IsRegistrationEnabled() {
			t.Errorf("Expected registration to be enabled")
		}

		err = services.SetRegistrationEnabled(false)
		if err != nil {
			t.Fatalf("Failed to set registration to false: %v", err)
		}
		if services.IsRegistrationEnabled() {
			t.Errorf("Expected registration to be disabled")
		}
	})

	t.Run("Server Creation Status", func(t *testing.T) {
		err := services.SetServerCreationEnabled(true)
		if err != nil {
			t.Fatalf("Failed to set server creation to true: %v", err)
		}
		if !services.IsServerCreationEnabled() {
			t.Errorf("Expected server creation to be enabled")
		}

		err = services.SetServerCreationEnabled(false)
		if err != nil {
			t.Fatalf("Failed to set server creation to false: %v", err)
		}
		if services.IsServerCreationEnabled() {
			t.Errorf("Expected server creation to be disabled")
		}
	})

	t.Run("Email Verification Status", func(t *testing.T) {
		err := services.SetEmailVerificationEnabled(true)
		if err != nil {
			t.Fatalf("Failed to set email verification to true: %v", err)
		}
		if !services.IsEmailVerificationEnabled() {
			t.Errorf("Expected email verification to be enabled")
		}

		err = services.SetEmailVerificationEnabled(false)
		if err != nil {
			t.Fatalf("Failed to set email verification to false: %v", err)
		}
		if services.IsEmailVerificationEnabled() {
			t.Errorf("Expected email verification to be disabled")
		}
	})
}
