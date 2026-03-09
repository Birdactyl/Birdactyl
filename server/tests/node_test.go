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

func mockNodeApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_node_admin_tests",
		Email:    "test_node_admin_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestNodeRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockNodeApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/nodes", handlers.AdminGetNodes)
	app.Post("/nodes", handlers.AdminCreateNode)
	app.Get("/nodes/:id", handlers.AdminGetNode)

	t.Run("Get Nodes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nodes", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get nodes: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Node", func(t *testing.T) {
		uid := uuid.New().String()[:8]
		req := httptest.NewRequest("POST", "/nodes", toJSONBody(map[string]interface{}{
			"name": fmt.Sprintf("test_node_%s", uid),
			"fqdn": fmt.Sprintf("test-node-%s.local", uid),
			"port": 8443,
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to create node: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			body := parseJSONResponse(resp)
			t.Errorf("Expected status 201, got %d. error: %v", resp.StatusCode, body["error"])
		}
	})
}
