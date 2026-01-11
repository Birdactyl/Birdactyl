package middleware

import (
	"strings"

	"birdactyl-panel-backend/internal/handlers/auth"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimSpace(extractToken(c))
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization required",
			})
		}

		if strings.HasPrefix(token, "birdactyl_") {
			user, err := auth.ValidateAPIKey(token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Invalid or expired API key",
				})
			}
			c.Locals("user", user)
			c.Locals("via_api_key", true)
			return c.Next()
		}

		claims, needsRefresh, err := services.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		c.Locals("claims", claims)

		user, err := services.GetUserByID(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "User not found",
			})
		}
		c.Locals("user", user)

		if needsRefresh {
			newTokens, err := services.RefreshBySessionID(
				claims.SessionID,
				c.IP(),
				c.Get("User-Agent"),
			)
			if err == nil {
				c.Locals("new_tokens", newTokens)
			}
		}

		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if token := c.Query("token"); token != "" {
		return token
	}
	return ""
}

func WebSocketAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization required",
			})
		}

		if strings.HasPrefix(token, "birdactyl_") {
			user, err := auth.ValidateAPIKey(token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Invalid or expired API key",
				})
			}
			c.Locals("userID", user.ID)
			c.Locals("isAdmin", user.IsAdmin)
			return c.Next()
		}

		claims, _, err := services.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		c.Locals("userID", claims.UserID)

		user, err := services.GetUserByID(claims.UserID)
		if err == nil && user != nil {
			c.Locals("isAdmin", user.IsAdmin)
		}

		return c.Next()
	}
}