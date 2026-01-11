package server

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func getServerWithDBPerm(c *fiber.Ctx, perm string) (*models.Server, error) {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
		return nil, errHandled
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
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

func GetServerDatabases(c *fiber.Ctx) error {
	server, err := getServerWithDBPerm(c, models.PermDatabaseView)
	if err != nil {
		return nil
	}

	databases, err := services.GetServerDatabases(server.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	result := make([]fiber.Map, len(databases))
	for i, db := range databases {
		result[i] = fiber.Map{
			"id":            db.ID,
			"database_name": db.DatabaseName,
			"username":      db.Username,
			"password":      db.Password,
			"host":          db.Host.Host,
			"port":          db.Host.Port,
			"created_at":    db.CreatedAt,
		}
	}

	plugins.ExecuteMixin(string(plugins.MixinDatabaseList), map[string]interface{}{"server_id": server.ID.String(), "databases": result}, func(input map[string]interface{}) (interface{}, error) {
		return input["databases"], nil
	})

	return c.JSON(fiber.Map{"success": true, "data": result})
}

func CreateServerDatabase(c *fiber.Ctx) error {
	server, err := getServerWithDBPerm(c, models.PermDatabaseCreate)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)

	var req struct {
		HostID string `json:"host_id"`
		Name   string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.Name == "" {
		req.Name = "default"
	}

	hostID, err := uuid.Parse(req.HostID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid host ID"})
	}

	if allow, msg := plugins.Emit(plugins.EventDatabaseCreating, map[string]string{"server_id": server.ID.String(), "host_id": req.HostID, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"host_id":   hostID.String(),
		"name":      req.Name,
		"user_id":   user.ID.String(),
	}

	var db *models.ServerDatabase
	_, err = plugins.ExecuteMixin(string(plugins.MixinDatabaseCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var createErr error
		db, createErr = services.CreateServerDatabase(server.ID, hostID, req.Name)
		return db, createErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	var host models.DatabaseHost
	database.DB.First(&host, "id = ?", db.HostID)

	handlers.Log(c, user, handlers.ActionDatabaseCreate, "Created database "+db.DatabaseName, map[string]interface{}{"server_id": server.ID, "database_id": db.ID, "database_name": db.DatabaseName})
	plugins.Emit(plugins.EventDatabaseCreated, map[string]string{"server_id": server.ID.String(), "database_id": db.ID.String(), "database_name": db.DatabaseName})

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":            db.ID,
			"database_name": db.DatabaseName,
			"username":      db.Username,
			"password":      db.Password,
			"host":          host.Host,
			"port":          host.Port,
			"created_at":    db.CreatedAt,
		},
	})
}

func DeleteServerDatabase(c *fiber.Ctx) error {
	server, err := getServerWithDBPerm(c, models.PermDatabaseDelete)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)

	dbID, err := uuid.Parse(c.Params("dbId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid database ID"})
	}

	var db models.ServerDatabase
	database.DB.First(&db, "id = ?", dbID)

	if allow, msg := plugins.Emit(plugins.EventDatabaseDeleting, map[string]string{"server_id": server.ID.String(), "database_id": dbID.String(), "database_name": db.DatabaseName}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id":     server.ID.String(),
		"database_id":   dbID.String(),
		"database_name": db.DatabaseName,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinDatabaseDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.DeleteServerDatabase(dbID)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionDatabaseDelete, "Deleted database "+db.DatabaseName, map[string]interface{}{"server_id": server.ID, "database_id": dbID, "database_name": db.DatabaseName})
	plugins.Emit(plugins.EventDatabaseDeleted, map[string]string{"server_id": server.ID.String(), "database_id": dbID.String(), "database_name": db.DatabaseName})

	return c.JSON(fiber.Map{"success": true})
}

func RotateDatabasePassword(c *fiber.Ctx) error {
	server, err := getServerWithDBPerm(c, models.PermDatabaseUpdate)
	if err != nil {
		return nil
	}
	user := c.Locals("user").(*models.User)

	dbID, err := uuid.Parse(c.Params("dbId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid database ID"})
	}

	db, err := services.RotateDatabasePassword(dbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	var host models.DatabaseHost
	database.DB.First(&host, "id = ?", db.HostID)

	handlers.Log(c, user, handlers.ActionDatabaseRotatePassword, "Rotated password for database "+db.DatabaseName, map[string]interface{}{"server_id": server.ID, "database_id": dbID, "database_name": db.DatabaseName})

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":            db.ID,
			"database_name": db.DatabaseName,
			"username":      db.Username,
			"password":      db.Password,
			"host":          host.Host,
			"port":          host.Port,
		},
	})
}

func GetDatabaseHosts(c *fiber.Ctx) error {
	_, err := getServerWithDBPerm(c, models.PermDatabaseCreate)
	if err != nil {
		return nil
	}

	var hosts []models.DatabaseHost
	database.DB.Find(&hosts)

	result := make([]fiber.Map, len(hosts))
	for i, h := range hosts {
		var count int64
		database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", h.ID).Count(&count)

		result[i] = fiber.Map{
			"id":            h.ID,
			"name":          h.Name,
			"host":          h.Host,
			"port":          h.Port,
			"max_databases": h.MaxDatabases,
			"used":          count,
		}
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}
