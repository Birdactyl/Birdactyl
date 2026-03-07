package benchmarks

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkHealthEndpoint(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/v1/health", handlers.Health)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		app.Test(req, -1)
	}
}

func BenchmarkHealthEndpoint_Parallel(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/api/v1/health", handlers.Health)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/api/v1/health", nil)
			app.Test(req, -1)
		}
	})
}

func BenchmarkFiberJSONResponse(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/bench", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"id":       "00000000-0000-0000-0000-000000000000",
				"name":     "benchmark",
				"status":   "running",
				"memory":   4096,
				"cpu":      200,
				"disk":     10240,
				"is_owner": true,
			},
		})
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberJSONResponse_Large(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/bench", func(c *fiber.Ctx) error {
		items := make([]fiber.Map, 50)
		for i := range items {
			items[i] = fiber.Map{
				"id":          fmt.Sprintf("server-%d", i),
				"name":        fmt.Sprintf("Server %d", i),
				"status":      "running",
				"memory":      4096,
				"cpu":         200,
				"disk":        10240,
				"description": "A benchmark server for testing response serialization",
			}
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data":    items,
		})
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberRouting_Static(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := func(c *fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/api/v1/health", handler)
	app.Get("/api/v1/auth/me", handler)
	app.Post("/api/v1/auth/login", handler)
	app.Get("/api/v1/servers", handler)
	app.Get("/api/v1/admin/users", handler)
	app.Get("/api/v1/admin/nodes", handler)
	app.Get("/api/v1/admin/packages", handler)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/health", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberRouting_Parameterized(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := func(c *fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/api/v1/servers/:id", handler)
	app.Get("/api/v1/servers/:id/status", handler)
	app.Get("/api/v1/servers/:id/files", handler)
	app.Get("/api/v1/servers/:id/backups", handler)
	app.Get("/api/v1/servers/:id/databases", handler)
	app.Get("/api/v1/servers/:id/subusers", handler)
	app.Get("/api/v1/servers/:id/schedules", handler)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/servers/550e8400-e29b-41d4-a716-446655440000/status", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberRouting_DeepNested(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	handler := func(c *fiber.Ctx) error { return c.SendStatus(200) }
	app.Delete("/api/v1/servers/:id/backups/:backupId", handler)
	app.Get("/api/v1/servers/:id/backups/:backupId/download", handler)
	app.Post("/api/v1/servers/:id/databases/:dbId/rotate", handler)
	app.Get("/api/v1/servers/:id/schedules/:scheduleId", handler)
	app.Patch("/api/v1/servers/:id/subusers/:subuserId", handler)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/servers/550e8400-e29b-41d4-a716-446655440000/backups/660e8400-e29b-41d4-a716-446655440001/download", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberMiddlewareChain(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	noop := func(c *fiber.Ctx) error { return c.Next() }
	app.Get("/bench", noop, noop, noop, noop, func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFiberErrorResponse(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/bench", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authorization required",
		})
	})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		app.Test(req, -1)
	}
}
