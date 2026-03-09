package tests

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers/server"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupBackupMockNodeAndServer() (*fiber.App, *models.User, *models.Server, *httptest.Server, func()) {
	mockDaemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/backups") {
			w.Write([]byte(`{"success":true, "data": []}`))
		} else if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/download") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("mock backup content"))
		} else {
			w.Write([]byte(`{"success": true}`))
		}
	}))

	u, _ := url.Parse(mockDaemon.URL)
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_backups_admin",
		Email:    "test_backups_admin@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	testNode := &models.Node{
		ID:          uuid.New(),
		Name:        "Mock Node - Backups",
		FQDN:        host,
		Port:        port,
		TokenID:     uuid.New().String(),
		DaemonToken: "mock_token_123",
	}
	database.DB.Create(testNode)

	testPkg := &models.Package{
		ID:          uuid.New(),
		Name:        "Mock Package Backups",
	}
	database.DB.Create(testPkg)

	testServer := &models.Server{
		ID:        uuid.New(),
		Name:      "Mock Server Backups",
		NodeID:    testNode.ID,
		UserID:    adminUser.ID,
		PackageID: testPkg.ID,
	}
	database.DB.Create(testServer)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", adminUser)
		return c.Next()
	})

	cleanup := func() {
		mockDaemon.Close()
		database.DB.Where("id = ?", testServer.ID).Delete(&models.Server{})
		database.DB.Where("id = ?", testPkg.ID).Delete(&models.Package{})
		database.DB.Where("id = ?", testNode.ID).Delete(&models.Node{})
		database.DB.Where("id = ?", adminUser.ID).Delete(&models.User{})
	}

	return app, adminUser, testServer, mockDaemon, cleanup
}

func TestBackupsRoutes(t *testing.T) {
	requireDB(t)

	app, _, testServer, _, cleanup := setupBackupMockNodeAndServer()
	defer cleanup()

	app.Get("/servers/:id/backups", server.ListBackups)
	app.Post("/servers/:id/backups", server.CreateBackup)
	app.Delete("/servers/:id/backups/:backupId", server.DeleteBackup)
	app.Get("/servers/:id/backups/:backupId/download", server.DownloadBackup)
	app.Post("/servers/:id/backups/:backupId/restore", server.RestoreBackup)

	t.Run("List Backups", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/backups", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test list backups: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Backup", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/backups", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test create backup: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode)
		}
	})

	t.Run("Delete Backup", func(t *testing.T) {
		mockBackupID := uuid.New().String()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/servers/%s/backups/%s", testServer.ID.String(), mockBackupID), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test delete backup: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Restore Backup", func(t *testing.T) {
		mockBackupID := uuid.New().String()
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/backups/%s/restore", testServer.ID.String(), mockBackupID), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test restore backup: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Download Backup Option", func(t *testing.T) {
		mockBackupID := uuid.New().String()
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/backups/%s/download", testServer.ID.String(), mockBackupID), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test download backup: %v", err)
		}
		
		if resp.StatusCode != fiber.StatusFound {
			t.Errorf("Expected status 302, got %d", resp.StatusCode)
		}
	})
}
