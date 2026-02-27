package admin

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func IsIPBanned(ip string) bool {
	var count int64
	database.DB.Model(&models.IPBan{}).Where("ip = ?", ip).Count(&count)
	return count > 0
}

func AdminGetIPBans(c *fiber.Ctx) error {
	var bans []models.IPBan
	database.DB.Order("created_at DESC").Find(&bans)

	result, _ := plugins.ExecuteMixin(string(plugins.MixinIPBanList), map[string]interface{}{"bans": bans}, func(input map[string]interface{}) (interface{}, error) {
		return input["bans"], nil
	})
	if result != nil {
		bans = result.([]models.IPBan)
	}

	return c.JSON(fiber.Map{"success": true, "data": bans})
}

func AdminCreateIPBan(c *fiber.Ctx) error {
	var req struct {
		IP     string `json:"ip"`
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil || req.IP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "IP is required"})
	}

	var existing models.IPBan
	if database.DB.Where("ip = ?", req.IP).First(&existing).Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "IP already banned"})
	}

	mixinInput := map[string]interface{}{
		"ip":     req.IP,
		"reason": req.Reason,
	}

	var ban models.IPBan
	_, err := plugins.ExecuteMixin(string(plugins.MixinIPBanCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		ban = models.IPBan{IP: req.IP, Reason: req.Reason}
		return &ban, database.DB.Create(&ban).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminIPBanCreate, "Banned IP: "+req.IP, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"banned_ip": req.IP, "reason": req.Reason})

	plugins.Emit(plugins.EventIPBanCreated, map[string]string{"ip": req.IP, "reason": req.Reason})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": ban})
}

func AdminDeleteIPBan(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid ban ID"})
	}

	var ban models.IPBan
	if err := database.DB.Where("id = ?", id).First(&ban).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Ban not found"})
	}

	mixinInput := map[string]interface{}{
		"ban_id": id.String(),
		"ip":     ban.IP,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinIPBanDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Delete(&ban).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminIPBanDelete, "Unbanned IP: "+ban.IP, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"unbanned_ip": ban.IP})

	plugins.Emit(plugins.EventIPBanDeleted, map[string]string{"ip": ban.IP})

	return c.JSON(fiber.Map{"success": true, "message": "IP ban removed"})
}
