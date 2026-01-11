package server

import (
	"encoding/json"

	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetServerAllocations(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := services.GetServerByID(serverID, user.ID, user.IsAdmin)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	plugins.ExecuteMixin(string(plugins.MixinAllocationList), map[string]interface{}{"server_id": serverID.String(), "allocations": ports}, func(input map[string]interface{}) (interface{}, error) {
		return input["allocations"], nil
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    ports,
	})
}

func AddAllocation(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"user_id":   user.ID.String(),
	}

	var server *models.Server
	_, err = plugins.ExecuteMixin(string(plugins.MixinAllocationAdd), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var addErr error
		server, addErr = services.AddAllocation(serverID, user.ID, user.IsAdmin)
		return server, addErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionAllocationAdd, "Added allocation", map[string]interface{}{"server_id": serverID})

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": ports})
}

func DeleteAllocation(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var req struct {
		Port int `json:"port"`
	}
	if err := c.BodyParser(&req); err != nil || req.Port == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Port is required"})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"port":      req.Port,
		"user_id":   user.ID.String(),
	}

	var server *models.Server
	_, err = plugins.ExecuteMixin(string(plugins.MixinAllocationDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var delErr error
		server, delErr = services.DeleteAllocation(serverID, user.ID, req.Port, user.IsAdmin)
		return server, delErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionAllocationDelete, "Deleted allocation", map[string]interface{}{"server_id": serverID, "port": req.Port})

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	return c.JSON(fiber.Map{"success": true, "data": ports})
}

func SetPrimaryAllocation(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var req struct {
		Port int `json:"port"`
	}
	if err := c.BodyParser(&req); err != nil || req.Port == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Port is required"})
	}

	mixinInput := map[string]interface{}{
		"server_id": serverID.String(),
		"port":      req.Port,
		"user_id":   user.ID.String(),
	}

	var server *models.Server
	_, err = plugins.ExecuteMixin(string(plugins.MixinAllocationSetPrimary), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var setErr error
		server, setErr = services.SetPrimaryAllocation(serverID, user.ID, req.Port, user.IsAdmin)
		return server, setErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionAllocationSetPrimary, "Set primary allocation", map[string]interface{}{"server_id": serverID, "port": req.Port})

	var ports []models.ServerPort
	json.Unmarshal(server.Ports, &ports)

	return c.JSON(fiber.Map{"success": true, "data": ports})
}
