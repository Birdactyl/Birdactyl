package auth

import (
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func TwoFactorSetup(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	setup, err := services.SetupTOTP(claims.UserID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err == services.ErrTOTPAlreadyEnabled {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": setup})
}

func TwoFactorEnable(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Code is required"})
	}

	codes, err := services.EnableTOTP(claims.UserID, req.Code)
	if err != nil {
		status := fiber.StatusBadRequest
		if err == services.ErrTOTPAlreadyEnabled {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.Action2FAEnable, "Enabled two-factor authentication", nil)
	}

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"backup_codes": codes}})
}

func TwoFactorDisable(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Password is required"})
	}

	if err := services.DisableTOTP(claims.UserID, req.Password); err != nil {
		status := fiber.StatusBadRequest
		if err == services.ErrTOTPInvalidPassword {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.Action2FADisable, "Disabled two-factor authentication", nil)
	}

	return c.JSON(fiber.Map{"success": true, "message": "Two-factor authentication disabled"})
}

func TwoFactorBackupCodes(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)

	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Password is required"})
	}

	codes, err := services.RegenerateBackupCodes(claims.UserID, req.Password)
	if err != nil {
		status := fiber.StatusBadRequest
		if err == services.ErrTOTPInvalidPassword {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.Action2FABackupCodes, "Regenerated backup codes", nil)
	}

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"backup_codes": codes}})
}

func TwoFactorVerify(c *fiber.Ctx) error {
	var req struct {
		ChallengeToken string `json:"challenge_token"`
		Code           string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil || req.ChallengeToken == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Challenge token and code are required"})
	}

	user, tokens, err := services.VerifyTwoFactor(req.ChallengeToken, req.Code, c.IP(), c.Get("User-Agent"))
	if err != nil {
		status := fiber.StatusUnauthorized
		if err == services.ErrTOTPInvalidCode {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	handlers.LogActivity(user.ID, user.Username, handlers.ActionAuthLogin, "User logged in (2FA)", c.IP(), c.Get("User-Agent"), user.IsAdmin, nil)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":   user,
			"tokens": tokens,
		},
	})
}
