package middleware

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := c.Locals("claims").(*services.AccessClaims)

		var user models.User
		if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "User not found",
			})
		}

		if !user.IsAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Admin access required",
			})
		}

		c.Locals("user", &user)
		c.Locals("admin", true)
		return c.Next()
	}
}
