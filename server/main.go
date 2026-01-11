package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/logger"
	"birdactyl-panel-backend/internal/middleware"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/routes"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		if err == config.ErrConfigGenerated {
			logger.Success("Generated default config.yaml")
			logger.Warn("Please configure your database settings, then restart the server.")
			os.Exit(0)
		}
		logger.Fatal("Config load failed: %v", err)
	}

	if cfg.Logging.File != "" {
		os.MkdirAll(filepath.Dir(cfg.Logging.File), 0755)
		f, err := os.OpenFile(cfg.Logging.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatal("Failed to open log file: %v", err)
		}
		defer f.Close()
		logger.SetFile(f)
	}

	log.SetOutput(logger.NewStdLogger())
	log.SetFlags(0)

	if err := database.Connect(&cfg.Database); err != nil {
		logger.Fatal("Database connection failed: %v", err)
	}
	logger.Success("Database connected (%s)", cfg.Database.Driver)

	services.InitScheduler()

	if err := plugins.StartServer(cfg.Plugins.Address); err != nil {
		logger.Error("Plugin server failed: %v", err)
	} else {
		logger.Success("Plugin server on %s", cfg.Plugins.Address)

		if cfg.Plugins.Container.Enabled {
			if err := plugins.InitPluginSystem(cfg.Plugins); err != nil {
				logger.Error("Plugin container failed: %v", err)
			} else {
				logger.Success("Plugin container started")
			}
		}

		plugins.StartScheduler()
		plugins.StartHealthCheck()
		plugins.LoadPlugins(cfg.Plugins.Directory)
	}

	app := fiber.New(fiber.Config{
		ProxyHeader:          "X-Forwarded-For",
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	plugins.RegisterPluginRoutes(app)
	middleware.CleanupRateLimitStore()

	stopSessionCleanup := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				services.CleanExpiredSessions()
			case <-stopSessionCleanup:
				return
			}
		}
	}()

	routes.SetupRoutes(app)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Warn("Shutting down...")
		close(stopSessionCleanup)
		middleware.StopCleanup()
		services.StopScheduler()
		plugins.StopHealthCheck()
		plugins.StopScheduler()
		if plugins.GetContainerManager().IsRunning() {
			plugins.GetContainerManager().Shutdown()
		} else {
			plugins.GetProcessManager().StopAll()
		}
		plugins.StopServer()
		database.Close()
		app.Shutdown()
	}()

	logger.Success("Server listening on %s", cfg.Server.Address())
	if err := app.Listen(cfg.Server.Address()); err != nil {
		logger.Fatal("Server error: %v", err)
	}
}
