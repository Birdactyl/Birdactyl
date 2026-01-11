package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cauthon-axis/internal/api"
	"cauthon-axis/internal/config"
	"cauthon-axis/internal/docker"
	"cauthon-axis/internal/logger"
	"cauthon-axis/internal/pairing"
	"cauthon-axis/internal/panel"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pair" {
		runPairingMode()
		return
	}

	logger.Info("Starting Cauthon Axis...")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		if err == config.ErrConfigGenerated {
			logger.Info("Generated default config.yaml")
			logger.Info("Please configure your panel URL and token, then restart Axis.")
			os.Exit(0)
		}
		logger.Fatal("Failed to load config: %v", err)
	}

	if cfg.Logging.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0755); err != nil {
			logger.Warn("Could not create log directory: %v", err)
		} else {
			f, err := os.OpenFile(cfg.Logging.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Warn("Could not open log file: %v (logging to stdout only)", err)
			} else {
				defer f.Close()
				logger.SetFile(f)
			}
		}
	}

	if len(cfg.Panel.Token) > 10 {
		logger.Info("Loaded token: %s...", cfg.Panel.Token[:10])
	} else if len(cfg.Panel.Token) > 0 {
		logger.Info("Loaded token: %s...", cfg.Panel.Token[:len(cfg.Panel.Token)/2])
	}

	ensureBirdactylUser()

	if err := ensureDataDirectories(cfg); err != nil {
		logger.Fatal("%v", err)
	}

	logger.Info("Initializing Docker...")
	if err := docker.Init(); err != nil {
		logger.Fatal("Docker initialization failed: %v", err)
	}
	logger.Success("Docker ready")

	if cfg.Panel.Token == "" {
		logger.Fatal("Panel token not configured. Create a node in the panel and add the token to config.yaml")
	}

	client := panel.NewClient()

	if err := client.SendHeartbeat(); err != nil {
		logger.Warn("Initial heartbeat failed: %v", err)
	} else {
		logger.Success("Connected to panel")
	}

	go heartbeatLoop(client)

	app := api.NewServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down...")
		app.Shutdown()
	}()

	logger.Info("API server listening on %s", cfg.Node.Listen)
	if err := app.Listen(cfg.Node.Listen); err != nil {
		logger.Fatal("%v", err)
	}
}

func heartbeatLoop(client *panel.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := client.SendHeartbeat(); err != nil {
			logger.Warn("Heartbeat failed: %v", err)
		}
	}
}

func ensureBirdactylUser() {
	if out, err := exec.Command("id", "-u", "birdactyl").Output(); err == nil {
		uid := strings.TrimSpace(string(out))
		logger.Info("birdactyl user exists (UID %s)", uid)
		return
	}

	logger.Info("Creating birdactyl user...")

	if err := exec.Command("useradd", "-r", "-M", "-s", "/bin/false", "birdactyl").Run(); err != nil {
		logger.Warn("Could not create birdactyl user: %v", err)
	} else {
		if out, _ := exec.Command("id", "-u", "birdactyl").Output(); len(out) > 0 {
			logger.Success("Created birdactyl user (UID %s)", strings.TrimSpace(string(out)))
		}
	}
}

func ensureDataDirectories(cfg *config.Config) error {
	dirs := []string{cfg.Node.DataDir, cfg.Node.BackupDir}

	isRoot := os.Geteuid() == 0

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("cannot create %s: permission denied\nRun once with sudo to set up directories: sudo ./axis", dir)
			}
			return fmt.Errorf("cannot create %s: %v", dir, err)
		}

		if isRoot {
			if out, err := exec.Command("chmod", "777", dir).CombinedOutput(); err != nil {
				logger.Warn("chmod %s failed: %v - %s", dir, err, string(out))
			}
		} else {
			os.Chmod(dir, 0777)
		}

		testFile := filepath.Join(dir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("cannot write to %s: permission denied\nRun once with sudo to fix permissions: sudo ./axis", dir)
			}
			return fmt.Errorf("cannot write to %s: %v", dir, err)
		}
		os.Remove(testFile)
	}

	logger.Success("Data directories ready")
	return nil
}

func runPairingMode() {
	logger.Info("Starting Axis in pairing mode...")

	cfg, err := config.Load("config.yaml")
	if err != nil && err != config.ErrConfigGenerated {
		logger.Fatal("Failed to load config: %v", err)
	}

	if cfg == nil {
		cfg, _ = config.Load("config.yaml")
	}

	pairing.StartPairingMode(60 * time.Second)

	app := api.NewServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down...")
		app.Shutdown()
	}()

	go func() {
		time.Sleep(65 * time.Second)
		if !pairing.IsActive() {
			logger.Info("Pairing window closed. Shutting down...")
			app.Shutdown()
		}
	}()

	logger.Info("Listening on %s", cfg.Node.Listen)
	logger.Info("Waiting for panel to connect...")
	logger.Info("")

	if err := app.Listen(cfg.Node.Listen); err != nil {
		logger.Fatal("%v", err)
	}
}
