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

func mockScheduleApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_schedule_admin_tests",
		Email:    "test_schedule_admin_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestScheduleRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockScheduleApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/servers/:id/schedules", handlers.GetServerSchedules)
	app.Post("/servers/:id/schedules", handlers.CreateSchedule)

	t.Run("Get Schedules - Invalid Server", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/servers/invalid-uuid/schedules", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test schedules route: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Schedules - Non-existent Server", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/schedules", uuid.New().String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test schedules route: %v", err)
		}
		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}
