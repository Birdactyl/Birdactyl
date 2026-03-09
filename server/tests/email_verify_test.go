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

func TestEmailVerificationRoutes(t *testing.T) {
	requireDB(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/auth/verify/send", auth.SendVerification)
	app.Post("/auth/verify", auth.VerifyEmail)

	uid := uuid.New().String()[:8]
	email := fmt.Sprintf("test_verify_%s@test.com", uid)

	user := models.User{
		ID:       uuid.New(),
		Username: fmt.Sprintf("test_verify_user_%s", uid),
		Email:    email,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	defer database.DB.Where("id = ?", user.ID).Delete(&models.User{})

	t.Run("Send Verification", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/verify/send", toJSONBody(map[string]string{
			"email": email,
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test send verification: %v", err)
		}
		// o_o
		if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 200 or 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Verify Email - Missing Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/verify", toJSONBody(map[string]string{
			"token": "",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test verify email: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 400 for missing token, got %d", resp.StatusCode)
		}
	})

	t.Run("Verify Email - Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/verify", toJSONBody(map[string]string{
			"token": "invalid.token.here",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test verify email: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("Expected status 401 for invalid token, got %d", resp.StatusCode)
		}
	})
}
