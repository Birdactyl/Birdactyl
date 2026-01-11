package server

import (
	"errors"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var errHandled = errors.New("handled")

func checkServerPerm(c *fiber.Ctx, serverID uuid.UUID, perm string) (*models.Server, error) {
	user := c.Locals("user").(*models.User)

	var server models.Server
	if err := database.DB.Preload("Node").Preload("Package").Where("id = ?", serverID).First(&server).Error; err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
		return nil, errHandled
	}

	if user.IsAdmin || server.UserID == user.ID {
		return &server, nil
	}

	if !services.HasServerPermission(user.ID, serverID, false, perm) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Permission denied"})
		return nil, errHandled
	}

	return &server, nil
}

func GetServers(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	servers, err := services.GetServersByUser(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "Failed to fetch servers",
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinServerList), map[string]interface{}{"user_id": user.ID.String(), "servers": servers}, func(input map[string]interface{}) (interface{}, error) {
		return input["servers"], nil
	})
	if result != nil {
		return c.JSON(fiber.Map{"success": true, "data": result})
	}

	return c.JSON(fiber.Map{"success": true, "data": servers})
}

func GetServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Invalid server ID",
		})
	}

	server, err := services.GetServerByID(serverID, user.ID, user.IsAdmin)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "error": "Server not found",
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinServerGet), map[string]interface{}{"server_id": serverID.String(), "server": server}, func(input map[string]interface{}) (interface{}, error) {
		return input["server"], nil
	})
	if result != nil {
		server = result.(*models.Server)
	}

	return c.JSON(fiber.Map{"success": true, "data": server})
}

func CreateServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	if !user.IsAdmin && !services.IsServerCreationEnabled() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false, "error": "Server creation is currently disabled",
		})
	}

	var req services.CreateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Invalid request body",
		})
	}

	if req.Name == "" || req.NodeID == uuid.Nil || req.PackageID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Name, node_id, and package_id are required",
		})
	}

	if req.Memory < 128 {
		req.Memory = 128
	}
	if req.CPU < 25 {
		req.CPU = 25
	}
	if req.Disk < 256 {
		req.Disk = 256
	}

	cfg := config.Get()
	if !user.IsAdmin && cfg.Resources.Enabled {
		used := services.GetUserResourceUsage(user.ID)

		ramLimit := cfg.Resources.DefaultRAM
		cpuLimit := cfg.Resources.DefaultCPU
		diskLimit := cfg.Resources.DefaultDisk
		serverLimit := cfg.Resources.MaxServers

		if user.RAMLimit != nil {
			ramLimit = *user.RAMLimit
		}
		if user.CPULimit != nil {
			cpuLimit = *user.CPULimit
		}
		if user.DiskLimit != nil {
			diskLimit = *user.DiskLimit
		}
		if user.ServerLimit != nil {
			serverLimit = *user.ServerLimit
		}

		if used.Servers >= serverLimit {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false, "error": "Maximum server limit reached",
			})
		}
		if used.RAM+req.Memory > ramLimit {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false, "error": "Not enough RAM available",
			})
		}
		if used.CPU+req.CPU > cpuLimit {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false, "error": "Not enough CPU available",
			})
		}
		if used.Disk+req.Disk > diskLimit {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false, "error": "Not enough disk space available",
			})
		}
	}

	if allow, msg := plugins.Emit(plugins.EventServerCreating, map[string]string{"user_id": user.ID.String(), "name": req.Name}); !allow {
		if msg == "" {
			msg = "Server creation blocked by plugin"
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"user_id":    user.ID.String(),
		"name":       req.Name,
		"node_id":    req.NodeID.String(),
		"package_id": req.PackageID.String(),
		"memory":     req.Memory,
		"cpu":        req.CPU,
		"disk":       req.Disk,
	}

	result, err := plugins.ExecuteMixin(string(plugins.MixinServerCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		if name, ok := input["name"].(string); ok {
			req.Name = name
		}
		if mem, ok := input["memory"].(float64); ok {
			req.Memory = int(mem)
		}
		if cpu, ok := input["cpu"].(float64); ok {
			req.CPU = int(cpu)
		}
		if disk, ok := input["disk"].(float64); ok {
			req.Disk = int(disk)
		}

		return services.CreateServer(user.ID, req)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrNodeNotFound, services.ErrPackageNotFound:
			status = fiber.StatusBadRequest
		case services.ErrNodeOffline:
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	server, ok := result.(*models.Server)
	if !ok {
		if resultMap, ok := result.(map[string]interface{}); ok {
			handlers.Log(c, user, handlers.ActionServerCreate, "Created server (mixin)", resultMap)
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": resultMap})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": result})
	}

	handlers.Log(c, user, handlers.ActionServerCreate, "Created server: "+server.Name, map[string]interface{}{"server_id": server.ID, "server_name": server.Name})

	go func() {
		if err := services.SendCreateServer(server); err != nil {
			services.UpdateServerStatus(server.ID, models.ServerStatusFailed, "")
		} else {
			services.UpdateServerStatus(server.ID, models.ServerStatusRunning, "")
		}
		plugins.Emit(plugins.EventServerCreated, map[string]string{"server_id": server.ID.String(), "name": server.Name, "user_id": server.UserID.String()})
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": server})
}

func DeleteServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "Invalid server ID",
		})
	}

	server, _ := services.GetServerByID(serverID, user.ID, user.IsAdmin)

	if allow, msg := plugins.Emit(plugins.EventServerDeleting, map[string]string{"server_id": serverID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	serverName := ""
	if server != nil {
		serverName = server.Name
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      serverName,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		services.SendDeleteServer(serverID)
		return nil, services.DeleteServer(serverID, user.ID, user.IsAdmin)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "error": "Server not found",
		})
	}

	handlers.Log(c, user, handlers.ActionServerDelete, "Deleted server: "+serverName, map[string]interface{}{"server_id": serverID})
	plugins.Emit(plugins.EventServerDeleted, map[string]string{"server_id": serverID.String(), "name": serverName})

	return c.JSON(fiber.Map{"success": true, "message": "Server deleted"})
}

func StartServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermPowerStart)
	if err != nil {
		return nil
	}

	if suspended, _ := services.IsServerSuspended(serverID); suspended {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Server is suspended"})
	}

	if allow, msg := plugins.Emit(plugins.EventServerStarting, map[string]string{"server_id": serverID.String(), "name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerStart), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SendStartServer(serverID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerStart, "Started server: "+server.Name, map[string]interface{}{"server_id": serverID})
	services.UpdateServerStatus(serverID, models.ServerStatusRunning, "")
	plugins.Emit(plugins.EventServerStarted, map[string]string{"server_id": serverID.String(), "name": server.Name})
	return c.JSON(fiber.Map{"success": true, "message": "Server started"})
}

func StopServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermPowerStop)
	if err != nil {
		return nil
	}

	if allow, msg := plugins.Emit(plugins.EventServerStopping, map[string]string{"server_id": serverID.String(), "name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerStop), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SendStopServer(serverID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerStop, "Stopped server: "+server.Name, map[string]interface{}{"server_id": serverID})
	services.UpdateServerStatus(serverID, models.ServerStatusStopped, "")
	plugins.Emit(plugins.EventServerStopped, map[string]string{"server_id": serverID.String(), "name": server.Name})
	return c.JSON(fiber.Map{"success": true, "message": "Server stopped"})
}

func KillServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermPowerKill)
	if err != nil {
		return nil
	}

	if allow, msg := plugins.Emit(plugins.EventServerKilling, map[string]string{"server_id": serverID.String(), "name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerKill), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SendKillServer(serverID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerKill, "Killed server: "+server.Name, map[string]interface{}{"server_id": serverID})
	services.UpdateServerStatus(serverID, models.ServerStatusStopped, "")
	return c.JSON(fiber.Map{"success": true, "message": "Server killed"})
}

func RestartServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermPowerRestart)
	if err != nil {
		return nil
	}

	if allow, msg := plugins.Emit(plugins.EventServerRestarting, map[string]string{"server_id": serverID.String(), "name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerRestart), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SendRestartServer(serverID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerRestart, "Restarted server: "+server.Name, map[string]interface{}{"server_id": serverID})
	return c.JSON(fiber.Map{"success": true, "message": "Server restarted"})
}

func ReinstallServer(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermReinstall)
	if err != nil {
		return nil
	}

	if allow, msg := plugins.Emit(plugins.EventServerReinstall, map[string]string{"server_id": serverID.String(), "name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerReinstall), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SendReinstallServer(server)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerReinstall, "Reinstalled server: "+server.Name, map[string]interface{}{"server_id": serverID})
	services.UpdateServerStatus(serverID, models.ServerStatusInstalling, "")
	return c.JSON(fiber.Map{"success": true, "message": "Server reinstalling"})
}

func UpdateServerResources(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermSettingsResources)
	if err != nil {
		return nil
	}

	var req struct {
		Memory int `json:"memory"`
		CPU    int `json:"cpu"`
		Disk   int `json:"disk"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	updates := map[string]interface{}{}
	if req.Memory > 0 {
		updates["memory"] = req.Memory
	}
	if req.CPU > 0 {
		updates["cpu"] = req.CPU
	}
	if req.Disk > 0 {
		updates["disk"] = req.Disk
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No changes provided"})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"name":      server.Name,
		"updates":   updates,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionServerResourcesUpdate, "Updated server resources", map[string]interface{}{"server_id": serverID, "updates": updates})

	plugins.Emit(plugins.EventServerUpdated, map[string]string{"server_id": serverID.String(), "update_type": "resources"})

	return c.JSON(fiber.Map{"success": true, "message": "Resources updated"})
}

func UpdateServerName(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	_, err = checkServerPerm(c, serverID, models.PermSettingsRename)
	if err != nil {
		return nil
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	updates := map[string]interface{}{}
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No changes provided"})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"updates":   updates,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionServerNameUpdate, "Updated server settings", map[string]interface{}{"server_id": serverID, "updates": updates})

	plugins.Emit(plugins.EventServerUpdated, map[string]string{"server_id": serverID.String(), "update_type": "name"})

	return c.JSON(fiber.Map{"success": true, "message": "Server updated"})
}

func UpdateServerVariables(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermStartupUpdate)
	if err != nil {
		return nil
	}

	var req struct {
		Variables   map[string]string `json:"variables"`
		Startup     string            `json:"startup"`
		DockerImage string            `json:"docker_image"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	mixinInput := map[string]interface{}{
		"server_id":    serverID.String(),
		"name":         server.Name,
		"variables":    req.Variables,
		"startup":      req.Startup,
		"docker_image": req.DockerImage,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinServerUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		_, updateErr := services.UpdateServerVariables(serverID, user.ID, req.Variables, req.Startup, req.DockerImage, user.IsAdmin)
		return nil, updateErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerVariablesUpdate, "Updated server variables", map[string]interface{}{"server_id": serverID})

	plugins.Emit(plugins.EventServerUpdated, map[string]string{"server_id": serverID.String(), "update_type": "variables"})

	return c.JSON(fiber.Map{"success": true, "message": "Variables updated"})
}

func SendCommand(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermConsoleWrite)
	if err != nil {
		return nil
	}

	var req struct {
		Command string `json:"command"`
	}
	if err := c.BodyParser(&req); err != nil || req.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Command is required"})
	}

	if err := services.SendCommand(serverID, req.Command); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionServerCommand, "Sent command to server", map[string]interface{}{"server_id": serverID, "command": req.Command, "server_name": server.Name})

	return c.JSON(fiber.Map{"success": true, "message": "Command sent"})
}

func GetServerStatus(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	_, err = checkServerPerm(c, serverID, models.PermConsoleRead)
	if err != nil {
		return nil
	}

	stats := services.GetServerStats(serverID)
	if stats == nil {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status": "offline",
				"stats":  nil,
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"status": stats.State,
			"stats": fiber.Map{
				"memory":       stats.MemoryBytes,
				"memory_limit": stats.MemoryLimit,
				"cpu":          stats.CPUPercent,
				"disk":         stats.DiskBytes,
				"network_rx":   stats.NetworkRx,
				"network_tx":   stats.NetworkTx,
			},
		},
	})
}

func GetConsoleLogs(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	_, err = checkServerPerm(c, serverID, models.PermConsoleRead)
	if err != nil {
		return nil
	}

	lines := c.QueryInt("lines", 100)
	if lines < 1 {
		lines = 1
	}
	if lines > 1000 {
		lines = 1000
	}

	logs := services.GetConsoleLog(serverID, lines)
	if logs == nil {
		logs = []string{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"lines": logs,
		},
	})
}
