package admin

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/logger"
	"birdactyl-panel-backend/internal/models"
	"math"

	"github.com/gofiber/fiber/v2"
)

func AdminGetMounts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	search := c.Query("search", "")

	var mounts []models.Mount
	var total int64

	q := database.DB.Model(&models.Mount{}).Preload("Servers").Preload("Nodes").Preload("Packages")

	if search != "" {
		q = q.Where(database.ILike("name", "?"), database.ILikeValue(search)).
			Or(database.ILike("source", "?"), database.ILikeValue(search)).
			Or(database.ILike("target", "?"), database.ILikeValue(search))
	}

	if err := q.Count(&total).Error; err != nil {
		logger.Error("Failed to count mounts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load mounts"})
	}

	offset := (page - 1) * perPage
	if err := q.Offset(offset).Limit(perPage).Find(&mounts).Error; err != nil {
		logger.Error("Failed to load mounts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load mounts"})
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"mounts":      mounts,
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func AdminCreateMount(c *fiber.Ctx) error {
	var req struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Source        string   `json:"source"`
		Target        string   `json:"target"`
		ReadOnly      bool     `json:"read_only"`
		UserMountable bool     `json:"user_mountable"`
		Navigable     bool     `json:"navigable"`
		Nodes         []string `json:"nodes"`
		Packages      []string `json:"packages"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Source == "" || req.Target == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, Source, and Target are required"})
	}

	mount := models.Mount{
		Name:          req.Name,
		Description:   req.Description,
		Source:        req.Source,
		Target:        req.Target,
		ReadOnly:      req.ReadOnly,
		UserMountable: req.UserMountable,
		Navigable:     req.Navigable,
	}

	if len(req.Nodes) > 0 {
		database.DB.Where("id IN ?", req.Nodes).Find(&mount.Nodes)
	}
	if len(req.Packages) > 0 {
		database.DB.Where("id IN ?", req.Packages).Find(&mount.Packages)
	}

	if err := database.DB.Create(&mount).Error; err != nil {
		logger.Error("Failed to create mount: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create mount"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, "admin.mount.create", "Created mount: "+mount.Name, c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mount,
	})
}

func AdminUpdateMount(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name          *string  `json:"name"`
		Description   *string  `json:"description"`
		Source        *string  `json:"source"`
		Target        *string  `json:"target"`
		ReadOnly      *bool    `json:"read_only"`
		UserMountable *bool    `json:"user_mountable"`
		Navigable     *bool    `json:"navigable"`
		Nodes         []string `json:"nodes"`
		Packages      []string `json:"packages"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var mount models.Mount
	if err := database.DB.First(&mount, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mount not found"})
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Source != nil {
		updates["source"] = *req.Source
	}
	if req.Target != nil {
		updates["target"] = *req.Target
	}
	if req.ReadOnly != nil {
		updates["read_only"] = *req.ReadOnly
	}
	if req.UserMountable != nil {
		updates["user_mountable"] = *req.UserMountable
	}
	if req.Navigable != nil {
		updates["navigable"] = *req.Navigable
	}

	if req.Nodes != nil {
		var nodes []models.Node
		if len(req.Nodes) > 0 {
			database.DB.Where("id IN ?", req.Nodes).Find(&nodes)
		}
		database.DB.Model(&mount).Association("Nodes").Replace(&nodes)
	}
	if req.Packages != nil {
		var packages []models.Package
		if len(req.Packages) > 0 {
			database.DB.Where("id IN ?", req.Packages).Find(&packages)
		}
		database.DB.Model(&mount).Association("Packages").Replace(&packages)
	}

	if err := database.DB.Model(&mount).Updates(updates).Error; err != nil {
		logger.Error("Failed to update mount: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update mount"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, "admin.mount.update", "Updated mount: "+mount.Name, c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mount,
	})
}

func AdminDeleteMount(c *fiber.Ctx) error {
	id := c.Params("id")

	var mount models.Mount
	if err := database.DB.Preload("Servers").First(&mount, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mount not found"})
	}

	if err := database.DB.Model(&mount).Association("Servers").Clear(); err != nil {
		logger.Error("Failed to clear mount associations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete mount"})
	}

	if err := database.DB.Delete(&mount).Error; err != nil {
		logger.Error("Failed to delete mount: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete mount"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, "admin.mount.delete", "Deleted mount: "+mount.Name, c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func AdminAttachMount(c *fiber.Ctx) error {
	mountId := c.Params("id")
	serverId := c.Params("serverId")

	var mount models.Mount
	if err := database.DB.First(&mount, "id = ?", mountId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mount not found"})
	}

	var server models.Server
	if err := database.DB.First(&server, "id = ?", serverId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Server not found"})
	}

	if err := database.DB.Model(&mount).Association("Servers").Append(&server); err != nil {
		logger.Error("Failed to attach mount: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to attach mount"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, "admin.mount.attach", "Attached mount "+mount.Name+" to server "+server.Name, c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func AdminDetachMount(c *fiber.Ctx) error {
	mountId := c.Params("id")
	serverId := c.Params("serverId")

	var mount models.Mount
	if err := database.DB.First(&mount, "id = ?", mountId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mount not found"})
	}

	var server models.Server
	if err := database.DB.First(&server, "id = ?", serverId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Server not found"})
	}

	if err := database.DB.Model(&mount).Association("Servers").Delete(&server); err != nil {
		logger.Error("Failed to detach mount: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to detach mount"})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, "admin.mount.detach", "Detached mount "+mount.Name+" from server "+server.Name, c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{
		"success": true,
	})
}
