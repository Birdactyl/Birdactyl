package server

import (
	"strings"
	"sync"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AdminGetServers(c *fiber.Ctx) error {
	servers, err := services.GetAllServers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "Failed to fetch servers",
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": servers})
}

func AdminCreateServer(c *fiber.Ctx) error {
	admin := c.Locals("user").(*models.User)

	var req struct {
		Name      string    `json:"name"`
		NodeID    uuid.UUID `json:"node_id"`
		PackageID uuid.UUID `json:"package_id"`
		Memory    int       `json:"memory"`
		CPU       int       `json:"cpu"`
		Disk      int       `json:"disk"`
		UserID    string    `json:"user_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.Name == "" || req.NodeID == uuid.Nil || req.PackageID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Name, node_id, and package_id are required"})
	}

	ownerID := admin.ID
	ownerName := admin.Username

	if req.UserID != "" {
		var user models.User
		var found bool
		if uid, err := uuid.Parse(req.UserID); err == nil {
			if database.DB.Where("id = ?", uid).First(&user).Error == nil {
				found = true
			}
		}
		if !found && len(req.UserID) >= 8 {
			var users []models.User
			database.DB.Select("id, username").Find(&users)
			for _, u := range users {
				if strings.HasPrefix(u.ID.String(), req.UserID) {
					database.DB.Where("id = ?", u.ID).First(&user)
					found = true
					break
				}
			}
		}
		if !found {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "User not found"})
		}
		ownerID = user.ID
		ownerName = user.Username
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

	createReq := services.CreateServerRequest{
		Name:      req.Name,
		NodeID:    req.NodeID,
		PackageID: req.PackageID,
		Memory:    req.Memory,
		CPU:       req.CPU,
		Disk:      req.Disk,
	}

	server, err := services.CreateServer(ownerID, createReq)
	if err != nil {
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrNodeNotFound, services.ErrPackageNotFound:
			status = fiber.StatusBadRequest
		case services.ErrNodeOffline:
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerCreate, "Created server for "+ownerName+": "+server.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"server_id": server.ID, "server_name": server.Name, "owner_id": ownerID})

	go func() {
		if err := services.SendCreateServer(server); err != nil {
			services.UpdateServerStatus(server.ID, models.ServerStatusFailed, "")
		} else {
			services.UpdateServerStatus(server.ID, models.ServerStatusRunning, "")
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": server})
}

func AdminViewServer(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var server models.Server
	if err := database.DB.Preload("Node").Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerView, "Viewing server: "+server.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"server_id": serverID, "server_name": server.Name})

	return c.JSON(fiber.Map{"success": true, "data": server})
}

func AdminSuspendServers(c *fiber.Ctx) error {
	var req struct {
		ServerIDs []string `json:"server_ids"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.ServerIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	affected := 0
	var serverNames []string
	for _, id := range req.ServerIDs {
		serverID, err := uuid.Parse(id)
		if err != nil {
			continue
		}

		var server models.Server
		if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
			continue
		}

		mixinInput := map[string]interface{}{
			"server_id": serverID.String(),
			"name":      server.Name,
		}

		_, err = plugins.ExecuteMixin(string(plugins.MixinServerSuspend), mixinInput, func(input map[string]interface{}) (interface{}, error) {
			if server.Status == models.ServerStatusRunning {
				services.SendKillServer(serverID)
				services.UpdateServerStatus(serverID, models.ServerStatusStopped, "")
			}
			return nil, services.SuspendServer(serverID)
		})

		if err == nil {
			affected++
			serverNames = append(serverNames, server.Name)
		}
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerSuspend, "Suspended servers: "+strings.Join(serverNames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"servers": serverNames})

	for _, name := range serverNames {
		plugins.Emit(plugins.EventServerSuspended, map[string]string{"server_name": name})
	}

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}

func AdminUnsuspendServers(c *fiber.Ctx) error {
	var req struct {
		ServerIDs []string `json:"server_ids"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.ServerIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	affected := 0
	var serverNames []string
	for _, id := range req.ServerIDs {
		serverID, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		var server models.Server
		if database.DB.Where("id = ?", serverID).First(&server).Error != nil {
			continue
		}

		mixinInput := map[string]interface{}{
			"server_id": serverID.String(),
			"name":      server.Name,
		}

		_, err = plugins.ExecuteMixin(string(plugins.MixinServerUnsuspend), mixinInput, func(input map[string]interface{}) (interface{}, error) {
			return nil, services.UnsuspendServer(serverID)
		})

		if err == nil {
			affected++
			serverNames = append(serverNames, server.Name)
		}
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerUnsuspend, "Unsuspended servers: "+strings.Join(serverNames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"servers": serverNames})

	for _, name := range serverNames {
		plugins.Emit(plugins.EventServerUnsuspended, map[string]string{"server_name": name})
	}

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}

func AdminDeleteServers(c *fiber.Ctx) error {
	var req struct {
		ServerIDs []string `json:"server_ids"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.ServerIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	affected := 0
	var serverNames []string

	var servers []models.Server
	var serverIDs []uuid.UUID
	for _, id := range req.ServerIDs {
		if sid, err := uuid.Parse(id); err == nil {
			serverIDs = append(serverIDs, sid)
		}
	}
	database.DB.Where("id IN ?", serverIDs).Find(&servers)
	serverNameMap := make(map[uuid.UUID]string)
	for _, s := range servers {
		serverNameMap[s.ID] = s.Name
	}

	for _, id := range req.ServerIDs {
		serverID, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(sid uuid.UUID) {
			defer wg.Done()

			mixinInput := map[string]interface{}{
				"server_id": sid.String(),
				"name":      serverNameMap[sid],
			}

			_, err := plugins.ExecuteMixin(string(plugins.MixinServerDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
				services.SendDeleteServer(sid)
				return nil, services.DeleteServerAdmin(sid)
			})

			if err == nil {
				mu.Lock()
				affected++
				if name, ok := serverNameMap[sid]; ok {
					serverNames = append(serverNames, name)
				}
				mu.Unlock()
			}
		}(serverID)
	}
	wg.Wait()

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerDelete, "Deleted servers: "+strings.Join(serverNames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"servers": serverNames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}

func AdminUpdateServerResources(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var req struct {
		Name   string `json:"name"`
		UserID string `json:"user_id"`
		Memory int    `json:"memory"`
		CPU    int    `json:"cpu"`
		Disk   int    `json:"disk"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	updates := map[string]interface{}{}
	if req.Name != "" && req.Name != server.Name {
		updates["name"] = req.Name
	}
	if req.UserID != "" {
		var user models.User
		var found bool
		if newUserID, err := uuid.Parse(req.UserID); err == nil {
			if database.DB.Where("id = ?", newUserID).First(&user).Error == nil {
				found = true
			}
		}
		if !found && len(req.UserID) >= 8 {
			var users []models.User
			database.DB.Select("id").Find(&users)
			for _, u := range users {
				if strings.HasPrefix(u.ID.String(), req.UserID) {
					database.DB.Where("id = ?", u.ID).First(&user)
					found = true
					break
				}
			}
		}
		if !found {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "User not found"})
		}
		if user.ID != server.UserID {
			updates["user_id"] = user.ID
		}
	}
	if req.Memory > 0 && req.Memory != server.Memory {
		updates["memory"] = req.Memory
	}
	if req.CPU > 0 && req.CPU != server.CPU {
		updates["cpu"] = req.CPU
	}
	if req.Disk > 0 && req.Disk != server.Disk {
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
		database.DB.Model(&server).Updates(updates)
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	database.DB.Preload("User").Preload("Node").Preload("Package").Where("id = ?", serverID).First(&server)

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerResources, "Updated server: "+server.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"server_id": serverID, "updates": updates})

	return c.JSON(fiber.Map{"success": true, "data": server})
}

func AdminTransferServer(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var req struct {
		TargetNodeID string `json:"target_node_id"`
	}
	if err := c.BodyParser(&req); err != nil || req.TargetNodeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "target_node_id required"})
	}

	targetNodeID, err := uuid.Parse(req.TargetNodeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid target node ID"})
	}

	var server models.Server
	database.DB.Where("id = ?", serverID).First(&server)

	mixinInput := map[string]interface{}{
		"server_id":      serverID.String(),
		"name":           server.Name,
		"target_node_id": targetNodeID.String(),
	}

	result, err := plugins.ExecuteMixin(string(plugins.MixinServerTransfer), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return services.StartTransfer(serverID, targetNodeID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	transferID := result.(string)

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminServerTransfer, "Started server transfer", c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"server_id": serverID, "target_node_id": targetNodeID, "transfer_id": transferID})

	plugins.Emit(plugins.EventServerTransferred, map[string]string{"server_id": serverID.String(), "target_node_id": targetNodeID.String(), "transfer_id": transferID})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"transfer_id": transferID}})
}

func AdminGetTransferStatus(c *fiber.Ctx) error {
	transferID := c.Params("transferId")
	status := services.GetTransferStatus(transferID)
	if status == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Transfer not found"})
	}
	return c.JSON(fiber.Map{"success": true, "data": status})
}

func AdminGetAllTransfers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "data": services.GetAllTransfers()})
}
