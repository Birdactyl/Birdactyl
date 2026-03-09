package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers/admin"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func mockAdminApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_admin_user_admin_tests",
		Email:    "test_admin_admin_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestAdminRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockAdminApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/admin/users", admin.AdminGetUsers)
	app.Post("/admin/users", admin.AdminCreateUser)
	app.Get("/admin/settings", admin.AdminGetSettings)
	app.Get("/admin/databases", admin.AdminGetDatabaseHosts)
	app.Get("/admin/plugins", admin.AdminListPlugins)
	app.Get("/admin/ipbans", admin.AdminGetIPBans)
	app.Get("/admin/logs", admin.AdminGetLogs)

	t.Run("Get Users", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/users", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Create User", func(t *testing.T) {
		uid := uuid.New().String()[:8]
		req := httptest.NewRequest("POST", "/admin/users", toJSONBody(map[string]string{
			"email":    fmt.Sprintf("test_new_%s@test.com", uid),
			"username": fmt.Sprintf("test_new_%s", uid),
			"password": "SecurePassword123!",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			body := parseJSONResponse(resp)
			t.Errorf("Expected status 201, got %d. error: %v", resp.StatusCode, body["error"])
		}
	})

	t.Run("Get Settings", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/settings", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get settings: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Database Hosts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/databases", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get database hosts: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Plugins", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/plugins", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get plugins: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Get IP Bans", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/ipbans", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get ip bans: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Activity Logs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/logs", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
