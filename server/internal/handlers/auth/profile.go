package auth

import (
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func Me(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	response := fiber.Map{
		"success": true,
		"data":    user,
	}

	if newTokens, ok := c.Locals("new_tokens").(*services.TokenPair); ok && newTokens != nil {
		response["tokens"] = newTokens
	}

	return c.JSON(response)
}

func GetResources(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	cfg := config.Get()
	usage := services.GetUserResourceUsage(user.ID)

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

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"enabled": cfg.Resources.Enabled,
			"limits": fiber.Map{
				"ram":     ramLimit,
				"cpu":     cpuLimit,
				"disk":    diskLimit,
				"servers": serverLimit,
			},
			"used": fiber.Map{
				"ram":     usage.RAM,
				"cpu":     usage.CPU,
				"disk":    usage.Disk,
				"servers": usage.Servers,
			},
		},
	})
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func UpdateProfile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	user, err := services.UpdateProfile(claims.UserID, req.Username, req.Email)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err == services.ErrEmailTaken || err == services.ErrUsernameTaken {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.Log(c, user, handlers.ActionProfileUpdate, "Updated profile", map[string]interface{}{"username": req.Username, "email": req.Email})

	return c.JSON(fiber.Map{"success": true, "data": user})
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func UpdatePassword(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Password must be at least 8 characters"})
	}

	if err := services.UpdatePassword(claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.ActionProfilePasswordChange, "Changed password", nil)
	}

	return c.JSON(fiber.Map{"success": true, "message": "Password updated"})
}

func GetSessions(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)
	sessions := services.GetUserSessions(claims.UserID, claims.SessionID)
	return c.JSON(fiber.Map{"success": true, "data": sessions})
}

func RevokeSession(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)
	sessionID := c.Params("id")

	if err := services.RevokeSession(claims.UserID, sessionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.ActionProfileSessionRevoke, "Revoked session", map[string]interface{}{"session_id": sessionID})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Session revoked"})
}

func RevokeAllSessions(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)
	services.RevokeOtherSessions(claims.UserID, claims.SessionID)

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.ActionProfileSessionsRevoke, "Revoked all other sessions", nil)
	}

	return c.JSON(fiber.Map{"success": true, "message": "All other sessions revoked"})
}
