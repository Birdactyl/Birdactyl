package auth

import (
	"fmt"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func SendVerification(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	c.BodyParser(&req)

	var requestEmail string
	if claims, ok := c.Locals("claims").(*services.AccessClaims); ok && claims != nil {
		if dbUser, err := services.GetUserByID(claims.UserID); err == nil {
			requestEmail = dbUser.Email
		}
	} else if req.Email != "" {
		requestEmail = req.Email
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Email is required"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", requestEmail).First(&user).Error; err != nil {
		return c.JSON(fiber.Map{"success": true, "message": "Verification email sent"})
	}

	cfg := config.Get()
	if !cfg.SMTPEnabled() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Email is not configured"})
	}

	baseURL := cfg.Server.BaseURL
	if baseURL == "" {
		proto := "http"
		if c.Get("X-Forwarded-Proto") == "https" || c.Secure() {
			proto = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", proto, c.Hostname())
	}

	if err := services.SendVerificationEmail(user.ID, baseURL); err != nil {
		if err == services.ErrAlreadyVerified {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Email is already verified"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to send verification email"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Verification email sent"})
}

func VerifyEmail(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Token is required"})
	}

	if err := services.VerifyEmail(req.Token); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Email verified"})
}
