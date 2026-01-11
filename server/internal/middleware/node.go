package middleware

import (
	"strings"

	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func RequireNodeAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization required",
			})
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token format",
			})
		}

		node, err := services.ValidateNodeToken(parts[0], parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		c.Locals("node", node)
		return c.Next()
	}
}
