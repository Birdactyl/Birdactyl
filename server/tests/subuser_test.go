package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func mockSubuserApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_subuser_admin_tests",
		Email:    "test_subuser_admin_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestSubuserRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockSubuserApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/servers/:id/subusers", handlers.GetSubusers)
	app.Post("/servers/:id/subusers", handlers.AddSubuser)

	t.Run("Get Subusers - Invalid Server", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/servers/invalid-uuid/subusers", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test subusers route: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Subusers - Non-existent Server", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/subusers", uuid.New().String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test subusers route: %v", err)
		}
		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}
