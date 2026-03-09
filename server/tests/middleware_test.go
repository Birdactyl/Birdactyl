package tests

import (
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func TestRateLimiter(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 2,
		BurstLimit:        2,
	}), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	t.Run("Under Limit (Allowed)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1") // arbitrary loopback jumpscare

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test rate limiter: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Over Limit (Rejected)", func(t *testing.T) {
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-Forwarded-For", "192.168.1.1")
		app.Test(req2, -1)

		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.Header.Set("X-Forwarded-For", "192.168.1.1")
		resp, _ := app.Test(req3, -1)

		if resp.StatusCode != fiber.StatusTooManyRequests {
			t.Logf("Expected status 429 after bursting, got %d. Ensure rate-limiter strict behavior.", resp.StatusCode)
		}
	})
}
