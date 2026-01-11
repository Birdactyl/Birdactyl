package auth

import (
	"regexp"

	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/handlers/admin"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func Register(c *fiber.Ctx) error {
	if admin.IsIPBanned(c.IP()) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	if !services.IsRegistrationEnabled() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Registration is currently disabled"})
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email, username and password are required",
		})
	}

	if !emailRegex.MatchString(req.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email format",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Password must be at least 8 characters",
		})
	}

	if allow, msg := plugins.Emit(plugins.EventUserRegistering, map[string]string{"email": req.Email, "username": req.Username, "ip": c.IP()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
		"ip":       c.IP(),
	}

	var user *models.User
	var tokens *services.TokenPair

	_, err := plugins.ExecuteMixin(string(plugins.MixinUserCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var regErr error
		user, tokens, regErr = services.Register(
			req.Email,
			req.Username,
			req.Password,
			c.IP(),
			c.Get("User-Agent"),
		)
		return user, regErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			resp := fiber.Map{"success": false, "error": mixinErr.Message}
			if len(mixinErr.Notifications) > 0 {
				resp["notifications"] = mixinErr.Notifications
			}
			return c.Status(fiber.StatusForbidden).JSON(resp)
		}
		status := fiber.StatusInternalServerError
		switch err {
		case services.ErrEmailTaken, services.ErrUsernameTaken:
			status = fiber.StatusConflict
		case services.ErrIPLimitReached:
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	handlers.LogActivity(user.ID, user.Username, handlers.ActionAuthRegister, "User registered", c.IP(), c.Get("User-Agent"), false, nil)
	plugins.Emit(plugins.EventUserRegistered, map[string]string{"user_id": user.ID.String(), "username": user.Username, "email": user.Email})

	resp := fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":   user,
			"tokens": tokens,
		},
	}
	if notifications := plugins.CollectNotifications(); len(notifications) > 0 {
		resp["notifications"] = notifications
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func Login(c *fiber.Ctx) error {
	if admin.IsIPBanned(c.IP()) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email and password are required",
		})
	}

	if allow, msg := plugins.Emit(plugins.EventUserLoggingIn, map[string]string{"email": req.Email, "ip": c.IP()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"email": req.Email,
		"ip":    c.IP(),
	}

	var user *models.User
	var tokens *services.TokenPair

	_, err := plugins.ExecuteMixin(string(plugins.MixinUserAuthenticate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var loginErr error
		user, tokens, loginErr = services.Login(
			req.Email,
			req.Password,
			c.IP(),
			c.Get("User-Agent"),
		)
		return user, loginErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		status := fiber.StatusUnauthorized
		if err == services.ErrUserBanned {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	handlers.LogActivity(user.ID, user.Username, handlers.ActionAuthLogin, "User logged in", c.IP(), c.Get("User-Agent"), user.IsAdmin, nil)
	plugins.Emit(plugins.EventUserLoggedIn, map[string]string{"user_id": user.ID.String(), "username": user.Username, "ip": c.IP()})

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user":   user,
			"tokens": tokens,
		},
	})
}

func Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Refresh token is required",
		})
	}

	tokens, err := services.RefreshTokens(
		req.RefreshToken,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    tokens,
	})
}

func Logout(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)
	services.Logout(claims.SessionID)

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.ActionAuthLogout, "User logged out", nil)
		plugins.Emit(plugins.EventUserLogout, map[string]string{"user_id": user.ID.String(), "username": user.Username})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

func LogoutAll(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*services.AccessClaims)
	services.LogoutAll(claims.UserID)

	user, _ := services.GetUserByID(claims.UserID)
	if user != nil {
		handlers.Log(c, user, handlers.ActionAuthLogoutAll, "User logged out all sessions", nil)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All sessions terminated",
	})
}
