package server

import (
	"crypto/rand"
	"encoding/base64"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func GetSFTPDetails(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermSFTPView)
	if err != nil {
		return nil
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", server.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to fetch user"})
	}

	username := user.Username + "." + server.ID.String()

	host := ""
	if server.Node != nil {
		host = server.Node.FQDN
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"host":     host,
			"port":     2022,
			"username": username,
		},
	})
}

func ResetSFTPPassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermSFTPResetPassword)
	if err != nil {
		return nil
	}

	password, err := generatePassword(24)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to generate password"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to hash password"})
	}

	if err := database.DB.Model(&models.Server{}).Where("id = ?", serverID).Update("sftp_password", string(hashedPassword)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to save password"})
	}

	handlers.Log(c, user, handlers.ActionSFTPPasswordReset, "Reset SFTP password", map[string]interface{}{"server_id": serverID, "server_name": server.Name})

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"password": password,
		},
	})
}
