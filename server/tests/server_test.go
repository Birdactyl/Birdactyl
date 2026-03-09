package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers/server"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func mockServerApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_server_user_tests",
		Email:    "test_server_user_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestServerRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockServerApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/servers", server.AdminGetServers)
	app.Post("/servers", server.AdminCreateServer)

	t.Run("List Servers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/servers", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to list servers: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	var testServer models.Server
	err := database.DB.First(&testServer).Error
	if err == nil {
		app.Get("/servers/:id", server.AdminViewServer)
		app.Get("/servers/:id/activity", server.GetServerActivity)
		app.Get("/servers/:id/databases", server.GetServerDatabases)
		app.Get("/servers/:id/allocations", server.GetServerAllocations)

		t.Run("View Server", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s", testServer.ID.String()), nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to view server: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNotFound {
				t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
			}
		})

		t.Run("Get Server Activity", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/activity", testServer.ID.String()), nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to get server activity: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNotFound {
				t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
			}
		})

		t.Run("Get Server Databases", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/databases", testServer.ID.String()), nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to get server databases: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNotFound {
				t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
			}
		})

		t.Run("Get Server Allocations", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/allocations", testServer.ID.String()), nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to get server allocations: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK && resp.StatusCode != fiber.StatusNotFound {
				t.Errorf("Expected status 200 or 404, got %d", resp.StatusCode)
			}
		})
	}
}
