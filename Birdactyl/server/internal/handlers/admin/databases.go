package admin

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AdminGetDatabaseHosts(c *fiber.Ctx) error {
	var hosts []models.DatabaseHost
	database.DB.Find(&hosts)

	result := make([]fiber.Map, len(hosts))
	for i, h := range hosts {
		var count int64
		database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", h.ID).Count(&count)
		result[i] = fiber.Map{
			"id":              h.ID,
			"name":            h.Name,
			"host":            h.Host,
			"port":            h.Port,
			"username":        h.Username,
			"max_databases":   h.MaxDatabases,
			"databases_count": count,
			"created_at":      h.CreatedAt,
		}
	}

	plugins.ExecuteMixin(string(plugins.MixinDBHostList), map[string]interface{}{"hosts": result}, func(input map[string]interface{}) (interface{}, error) {
		return input["hosts"], nil
	})

	return c.JSON(fiber.Map{"success": true, "data": result})
}

func AdminCreateDatabaseHost(c *fiber.Ctx) error {
	var req struct {
		Name         string `json:"name"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		MaxDatabases int    `json:"max_databases"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.Name == "" || req.Host == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Name, host, username and password are required"})
	}

	if req.Port == 0 {
		req.Port = 3306
	}
	if req.MaxDatabases == 0 {
		req.MaxDatabases = 100
	}

	mixinInput := map[string]interface{}{
		"name": req.Name,
		"host": req.Host,
		"port": req.Port,
	}

	var host *models.DatabaseHost
	_, err := plugins.ExecuteMixin(string(plugins.MixinDBHostCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var createErr error
		host, createErr = services.CreateDatabaseHost(req.Name, req.Host, req.Port, req.Username, req.Password, req.MaxDatabases)
		return host, createErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminDBHostCreate, "Created database host: "+req.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"host_name": req.Name})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": host})
}

func AdminUpdateDatabaseHost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid host ID"})
	}

	var host models.DatabaseHost
	if err := database.DB.Where("id = ?", id).First(&host).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Host not found"})
	}

	var req struct {
		Name         string `json:"name"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		MaxDatabases int    `json:"max_databases"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Host != "" {
		updates["host"] = req.Host
	}
	if req.Port > 0 {
		updates["port"] = req.Port
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.MaxDatabases > 0 {
		updates["max_databases"] = req.MaxDatabases
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No changes provided"})
	}

	mixinInput := map[string]interface{}{
		"host_id": id.String(),
		"name":    host.Name,
		"updates": updates,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinDBHostUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Model(&host).Updates(updates).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminDBHostUpdate, "Updated database host: "+host.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"host_id": id})

	return c.JSON(fiber.Map{"success": true, "message": "Host updated"})
}

func AdminDeleteDatabaseHost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid host ID"})
	}

	var host models.DatabaseHost
	if err := database.DB.Where("id = ?", id).First(&host).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Host not found"})
	}

	var dbCount int64
	database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", id).Count(&dbCount)
	if dbCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Cannot delete host with existing databases"})
	}

	mixinInput := map[string]interface{}{
		"host_id": id.String(),
		"name":    host.Name,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinDBHostDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Delete(&host).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminDBHostDelete, "Deleted database host: "+host.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"host_id": id})

	return c.JSON(fiber.Map{"success": true, "message": "Host deleted"})
}

func AdminGetHostDatabases(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid host ID"})
	}

	var databases []models.ServerDatabase
	database.DB.Preload("Server").Where("host_id = ?", id).Find(&databases)

	return c.JSON(fiber.Map{"success": true, "data": databases})
}

func AdminDeleteDatabase(c *fiber.Ctx) error {
	dbID, err := uuid.Parse(c.Params("dbId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid database ID"})
	}

	var db models.ServerDatabase
	if err := database.DB.First(&db, "id = ?", dbID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Database not found"})
	}

	if err := services.DeleteServerDatabase(dbID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminDbDelete, "Deleted database: "+db.DatabaseName, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"database_id": dbID})

	return c.JSON(fiber.Map{"success": true, "message": "Database deleted"})
}
