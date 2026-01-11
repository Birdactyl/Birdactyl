package handlers

import (
	"encoding/json"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetMyPermissions(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	perms, err := services.GetUserServerPermissions(user.ID, serverID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "No access"})
	}

	return c.JSON(fiber.Map{"success": true, "data": perms})
}

func GetSubusers(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	if server.UserID != user.ID && !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	subusers, err := services.GetSubusers(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to fetch subusers"})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinSubuserList), map[string]interface{}{"server_id": serverID.String(), "subusers": subusers}, func(input map[string]interface{}) (interface{}, error) {
		return input["subusers"], nil
	})
	if result != nil {
		subusers = result.([]models.Subuser)
	}

	return c.JSON(fiber.Map{"success": true, "data": subusers})
}

func AddSubuser(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	if server.UserID != user.ID && !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	var req struct {
		Email       string   `json:"email"`
		Permissions []string `json:"permissions"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Email is required"})
	}

	var targetUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	if targetUser.ID == server.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Cannot add server owner as subuser"})
	}

	var existing models.Subuser
	if database.DB.Where("server_id = ? AND user_id = ?", serverID, targetUser.ID).First(&existing).Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "User is already a subuser"})
	}

	if allow, msg := plugins.Emit(plugins.EventSubuserAdding, map[string]string{"server_id": serverID.String(), "user_id": targetUser.ID.String(), "email": req.Email}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id":   serverID.String(),
		"user_id":     targetUser.ID.String(),
		"email":       req.Email,
		"permissions": req.Permissions,
	}

	var subuser *models.Subuser
	_, err = plugins.ExecuteMixin(string(plugins.MixinSubuserAdd), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var addErr error
		subuser, addErr = services.AddSubuser(serverID, targetUser.ID, req.Permissions)
		return subuser, addErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to add subuser"})
	}

	database.DB.Preload("User").First(subuser, subuser.ID)

	Log(c, user, ActionSubuserAdd, "Added subuser to server", map[string]interface{}{"server_id": serverID, "subuser_email": req.Email})
	plugins.Emit(plugins.EventSubuserAdded, map[string]string{"server_id": serverID.String(), "subuser_id": subuser.ID.String(), "user_id": targetUser.ID.String()})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": subuser})
}

func UpdateSubuser(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	subuserID, err := uuid.Parse(c.Params("subuserId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid subuser ID"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	if server.UserID != user.ID && !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	var subuser models.Subuser
	if err := database.DB.Where("id = ? AND server_id = ?", subuserID, serverID).First(&subuser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Subuser not found"})
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	mixinInput := map[string]interface{}{
		"server_id":   serverID.String(),
		"subuser_id":  subuserID.String(),
		"user_id":     subuser.UserID.String(),
		"permissions": req.Permissions,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinSubuserUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		subuser.Permissions, _ = json.Marshal(req.Permissions)
		return nil, database.DB.Save(&subuser).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	database.DB.Preload("User").First(&subuser, subuser.ID)

	Log(c, user, ActionSubuserUpdate, "Updated subuser permissions", map[string]interface{}{"server_id": serverID, "subuser_id": subuserID})

	return c.JSON(fiber.Map{"success": true, "data": subuser})
}

func RemoveSubuser(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	subuserID, err := uuid.Parse(c.Params("subuserId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid subuser ID"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	if server.UserID != user.ID && !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	var subuser models.Subuser
	if err := database.DB.Preload("User").Where("id = ? AND server_id = ?", subuserID, serverID).First(&subuser).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Subuser not found"})
	}

	if allow, msg := plugins.Emit(plugins.EventSubuserRemoving, map[string]string{"server_id": serverID.String(), "subuser_id": subuserID.String(), "user_id": subuser.UserID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id":  serverID.String(),
		"subuser_id": subuserID.String(),
		"user_id":    subuser.UserID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinSubuserRemove), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Delete(&subuser).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	Log(c, user, ActionSubuserRemove, "Removed subuser from server", map[string]interface{}{"server_id": serverID, "subuser_id": subuserID})
	plugins.Emit(plugins.EventSubuserRemoved, map[string]string{"server_id": serverID.String(), "subuser_id": subuserID.String()})

	return c.JSON(fiber.Map{"success": true, "message": "Subuser removed"})
}
