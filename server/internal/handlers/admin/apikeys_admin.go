package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const apiKeyPrefix = "birdactyl_"

func generateAPIKey() (string, string) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	key := apiKeyPrefix + hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(key))
	return key, hex.EncodeToString(hash[:])
}

func canManageUserAPIKeys(currentUser *models.User, targetUser *models.User) bool {
	currentIsRoot := config.IsRootAdmin(currentUser.ID.String())
	targetIsRoot := config.IsRootAdmin(targetUser.ID.String())

	if targetIsRoot {
		return false
	}

	if targetUser.IsAdmin && !currentIsRoot {
		return false
	}

	return true
}

func AdminGetUserAPIKeys(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	var targetUser models.User
	if err := database.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	if !canManageUserAPIKeys(currentUser, &targetUser) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot manage API keys for this user"})
	}

	var keys []models.APIKey
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys)

	return c.JSON(fiber.Map{"success": true, "data": keys})
}

type AdminCreateAPIKeyRequest struct {
	Name      string `json:"name"`
	ExpiresIn *int   `json:"expires_in"`
}

func AdminCreateUserAPIKey(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	var targetUser models.User
	if err := database.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	if !canManageUserAPIKeys(currentUser, &targetUser) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot manage API keys for this user"})
	}

	var req AdminCreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.Name == "" {
		req.Name = "API Key"
	}

	plainKey, keyHash := generateAPIKey()

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &exp
	}

	apiKey := models.APIKey{
		UserID:    userID,
		Name:      req.Name,
		KeyHash:   keyHash,
		KeyPrefix: plainKey[:len(apiKeyPrefix)+8],
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to create API key"})
	}

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserUpdate, "Created API key for user: "+targetUser.Username, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"target_user_id": userID, "key_name": req.Name})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":         apiKey.ID,
			"name":       apiKey.Name,
			"key":        plainKey,
			"key_prefix": apiKey.KeyPrefix,
			"expires_at": apiKey.ExpiresAt,
			"created_at": apiKey.CreatedAt,
		},
	})
}

func AdminDeleteUserAPIKey(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	keyID, err := uuid.Parse(c.Params("keyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid key ID"})
	}

	var targetUser models.User
	if err := database.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	if !canManageUserAPIKeys(currentUser, &targetUser) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot manage API keys for this user"})
	}

	result := database.DB.Where("id = ? AND user_id = ?", keyID, userID).Delete(&models.APIKey{})
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "API key not found"})
	}

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserUpdate, "Deleted API key for user: "+targetUser.Username, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"target_user_id": userID, "key_id": keyID})

	return c.JSON(fiber.Map{"success": true, "message": "API key deleted"})
}
