package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "birdactyl-backend",
		"version":   "1.0.0",
	})
}
