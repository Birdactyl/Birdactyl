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

func TestAuthRegisterAndLogin(t *testing.T) {
	requireDB(t)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/register", auth.Register)
	app.Post("/login", auth.Login)

	uid := uuid.New().String()[:8]
	email := fmt.Sprintf("test_auth_%s@test.com", uid)
	username := fmt.Sprintf("test_auth_%s", uid)
	password := "SecurePassword123!"

	defer func() {
		database.DB.Where("email = ?", email).Delete(&models.User{})
		database.DB.Where("username = ?", username).Delete(&models.User{})
	}()

	t.Run("Register", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register", toJSONBody(map[string]string{
			"email":    email,
			"username": username,
			"password": password,
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test registration: %v", err)
		}

		if resp.StatusCode != fiber.StatusCreated && resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusForbidden {
			body := parseJSONResponse(resp)
			t.Errorf("Expected registration to succeed or yield 403 (verification enabled), got %d: %v", resp.StatusCode, body)
		}
	})

	database.DB.Model(&models.User{}).Where("email = ?", email).Update("email_verified", true)

	t.Run("Login", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", toJSONBody(map[string]string{
			"email":    email,
			"password": password,
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test login: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			body := parseJSONResponse(resp)
			t.Errorf("Expected login to succeed, got %d. error: %v", resp.StatusCode, body["error"])
		} else {
			body := parseJSONResponse(resp)
			if body["success"] != true {
				t.Errorf("Expected success=true, got %v", body)
			}
		}
	})

	t.Run("Login Invalid Password", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", toJSONBody(map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test login: %v", err)
		}

		if resp.StatusCode != fiber.StatusUnauthorized && resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("Expected login to fail with unauthorized, got %d", resp.StatusCode)
		}
	})
}
