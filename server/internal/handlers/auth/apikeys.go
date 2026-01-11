package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"birdactyl-panel-backend/internal/database"
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

func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func GetAPIKeys(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var keys []models.APIKey
	database.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&keys)

	return c.JSON(fiber.Map{"success": true, "data": keys})
}

type CreateAPIKeyRequest struct {
	Name      string `json:"name"`
	ExpiresIn *int   `json:"expires_in"`
}

func CreateAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var req CreateAPIKeyRequest
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
		UserID:    user.ID,
		Name:      req.Name,
		KeyHash:   keyHash,
		KeyPrefix: plainKey[:len(apiKeyPrefix)+8],
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(&apiKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to create API key"})
	}

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

func DeleteAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	keyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid key ID"})
	}

	result := database.DB.Where("id = ? AND user_id = ?", keyID, user.ID).Delete(&models.APIKey{})
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "API key not found"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "API key deleted"})
}

func ValidateAPIKey(key string) (*models.User, error) {
	keyHash := HashAPIKey(key)

	var apiKey models.APIKey
	if err := database.DB.Preload("User").Where("key_hash = ?", keyHash).First(&apiKey).Error; err != nil {
		return nil, err
	}

	if apiKey.IsExpired() {
		database.DB.Delete(&apiKey)
		return nil, fiber.NewError(fiber.StatusUnauthorized, "API key expired")
	}

	now := time.Now()
	database.DB.Model(&apiKey).Update("last_used_at", now)

	return apiKey.User, nil
}
