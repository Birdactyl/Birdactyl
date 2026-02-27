package api

import (
	"strings"

	"cauthon-axis/internal/config"
	"cauthon-axis/internal/logger"
	"cauthon-axis/internal/pairing"
	"cauthon-axis/internal/server"
	"cauthon-axis/internal/system"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func NewServer() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             100 * 1024 * 1024 * 1024,
	})

	app.Get("/api/health", handleHealth)
	app.Post("/api/pair", handlePairing)
	app.Get("/api/system", requirePanelAuth, handleSystemInfo)

	app.Use("/api/servers/:id/ws", func(c *fiber.Ctx) error {
		cfg := config.Get()
		token := c.Query("token")
		if token != cfg.Panel.Token {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		if err := server.ValidateServerID(c.Params("id")); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid server id"})
		}
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/api/servers/:id/ws", websocket.New(handleServerLogs))

	servers := app.Group("/api/servers", requirePanelAuth)
	servers.Post("/", handleCreateServer)
	servers.Use("/:id", validateServerID)
	servers.Use("/:id/*", validateServerID)
	servers.Get("/:id/status", handleServerStatus)
	servers.Get("/:id/logs", handleGetLogs)
	servers.Get("/:id/logs/full", handleGetFullLog)
	servers.Get("/:id/logs/search", handleSearchLogs)
	servers.Get("/:id/logs/files", handleListLogFiles)
	servers.Get("/:id/logs/file/:filename", handleReadLogFile)
	servers.Post("/:id/command", handleSendCommand)
	servers.Post("/:id/start", handleStartServer)
	servers.Post("/:id/stop", handleStopServer)
	servers.Post("/:id/kill", handleKillServer)
	servers.Post("/:id/restart", handleRestartServer)
	servers.Post("/:id/reinstall", handleReinstallServer)
	servers.Delete("/:id", handleDeleteServer)
	servers.Get("/:id/files", handleListFiles)
	servers.Get("/:id/files/hashes", handleListFilesWithHashes)
	servers.Get("/:id/files/read", handleReadFile)
	servers.Get("/:id/files/search", handleSearchFiles)
	servers.Post("/:id/files/folder", handleCreateFolder)
	servers.Post("/:id/files/write", handleWriteFile)
	servers.Post("/:id/files/upload", handleUploadFile)
	servers.Delete("/:id/files", handleDeletePath)
	servers.Post("/:id/files/move", handleMovePath)
	servers.Post("/:id/files/copy", handleCopyPath)
	servers.Post("/:id/files/compress", handleCompressPath)
	servers.Post("/:id/files/decompress", handleDecompressPath)
	servers.Get("/:id/files/download", handleDownloadFile)
	servers.Post("/:id/files/bulk-delete", handleBulkDelete)
	servers.Post("/:id/files/bulk-copy", handleBulkCopy)
	servers.Post("/:id/files/bulk-compress", handleBulkCompress)
	servers.Post("/:id/files/download-url", handleDownloadURL)
	servers.Post("/:id/modpack/install", handleInstallModpack)
	servers.Get("/:id/backups", handleListBackups)
	servers.Post("/:id/backups", handleCreateBackup)
	servers.Delete("/:id/backups/:backupId", handleDeleteBackup)
	servers.Get("/:id/backups/:backupId/download", handleDownloadBackup)
	servers.Post("/:id/backups/:backupId/restore", handleRestoreBackup)
	servers.Post("/:id/archive", handleCreateArchive)
	servers.Get("/:id/archive/download", handleDownloadArchive)
	servers.Delete("/:id/archive", handleDeleteArchive)
	servers.Post("/:id/import", handleImportServer)

	return app
}

func requirePanelAuth(c *fiber.Ctx) error {
	cfg := config.Get()
	auth := c.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")

	if token == "" || token != cfg.Panel.Token {
		logger.Warn("Auth rejected: got '%s...', expected '%s...'", token[:min(len(token), 10)], cfg.Panel.Token[:min(len(cfg.Panel.Token), 10)])
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false, "error": "Unauthorized",
		})
	}
	return c.Next()
}

func validateServerID(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := server.ValidateServerID(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "invalid server id",
		})
	}
	return c.Next()
}

func handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func handlePairing(c *fiber.Ctx) error {
	if !pairing.IsActive() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Pairing mode not active. Run 'axis pair' on the node first.",
		})
	}

	var req pairing.PairingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.PanelURL == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Panel URL and code are required",
		})
	}

	result := pairing.HandlePairingRequest(req.PanelURL, req.Code)
	if !result.Success {
		return c.Status(fiber.StatusForbidden).JSON(result)
	}

	return c.JSON(result)
}

func handleSystemInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "data": system.GetInfo()})
}

func handleCreateServer(c *fiber.Ctx) error {
	var cfg server.ServerConfig
	if err := c.BodyParser(&cfg); err != nil {
		logger.Error("Create server failed: invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Invalid request body",
		})
	}

	if err := server.ValidateServerID(cfg.ID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "invalid server id",
		})
	}

	logger.Info("Creating server %s (%s)", cfg.ID, cfg.Name)
	if err := server.Create(cfg); err != nil {
		logger.Error("Create server %s failed: %v", cfg.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	logger.Info("Server %s created, starting...", cfg.ID)
	if err := server.Start(cfg.ID); err != nil {
		logger.Error("Server %s start failed: %v", cfg.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "Created but failed to start: " + err.Error(),
		})
	}

	logger.Success("Server %s started successfully", cfg.ID)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true, "message": "Server created and started",
	})
}

func handleServerStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	status, err := server.GetStatus(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	stats, _ := server.GetStats(id)
	data := fiber.Map{"status": status}
	if stats != nil {
		data["stats"] = fiber.Map{
			"memory":       stats.MemoryUsage,
			"memory_limit": stats.MemoryLimit,
			"cpu":          stats.CPUPercent,
			"disk":         stats.DiskUsage,
			"network_rx":   stats.NetRx,
			"network_tx":   stats.NetTx,
		}
	} else {
		data["stats"] = fiber.Map{
			"disk": server.GetDiskUsage(id),
		}
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

func handleGetLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	lines := c.QueryInt("lines", 100)
	logs, err := server.GetLogLines(id, lines)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"lines": logs})
}

func handleGetFullLog(c *fiber.Ctx) error {
	id := c.Params("id")
	content, err := server.GetFullLog(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Send(content)
}

func handleSearchLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	pattern := c.Query("pattern")
	regex := c.QueryBool("regex", false)
	limit := c.QueryInt("limit", 100)
	since := int64(c.QueryInt("since", 0))
	matches, err := server.SearchLogs(id, pattern, regex, limit, since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"matches": matches})
}

func handleListLogFiles(c *fiber.Ctx) error {
	id := c.Params("id")
	files, err := server.ListLogFiles(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"files": files})
}

func handleReadLogFile(c *fiber.Ctx) error {
	id := c.Params("id")
	filename := c.Params("filename")
	content, err := server.ReadLogFile(id, filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Send(content)
}

func handleSendCommand(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Command string `json:"command"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if err := server.SendCommand(id, body.Command); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleStartServer(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("Starting server %s", id)

	var cfg server.ServerConfig
	if parseErr := c.BodyParser(&cfg); parseErr == nil && cfg.DockerImage != "" {
		cfg.ID = id
		if server.IsDataDirEmpty(id) && cfg.InstallScript != "" {
			logger.Info("Data empty, running full install for %s", id)
			if createErr := server.Create(cfg); createErr != nil {
				logger.Error("Failed to create server %s: %v", id, createErr)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false, "error": createErr.Error(),
				})
			}
		} else {
			server.RemoveContainer(id)
			logger.Info("Recreating container for %s", id)
			if createErr := server.CreateContainer(cfg); createErr != nil {
				logger.Error("Failed to create container for %s: %v", id, createErr)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false, "error": createErr.Error(),
				})
			}
		}
	}

	if err := server.Start(id); err != nil {
		logger.Error("Start server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Success("Server %s started", id)
	return c.JSON(fiber.Map{"success": true, "message": "Server started"})
}

func handleStopServer(c *fiber.Ctx) error {
	id := c.Params("id")
	timeout := c.QueryInt("timeout", 30)

	var stopCfg struct {
		StopCommand string `json:"stop_command"`
		StopSignal  string `json:"stop_signal"`
		StopTimeout int    `json:"stop_timeout"`
	}
	c.BodyParser(&stopCfg)

	if stopCfg.StopTimeout > 0 {
		timeout = stopCfg.StopTimeout
	}

	logger.Info("Stopping server %s (timeout: %ds, command: %s)", id, timeout, stopCfg.StopCommand)
	if err := server.StopWithConfig(id, timeout, stopCfg.StopCommand, stopCfg.StopSignal); err != nil {
		logger.Error("Stop server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Success("Server %s stopped", id)
	return c.JSON(fiber.Map{"success": true, "message": "Server stopped"})
}

func handleKillServer(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("Killing server %s", id)
	if err := server.Kill(id); err != nil {
		logger.Error("Kill server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Success("Server %s killed", id)
	return c.JSON(fiber.Map{"success": true, "message": "Server killed"})
}

func handleRestartServer(c *fiber.Ctx) error {
	id := c.Params("id")
	timeout := c.QueryInt("timeout", 30)

	var stopCfg struct {
		StopCommand string `json:"stop_command"`
		StopSignal  string `json:"stop_signal"`
		StopTimeout int    `json:"stop_timeout"`
	}
	c.BodyParser(&stopCfg)

	if stopCfg.StopTimeout > 0 {
		timeout = stopCfg.StopTimeout
	}

	logger.Info("Restarting server %s (timeout: %ds, command: %s)", id, timeout, stopCfg.StopCommand)
	if err := server.RestartWithConfig(id, timeout, stopCfg.StopCommand, stopCfg.StopSignal); err != nil {
		logger.Error("Restart server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Success("Server %s restarted", id)
	return c.JSON(fiber.Map{"success": true, "message": "Server restarted"})
}

func handleReinstallServer(c *fiber.Ctx) error {
	var cfg server.ServerConfig
	if err := c.BodyParser(&cfg); err != nil {
		logger.Error("Reinstall server failed: invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Invalid request body",
		})
	}

	id := c.Params("id")
	cfg.ID = id

	logger.Info("Reinstalling server %s", id)
	if err := server.Reinstall(cfg); err != nil {
		logger.Error("Reinstall server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	logger.Info("Server %s reinstalled, starting...", id)
	if err := server.Start(id); err != nil {
		logger.Error("Server %s start failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "Reinstalled but failed to start: " + err.Error(),
		})
	}

	logger.Success("Server %s reinstalled and started", id)
	return c.JSON(fiber.Map{"success": true, "message": "Server reinstalled and started"})
}

func handleDeleteServer(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("Deleting server %s", id)
	if err := server.Delete(id); err != nil {
		logger.Error("Delete server %s failed: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Server deleted"})
}
