package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers/auth"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestPasswordResetRoutes(t *testing.T) {
	requireDB(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/auth/reset/request", auth.RequestPasswordReset)
	app.Post("/auth/reset", auth.ResetPassword)

	uid := uuid.New().String()[:8]
	email := fmt.Sprintf("test_reset_%s@test.com", uid)

	user := models.User{
		ID:       uuid.New(),
		Username: fmt.Sprintf("test_reset_user_%s", uid),
		Email:    email,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	defer database.DB.Where("id = ?", user.ID).Delete(&models.User{})

	t.Run("Request Password Reset", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/reset/request", toJSONBody(map[string]string{
			"email": email,
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test request reset: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Reset Password - Missing Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/reset", toJSONBody(map[string]string{
			"token":    "",
			"password": "NewSecurePassword123!",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test reset password: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 400 for missing token, got %d", resp.StatusCode)
		}
	})

	t.Run("Reset Password - Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/reset", toJSONBody(map[string]string{
			"token":    "invalid.token.here",
			"password": "NewSecurePassword123!",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test reset password: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest && resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("Expected status 400 or 401 for invalid token, got %d", resp.StatusCode)
		}
	})
	
	t.Run("Reset Password - Short Password", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/reset", toJSONBody(map[string]string{
			"token":    "some_token",
			"password": "short",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test reset password: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 400 for short password, got %d", resp.StatusCode)
		}
	})
}
