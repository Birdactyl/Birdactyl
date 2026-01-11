package routes

import (
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/handlers/admin"
	"birdactyl-panel-backend/internal/handlers/auth"
	"birdactyl-panel-backend/internal/handlers/server"
	"birdactyl-panel-backend/internal/middleware"
	"birdactyl-panel-backend/internal/plugins"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	plugins.RegisterUIRoutes(app)
	plugins.RegisterPluginRoutes(app)

	api := app.Group("/api/v1")

	readLimit := middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 60,
		BurstLimit:        80,
	})
	writeLimit := middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 30,
		BurstLimit:        40,
	})
	strictLimit := middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 10,
		BurstLimit:        15,
	})

	api.Get("/health", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 120,
		BurstLimit:        150,
	}), handlers.Health)

	authRoutes := api.Group("/auth")
	authRoutes.Post("/register", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 5,
		BurstLimit:        5,
	}), auth.Register)
	authRoutes.Post("/login", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 10,
		BurstLimit:        10,
	}), auth.Login)
	authRoutes.Post("/refresh", middleware.ThousandTHR(middleware.ThousandTHRConfig{
		RequestsPerMinute: 30,
		BurstLimit:        30,
	}), auth.Refresh)
	authRoutes.Post("/logout", middleware.RequireAuth(), writeLimit, auth.Logout)
	authRoutes.Post("/logout-all", middleware.RequireAuth(), writeLimit, auth.LogoutAll)
	authRoutes.Get("/me", middleware.RequireAuth(), readLimit, auth.Me)
	authRoutes.Get("/resources", middleware.RequireAuth(), readLimit, auth.GetResources)
	authRoutes.Patch("/profile", middleware.RequireAuth(), writeLimit, auth.UpdateProfile)
	authRoutes.Patch("/password", middleware.RequireAuth(), strictLimit, auth.UpdatePassword)
	authRoutes.Get("/sessions", middleware.RequireAuth(), readLimit, auth.GetSessions)
	authRoutes.Delete("/sessions/:id", middleware.RequireAuth(), writeLimit, auth.RevokeSession)
	authRoutes.Delete("/sessions", middleware.RequireAuth(), strictLimit, auth.RevokeAllSessions)
	authRoutes.Get("/api-keys", middleware.RequireAuth(), readLimit, auth.GetAPIKeys)
	authRoutes.Post("/api-keys", middleware.RequireAuth(), writeLimit, auth.CreateAPIKey)
	authRoutes.Delete("/api-keys/:id", middleware.RequireAuth(), writeLimit, auth.DeleteAPIKey)

	adminRoutes := api.Group("/admin", middleware.RequireAuth(), middleware.RequireAdmin())
	adminRoutes.Get("/users", readLimit, admin.AdminGetUsers)
	adminRoutes.Post("/users", strictLimit, admin.AdminCreateUser)
	adminRoutes.Post("/users/ban", writeLimit, admin.AdminBanUsers)
	adminRoutes.Post("/users/unban", writeLimit, admin.AdminUnbanUsers)
	adminRoutes.Post("/users/delete", strictLimit, admin.AdminDeleteUsers)
	adminRoutes.Post("/users/set-admin", strictLimit, admin.AdminSetAdmin)
	adminRoutes.Post("/users/revoke-admin", strictLimit, admin.AdminRevokeAdmin)
	adminRoutes.Post("/users/force-reset", strictLimit, admin.AdminForcePasswordReset)
	adminRoutes.Patch("/users/:id", writeLimit, admin.AdminUpdateUser)
	adminRoutes.Get("/users/:userId/api-keys", readLimit, admin.AdminGetUserAPIKeys)
	adminRoutes.Post("/users/:userId/api-keys", writeLimit, admin.AdminCreateUserAPIKey)
	adminRoutes.Delete("/users/:userId/api-keys/:keyId", writeLimit, admin.AdminDeleteUserAPIKey)

	adminRoutes.Get("/nodes", readLimit, handlers.AdminGetNodes)
	adminRoutes.Post("/nodes", strictLimit, handlers.AdminCreateNode)
	adminRoutes.Post("/nodes/pair", strictLimit, handlers.AdminPairNode)
	adminRoutes.Get("/nodes/pairing-code", readLimit, handlers.AdminGeneratePairingCode)
	adminRoutes.Post("/nodes/refresh", writeLimit, handlers.AdminRefreshNodes)
	adminRoutes.Get("/nodes/:id", readLimit, handlers.AdminGetNode)
	adminRoutes.Patch("/nodes/:id", writeLimit, handlers.AdminUpdateNode)
	adminRoutes.Delete("/nodes/:id", strictLimit, handlers.AdminDeleteNode)
	adminRoutes.Post("/nodes/:id/reset-token", strictLimit, handlers.AdminResetNodeToken)

	adminRoutes.Get("/packages", readLimit, handlers.AdminGetPackages)
	adminRoutes.Post("/packages", strictLimit, handlers.AdminCreatePackage)
	adminRoutes.Get("/packages/:id", readLimit, handlers.AdminGetPackage)
	adminRoutes.Patch("/packages/:id", writeLimit, handlers.AdminUpdatePackage)
	adminRoutes.Delete("/packages/:id", strictLimit, handlers.AdminDeletePackage)

	adminRoutes.Get("/servers", readLimit, server.AdminGetServers)
	adminRoutes.Post("/servers", strictLimit, server.AdminCreateServer)
	adminRoutes.Post("/servers/:id/view", readLimit, server.AdminViewServer)
	adminRoutes.Post("/servers/suspend", writeLimit, server.AdminSuspendServers)
	adminRoutes.Post("/servers/unsuspend", writeLimit, server.AdminUnsuspendServers)
	adminRoutes.Post("/servers/delete", strictLimit, server.AdminDeleteServers)
	adminRoutes.Patch("/servers/:id/resources", writeLimit, server.AdminUpdateServerResources)
	adminRoutes.Post("/servers/:id/transfer", strictLimit, server.AdminTransferServer)
	adminRoutes.Get("/transfers", readLimit, server.AdminGetAllTransfers)
	adminRoutes.Get("/transfers/:transferId", readLimit, server.AdminGetTransferStatus)

	adminRoutes.Get("/logs", readLimit, admin.AdminGetLogs)

	adminRoutes.Get("/ip-bans", readLimit, admin.AdminGetIPBans)
	adminRoutes.Post("/ip-bans", strictLimit, admin.AdminCreateIPBan)
	adminRoutes.Delete("/ip-bans/:id", strictLimit, admin.AdminDeleteIPBan)

	adminRoutes.Get("/settings/registration", readLimit, admin.AdminGetRegistrationStatus)
	adminRoutes.Patch("/settings/registration", strictLimit, admin.AdminSetRegistrationStatus)

	adminRoutes.Get("/settings/server-creation", readLimit, admin.AdminGetServerCreationStatus)
	adminRoutes.Patch("/settings/server-creation", strictLimit, admin.AdminSetServerCreationStatus)

	adminRoutes.Get("/database-hosts", readLimit, admin.AdminGetDatabaseHosts)
	adminRoutes.Post("/database-hosts", strictLimit, admin.AdminCreateDatabaseHost)
	adminRoutes.Patch("/database-hosts/:id", writeLimit, admin.AdminUpdateDatabaseHost)
	adminRoutes.Delete("/database-hosts/:id", strictLimit, admin.AdminDeleteDatabaseHost)
	adminRoutes.Get("/database-hosts/:id/databases", readLimit, admin.AdminGetHostDatabases)
	adminRoutes.Delete("/database-hosts/:id/databases/:dbId", strictLimit, admin.AdminDeleteDatabase)

	adminRoutes.Get("/plugins", readLimit, admin.AdminListPlugins)
	adminRoutes.Get("/plugins/config", readLimit, admin.AdminGetPluginConfig)
	adminRoutes.Get("/plugins/files", readLimit, admin.AdminListPluginFiles)
	adminRoutes.Post("/plugins", strictLimit, admin.AdminLoadPlugin)
	adminRoutes.Post("/plugins/install-source", strictLimit, admin.AdminInstallPluginFromSource)
	adminRoutes.Post("/plugins/install-release", strictLimit, admin.AdminInstallPluginFromRelease)
	adminRoutes.Post("/plugins/upload", strictLimit, admin.AdminUploadPlugin)
	adminRoutes.Post("/plugins/:id/reload", writeLimit, admin.AdminReloadPlugin)
	adminRoutes.Delete("/plugins/:id", strictLimit, admin.AdminUnloadPlugin)
	adminRoutes.Delete("/plugins/file/:filename", strictLimit, admin.AdminDeletePluginFile)

	api.Get("/packages", middleware.RequireAuth(), readLimit, handlers.GetAvailablePackages)
	api.Get("/nodes", middleware.RequireAuth(), readLimit, handlers.GetAvailableNodes)

	api.Use("/servers/:id/logs", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, middleware.WebSocketAuth())
	api.Get("/servers/:id/logs", websocket.New(handlers.ServerLogsWS))

	servers := api.Group("/servers", middleware.RequireAuth())
	servers.Get("/", readLimit, server.GetServers)
	servers.Post("/", strictLimit, server.CreateServer)
	servers.Get("/:id", readLimit, server.GetServer)
	servers.Post("/:id/start", writeLimit, server.StartServer)
	servers.Post("/:id/stop", writeLimit, server.StopServer)
	servers.Post("/:id/restart", writeLimit, server.RestartServer)
	servers.Post("/:id/reinstall", writeLimit, server.ReinstallServer)
	servers.Post("/:id/kill", writeLimit, server.KillServer)
	servers.Post("/:id/command", writeLimit, server.SendCommand)
	servers.Get("/:id/status", readLimit, server.GetServerStatus)
	servers.Get("/:id/console", readLimit, server.GetConsoleLogs)
	servers.Delete("/:id", strictLimit, server.DeleteServer)
	servers.Post("/:id/allocations", strictLimit, server.AddAllocation)
	servers.Put("/:id/allocations/primary", writeLimit, server.SetPrimaryAllocation)
	servers.Delete("/:id/allocations", strictLimit, server.DeleteAllocation)
	servers.Patch("/:id/resources", writeLimit, server.UpdateServerResources)
	servers.Patch("/:id/name", writeLimit, server.UpdateServerName)
	servers.Patch("/:id/variables", writeLimit, server.UpdateServerVariables)
	servers.Get("/:id/backups", readLimit, server.ListBackups)
	servers.Post("/:id/backups", writeLimit, server.CreateBackup)
	servers.Delete("/:id/backups/:backupId", strictLimit, server.DeleteBackup)
	servers.Get("/:id/backups/:backupId/download", readLimit, server.DownloadBackup)
	servers.Post("/:id/backups/:backupId/restore", strictLimit, server.RestoreBackup)
	servers.Get("/:id/files", readLimit, server.ListFiles)
	servers.Get("/:id/files/read", readLimit, server.ReadFile)
	servers.Get("/:id/files/search", readLimit, server.SearchFiles)
	servers.Post("/:id/files/folder", writeLimit, server.CreateFolder)
	servers.Post("/:id/files/write", writeLimit, server.WriteFile)
	servers.Post("/:id/files/upload", writeLimit, server.UploadFile)
	servers.Delete("/:id/files", writeLimit, server.DeleteFile)
	servers.Post("/:id/files/move", writeLimit, server.MoveFile)
	servers.Post("/:id/files/copy", writeLimit, server.CopyFile)
	servers.Post("/:id/files/compress", strictLimit, server.CompressFile)
	servers.Post("/:id/files/decompress", strictLimit, server.DecompressFile)
	servers.Get("/:id/files/download", readLimit, server.DownloadFile)
	servers.Post("/:id/files/bulk-delete", writeLimit, server.BulkDeleteFiles)
	servers.Post("/:id/files/bulk-copy", writeLimit, server.BulkCopyFiles)
	servers.Post("/:id/files/bulk-compress", strictLimit, server.BulkCompressFiles)
	servers.Get("/:id/permissions", readLimit, handlers.GetMyPermissions)
	servers.Get("/:id/subusers", readLimit, handlers.GetSubusers)
	servers.Post("/:id/subusers", writeLimit, handlers.AddSubuser)
	servers.Patch("/:id/subusers/:subuserId", writeLimit, handlers.UpdateSubuser)
	servers.Delete("/:id/subusers/:subuserId", writeLimit, handlers.RemoveSubuser)
	servers.Get("/:id/addons/sources", readLimit, server.GetAddonSources)
	servers.Get("/:id/addons/search", readLimit, server.SearchAddons)
	servers.Get("/:id/addons/versions", readLimit, server.GetAddonVersions)
	servers.Get("/:id/addons/installed", readLimit, server.ListInstalledAddons)
	servers.Post("/:id/addons/install", writeLimit, server.InstallAddon)
	servers.Delete("/:id/addons", writeLimit, server.DeleteAddon)
	servers.Get("/:id/modpacks/search", readLimit, server.SearchModpacks)
	servers.Get("/:id/modpacks/versions", readLimit, server.GetModpackVersions)
	servers.Post("/:id/modpacks/install", writeLimit, server.InstallModpack)
	servers.Get("/:id/databases", readLimit, server.GetServerDatabases)
	servers.Get("/:id/databases/hosts", readLimit, server.GetDatabaseHosts)
	servers.Post("/:id/databases", writeLimit, server.CreateServerDatabase)
	servers.Post("/:id/databases/:dbId/rotate", writeLimit, server.RotateDatabasePassword)
	servers.Delete("/:id/databases/:dbId", writeLimit, server.DeleteServerDatabase)
	servers.Get("/:id/schedules", readLimit, handlers.GetServerSchedules)
	servers.Post("/:id/schedules", writeLimit, handlers.CreateSchedule)
	servers.Get("/:id/schedules/:scheduleId", readLimit, handlers.GetSchedule)
	servers.Put("/:id/schedules/:scheduleId", writeLimit, handlers.UpdateSchedule)
	servers.Delete("/:id/schedules/:scheduleId", writeLimit, handlers.DeleteSchedule)
	servers.Post("/:id/schedules/:scheduleId/run", writeLimit, handlers.RunScheduleNow)
	servers.Get("/:id/activity", readLimit, server.GetServerActivity)

	internal := api.Group("/internal")
	nodes := internal.Group("/nodes", middleware.RequireNodeAuth())
	nodes.Post("/heartbeat", handlers.NodeHeartbeat)
}
