package handlers

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func ValidateSFTPAuth(c *fiber.Ctx) error {
	var req struct {
		ServerID string `json:"server_id"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	serverID, err := uuid.Parse(req.ServerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var server models.Server
	if err := database.DB.First(&server, "id = ?", serverID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Authentication failed"})
	}

	if server.SFTPPassword == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "SFTP password not set"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(server.SFTPPassword), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "Authentication failed"})
	}

	return c.JSON(fiber.Map{"success": true})
}
