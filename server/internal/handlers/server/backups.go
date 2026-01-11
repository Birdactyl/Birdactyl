package server

import (
	"errors"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var errBackupHandled = errors.New("handled")

func checkBackupPerm(c *fiber.Ctx, perm string) (*models.Server, error) {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
		return nil, errBackupHandled
	}

	var server models.Server
	if err := database.DB.Preload("Node").Where("id = ?", serverID).First(&server).Error; err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
		return nil, errBackupHandled
	}

	if user.IsAdmin || server.UserID == user.ID {
		return &server, nil
	}

	if !services.HasServerPermission(user.ID, serverID, false, perm) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Permission denied"})
		return nil, errBackupHandled
	}

	return &server, nil
}

func ListBackups(c *fiber.Ctx) error {
	server, err := checkBackupPerm(c, models.PermBackupList)
	if err != nil {
		return nil
	}

	data, err := services.ProxyGetToNode(server, "/backups")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	plugins.ExecuteMixin(string(plugins.MixinBackupList), map[string]interface{}{"server_id": server.ID.String(), "data": data}, func(input map[string]interface{}) (interface{}, error) {
		return input["data"], nil
	})

	return c.JSON(data)
}

func CreateBackup(c *fiber.Ctx) error {
	server, err := checkBackupPerm(c, models.PermBackupCreate)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)

	if allow, msg := plugins.Emit(plugins.EventBackupCreating, map[string]string{"server_id": server.ID.String(), "server_name": server.Name, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id":   server.ID.String(),
		"server_name": server.Name,
		"user_id":     user.ID.String(),
	}

	var data interface{}
	_, err = plugins.ExecuteMixin(string(plugins.MixinBackupCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var proxyErr error
		data, proxyErr = services.ProxyPostToNode(server, "/backups", c.Body())
		return data, proxyErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionBackupCreate, "Created backup for server: "+server.Name, map[string]interface{}{"server_id": server.ID})
	plugins.Emit(plugins.EventBackupCreated, map[string]string{"server_id": server.ID.String(), "server_name": server.Name})

	return c.Status(fiber.StatusCreated).JSON(data)
}

func DeleteBackup(c *fiber.Ctx) error {
	server, err := checkBackupPerm(c, models.PermBackupDelete)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)
	backupID := c.Params("backupId")

	if allow, msg := plugins.Emit(plugins.EventBackupDeleting, map[string]string{"server_id": server.ID.String(), "backup_id": backupID}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"backup_id": backupID,
	}

	var data interface{}
	_, err = plugins.ExecuteMixin(string(plugins.MixinBackupDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var proxyErr error
		data, proxyErr = services.ProxyDeleteToNode(server, "/backups/"+backupID)
		return data, proxyErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionBackupDelete, "Deleted backup", map[string]interface{}{"server_id": server.ID, "backup_id": backupID})
	plugins.Emit(plugins.EventBackupDeleted, map[string]string{"server_id": server.ID.String(), "backup_id": backupID})

	return c.JSON(data)
}

func DownloadBackup(c *fiber.Ctx) error {
	server, err := checkBackupPerm(c, models.PermBackupDownload)
	if err != nil {
		return nil
	}

	backupID := c.Params("backupId")
	url, err := services.GetNodeProxyURL(server, "/backups/"+backupID+"/download")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Redirect(url)
}

func RestoreBackup(c *fiber.Ctx) error {
	server, err := checkBackupPerm(c, models.PermBackupRestore)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)
	backupID := c.Params("backupId")

	if allow, msg := plugins.Emit(plugins.EventBackupRestoring, map[string]string{"server_id": server.ID.String(), "backup_id": backupID}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"backup_id": backupID,
	}

	var data interface{}
	_, err = plugins.ExecuteMixin(string(plugins.MixinBackupRestore), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var proxyErr error
		data, proxyErr = services.ProxyPostToNode(server, "/backups/"+backupID+"/restore", nil)
		return data, proxyErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionBackupRestore, "Restored backup", map[string]interface{}{"server_id": server.ID, "backup_id": backupID})
	plugins.Emit(plugins.EventBackupRestored, map[string]string{"server_id": server.ID.String(), "backup_id": backupID})

	return c.JSON(data)
}
