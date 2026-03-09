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

func mockPackageApp() (*fiber.App, *models.User) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_package_admin_tests",
		Email:    "test_package_admin_tests@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	return app, adminUser
}

func TestPackageRoutes(t *testing.T) {
	requireDB(t)

	app, adminUser := mockPackageApp()
	defer database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})

	app.Get("/packages", handlers.AdminGetPackages)
	app.Post("/packages", handlers.AdminCreatePackage)

	t.Run("Get Packages", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/packages", nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to get packages: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Package", func(t *testing.T) {
		uid := uuid.New().String()[:8]
		req := httptest.NewRequest("POST", "/packages", toJSONBody(map[string]interface{}{
			"name":         fmt.Sprintf("test_pkg_%s", uid),
			"docker_image": "debian:latest",
			"startup":      "./start.sh",
		}))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to create package: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			body := parseJSONResponse(resp)
			t.Errorf("Expected status 201, got %d. error: %v", resp.StatusCode, body["error"])
		}
	})
}
