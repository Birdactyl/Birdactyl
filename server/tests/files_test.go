package tests

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers/server"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupFileMockNodeAndServer() (*fiber.App, *models.User, *models.Server, *httptest.Server, func()) {
	mockDaemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		if r.Method == "GET" && len(r.URL.Path) > 6 && r.URL.Path[len(r.URL.Path)-7:] == "/search" {
			w.Write([]byte(`{"success":true, "data": []}`))
		} else if r.Method == "GET" && len(r.URL.Path) > 5 && r.URL.Path[len(r.URL.Path)-6:] == "/files" {
			w.Write([]byte(`{"success":true, "data": []}`))
		} else if r.Method == "GET" && len(r.URL.Path) > 9 && r.URL.Path[len(r.URL.Path)-10:] == "/read" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("mock file content"))
		} else if r.Method == "GET" && len(r.URL.Path) > 13 && r.URL.Path[len(r.URL.Path)-14:] == "/download" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename=\"file.txt\"")
			w.Write([]byte("mock file content"))
		} else {
			w.Write([]byte(`{"success": true}`))
		}
	}))

	u, _ := url.Parse(mockDaemon.URL)
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_files_admin",
		Email:    "test_files_admin@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	testNode := &models.Node{
		ID:          uuid.New(),
		Name:        "Mock Node - Files",
		FQDN:        host,
		Port:        port,
		DaemonToken: "mock_token_123",
	}
	database.DB.Create(testNode)

	testPkg := &models.Package{
		ID:   uuid.New(),
		Name: "Mock Package Files",
	}
	database.DB.Create(testPkg)

	testServer := &models.Server{
		ID:        uuid.New(),
		Name:      "Mock Server Files",
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

func TestFilesRoutes(t *testing.T) {
	requireDB(t)

	app, _, testServer, _, cleanup := setupFileMockNodeAndServer()
	defer cleanup()

	app.Get("/servers/:id/files", server.ListFiles)
	app.Get("/servers/:id/files/read", server.ReadFile)
	app.Get("/servers/:id/files/search", server.SearchFiles)
	app.Post("/servers/:id/files/folder", server.CreateFolder)
	app.Post("/servers/:id/files/write", server.WriteFile)
	app.Delete("/servers/:id/files", server.DeleteFile)
	app.Post("/servers/:id/files/move", server.MoveFile)
	app.Post("/servers/:id/files/copy", server.CopyFile)
	app.Post("/servers/:id/files/compress", server.CompressFile)
	app.Post("/servers/:id/files/decompress", server.DecompressFile)
	app.Get("/servers/:id/files/download", server.DownloadFile)
	app.Post("/servers/:id/files/bulk-delete", server.BulkDeleteFiles)
	app.Post("/servers/:id/files/bulk-copy", server.BulkCopyFiles)
	app.Post("/servers/:id/files/bulk-compress", server.BulkCompressFiles)

	t.Run("List Files", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/files?path=/", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test list files: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Read File", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/files/read?path=/data.txt", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test read file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Search Files", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/files/search?q=query_string", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test search files: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Folder", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/folder", testServer.ID.String()), toJSONBody(map[string]string{"path": "/new_folder"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test create folder: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Write File", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/write", testServer.ID.String()), toJSONBody(map[string]string{"path": "/file.txt", "content": "hello text"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test write file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Delete File", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/servers/%s/files?path=/file.txt", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test delete file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Move File", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/move", testServer.ID.String()), toJSONBody(map[string]string{"from": "/file.txt", "to": "/new_file.txt"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test move file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Copy File", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/copy", testServer.ID.String()), toJSONBody(map[string]string{"from": "/file.txt", "to": "/copy_file.txt"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test copy file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Compress File", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/compress", testServer.ID.String()), toJSONBody(map[string]string{"path": "/folder", "dest": "/folder.zip", "format": "zip"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test compress file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Decompress File", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/decompress", testServer.ID.String()), toJSONBody(map[string]string{"path": "/archive.zip", "dest": "/destination"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test decompress file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Bulk Delete", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/bulk-delete", testServer.ID.String()), toJSONBody(map[string]interface{}{"paths": []string{"/dir1", "/dir2"}}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test bulk delete: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Bulk Copy", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/bulk-copy", testServer.ID.String()), toJSONBody(map[string]interface{}{"paths": []string{"/f1", "/f2"}, "dest": "/target"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test bulk copy: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Bulk Compress", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/files/bulk-compress", testServer.ID.String()), toJSONBody(map[string]interface{}{"paths": []string{"/f1", "/f2"}, "dest": "/target.zip", "format": "zip"}))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test bulk compress: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Download File", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/servers/%s/files/download?path=/file.txt", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to test download file: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
