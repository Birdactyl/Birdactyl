package benchmarks

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"birdactyl-panel-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func newBenchApp(rpm, burst int) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/bench", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: rpm,
		BurstLimit:        burst,
	}), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	return app
}

func BenchmarkRateLimiter_SingleClient(b *testing.B) {
	app := newBenchApp(999999999, 999999999)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		app.Test(req, -1)
	}
}

func BenchmarkRateLimiter_UniqueClients(b *testing.B) {
	app := newBenchApp(999999999, 999999999)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.%d.%d.%d", (i>>16)&0xFF, (i>>8)&0xFF, i&0xFF))
		app.Test(req, -1)
	}
}

func BenchmarkRateLimiter_Parallel(b *testing.B) {
	app := newBenchApp(999999999, 999999999)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := httptest.NewRequest("GET", "/bench", nil)
			req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.%d.%d.%d", (i>>16)&0xFF, (i>>8)&0xFF, i&0xFF))
			app.Test(req, -1)
			i++
		}
	})
}

func BenchmarkRateLimiter_HighContention(b *testing.B) {
	app := newBenchApp(999999999, 999999999)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/bench", nil)
			req.Header.Set("X-Forwarded-For", "10.0.0.1")
			app.Test(req, -1)
		}
	})
}

func BenchmarkRateLimiter_Rejection(b *testing.B) {
	app := newBenchApp(1, 1)
	exhaust := httptest.NewRequest("GET", "/bench", nil)
	exhaust.Header.Set("X-Forwarded-For", "10.99.99.99")
	app.Test(exhaust, -1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/bench", nil)
		req.Header.Set("X-Forwarded-For", "10.99.99.99")
		app.Test(req, -1)
	}
}

func BenchmarkRateLimiter_MixedEndpoints(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	paths := []string{"/read", "/write", "/strict"}
	configs := []middleware.ThousandTHRConfig{
		{RequestsPerMinute: 999999999, BurstLimit: 999999999},
		{RequestsPerMinute: 999999999, BurstLimit: 999999999},
		{RequestsPerMinute: 999999999, BurstLimit: 999999999},
	}
	for i, p := range paths {
		app.Get(p, middleware.ThousandTHR(configs[i]), func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", paths[i%len(paths)], nil)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.0.0.%d", i%255))
		app.Test(req, -1)
	}
}

func BenchmarkGetBucketCount(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.GetBucketCount()
	}
}

func BenchmarkGetShardStats(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware.GetShardStats()
	}
}
