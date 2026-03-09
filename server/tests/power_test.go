package tests

import (
	"encoding/json"
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

func setupMockNodeAndServer() (*fiber.App, *models.User, *models.Server, *httptest.Server, func()) {
	mockDaemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		errResp := `{"success": false, "error": "not implemented in mock"}`

		switch {
		case r.Method == "POST" && r.URL.Path == fmt.Sprintf("/api/servers/%s/start", ":id"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case r.Method == "POST" && (len(r.URL.Path) > 5 && r.URL.Path[len(r.URL.Path)-6:] == "/start"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case r.Method == "POST" && (len(r.URL.Path) > 4 && r.URL.Path[len(r.URL.Path)-5:] == "/stop"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case r.Method == "POST" && (len(r.URL.Path) > 4 && r.URL.Path[len(r.URL.Path)-5:] == "/kill"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case r.Method == "POST" && (len(r.URL.Path) > 7 && r.URL.Path[len(r.URL.Path)-8:] == "/restart"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		case r.Method == "POST" && (len(r.URL.Path) > 9 && r.URL.Path[len(r.URL.Path)-10:] == "/reinstall"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errResp))
		}
	}))

	u, _ := url.Parse(mockDaemon.URL)
	host, portStr, _ := net.SplitHostPort(u.Host)
	port, _ := strconv.Atoi(portStr)

	adminUser := &models.User{
		ID:       uuid.New(),
		Username: "test_power_admin",
		Email:    "test_power_admin@test.com",
		IsAdmin:  true,
	}
	database.DB.Create(adminUser)

	testNode := &models.Node{
		ID:          uuid.New(),
		Name:        "Mock Node - Power Tests",
		FQDN:        host,
		Port:        port,
		DaemonToken: "mock_token_123",
	}
	database.DB.Create(testNode)

	testPkg := &models.Package{
		ID:          uuid.New(),
		Name:        "Mock Package",
		DockerImage: "alpine:latest",
		StopCommand: "^C",
		StopSignal:  "SIGTERM",
		StopTimeout: 10,
	}
	database.DB.Create(testPkg)

	testServer := &models.Server{
		ID:        uuid.New(),
		Name:      "Mock Server",
		NodeID:    testNode.ID,
		UserID:    adminUser.ID,
		PackageID: testPkg.ID,
		Memory:    512,
		CPU:       100,
		Disk:      1024,
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

func TestPowerRoutes(t *testing.T) {
	requireDB(t)

	app, _, testServer, _, cleanup := setupMockNodeAndServer()
	defer cleanup()

	app.Post("/servers/:id/power/start", server.StartServer)
	app.Post("/servers/:id/power/stop", server.StopServer)
	app.Post("/servers/:id/power/kill", server.KillServer)
	app.Post("/servers/:id/power/restart", server.RestartServer)
	app.Post("/servers/:id/reinstall", server.ReinstallServer)

	t.Run("Start Server", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/power/start", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			var body map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&body)
			t.Errorf("Expected status 200, got %d. error: %v", resp.StatusCode, body["error"])
		}
	})

	t.Run("Stop Server", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/power/stop", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to stop server: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Kill Server", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/power/kill", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to kill server: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Restart Server", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/power/restart", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to restart server: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Reinstall Server", func(t *testing.T) {
		req := httptest.NewRequest("POST", fmt.Sprintf("/servers/%s/reinstall", testServer.ID.String()), nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Failed to reinstall server: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
