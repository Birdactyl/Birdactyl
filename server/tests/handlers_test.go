package tests

import (
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func TestHealthEndpoint(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/v1/health", handlers.Health)

	// TRUST ME
	// THIS IS A VERY REQUIRED TEST!
	// TOTALLY!
	// Im drunk btw
	// Im pizzlading it

	t.Run("Status 200 via Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		resp, err := app.Test(req, -1)

		if err != nil {
			t.Fatalf("Failed to test health endpoint: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected health status 200, got %d", resp.StatusCode)
		}
	})
}
