package admin

import (
	"strconv"

	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func AdminGetSettings(c *fiber.Ctx) error {
	data := fiber.Map{
		"registration_enabled":              services.IsRegistrationEnabled(),
		"server_creation_enabled":           services.IsServerCreationEnabled(),
		"email_verification_enabled":        services.IsEmailVerificationEnabled(),
		"email_verification_restrictions":   services.GetEmailVerificationRestrictions(),
	}

	plugins.ExecuteMixin(string(plugins.MixinSettingsGet), map[string]interface{}{"settings": data}, func(input map[string]interface{}) (interface{}, error) {
		return input["settings"], nil
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func AdminGetRegistrationStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enabled": services.IsRegistrationEnabled(),
		},
	})
}

func AdminSetRegistrationStatus(c *fiber.Ctx) error {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	mixinInput := map[string]interface{}{
		"setting": "registration_enabled",
		"value":   req.Enabled,
	}

	_, err := plugins.ExecuteMixin(string(plugins.MixinSettingsUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SetRegistrationEnabled(req.Enabled)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	admin := c.Locals("user").(*models.User)
	action := "Disabled registration"
	if req.Enabled {
		action = "Enabled registration"
	}
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminSettingsUpdate, action, c.IP(), c.Get("User-Agent"), true, nil)

	plugins.Emit(plugins.EventSettingsUpdated, map[string]string{"setting": "registration_enabled", "value": strconv.FormatBool(req.Enabled)})

	return c.JSON(fiber.Map{"success": true, "message": "Setting updated"})
}

func AdminGetServerCreationStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enabled": services.IsServerCreationEnabled(),
		},
	})
}

func AdminSetServerCreationStatus(c *fiber.Ctx) error {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	mixinInput := map[string]interface{}{
		"setting": "server_creation_enabled",
		"value":   req.Enabled,
	}

	_, err := plugins.ExecuteMixin(string(plugins.MixinSettingsUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.SetServerCreationEnabled(req.Enabled)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	admin := c.Locals("user").(*models.User)
	action := "Disabled server creation"
	if req.Enabled {
		action = "Enabled server creation"
	}
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminSettingsUpdate, action, c.IP(), c.Get("User-Agent"), true, nil)

	plugins.Emit(plugins.EventSettingsUpdated, map[string]string{"setting": "server_creation_enabled", "value": strconv.FormatBool(req.Enabled)})

	return c.JSON(fiber.Map{"success": true, "message": "Setting updated"})
}

func AdminSetRegistrationEnabled(c *fiber.Ctx) error {
	return AdminSetRegistrationStatus(c)
}

func AdminSetServerCreationEnabled(c *fiber.Ctx) error {
	return AdminSetServerCreationStatus(c)
}

func AdminGetEmailVerificationSettings(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enabled":      services.IsEmailVerificationEnabled(),
			"restrictions": services.GetEmailVerificationRestrictions(),
		},
	})
}

func AdminSetEmailVerificationSettings(c *fiber.Ctx) error {
	var req struct {
		Enabled      *bool    `json:"enabled"`
		Restrictions []string `json:"restrictions"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	if req.Enabled != nil {
		services.SetEmailVerificationEnabled(*req.Enabled)
	}

	if req.Restrictions != nil {
		services.SetEmailVerificationRestrictions(req.Restrictions)
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminSettingsUpdate, "Updated email verification settings", c.IP(), c.Get("User-Agent"), true, nil)

	return c.JSON(fiber.Map{"success": true, "message": "Settings updated"})
}
