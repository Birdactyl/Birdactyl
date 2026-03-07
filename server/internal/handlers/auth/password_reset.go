package auth

import (
	"fmt"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func RequestPasswordReset(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Email is required"})
	}

	cfg := config.Get()
	if !cfg.SMTPEnabled() {
		return c.JSON(fiber.Map{"success": true, "message": "If that email exists, a password reset link has been sent"})
	}

	baseURL := cfg.Server.BaseURL
	if baseURL == "" {
		proto := "http"
		if c.Get("X-Forwarded-Proto") == "https" || c.Secure() {
			proto = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", proto, c.Hostname())
	}

	services.RequestPasswordReset(req.Email, baseURL)

	return c.JSON(fiber.Map{"success": true, "message": "If that email exists, a password reset link has been sent"})
}

func ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil || req.Token == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Token and password are required"})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Password must be at least 8 characters"})
	}

	if err := services.ResetPassword(req.Token, req.Password); err != nil {
		status := fiber.StatusBadRequest
		if err == services.ErrResetTokenInvalid || err == services.ErrResetTokenUsed {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Password has been reset"})
}
