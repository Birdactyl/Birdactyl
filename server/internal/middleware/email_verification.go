package middleware

import (
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

var routeActionMap = map[string]map[string]string{
	"POST": {
		"/api/v1/servers":                     "server.create",
		"/api/v1/servers/*/start":             "server.start",
		"/api/v1/servers/*/stop":              "server.stop",
		"/api/v1/servers/*/kill":              "server.kill",
		"/api/v1/servers/*/restart":           "server.restart",
		"/api/v1/servers/*/reinstall":         "server.reinstall",
		"/api/v1/servers/*/command":           "server.command",
		"/api/v1/servers/*/backups":           "server.backup.create",
		"/api/v1/servers/*/subusers":          "server.subuser.add",
		"/api/v1/servers/*/files/folder":      "server.file.create_folder",
		"/api/v1/servers/*/files/write":       "server.file.write",
		"/api/v1/servers/*/files/upload":      "server.file.upload",
		"/api/v1/servers/*/files/move":        "server.file.move",
		"/api/v1/servers/*/files/copy":        "server.file.copy",
		"/api/v1/servers/*/files/compress":    "server.file.compress",
		"/api/v1/servers/*/files/decompress":  "server.file.decompress",
		"/api/v1/servers/*/files/bulk-delete": "server.file.bulk_delete",
		"/api/v1/servers/*/files/bulk-copy":   "server.file.bulk_copy",
		"/api/v1/servers/*/files/bulk-compress": "server.file.bulk_compress",
		"/api/v1/servers/*/databases":         "server.database.create",
		"/api/v1/servers/*/sftp/password":     "server.sftp.password_reset",
		"/api/v1/auth/profile":                "profile.update",
		"/api/v1/auth/2fa/setup":              "profile.2fa_setup",
		"/api/v1/auth/2fa/enable":             "profile.2fa_enable",
		"/api/v1/auth/2fa/disable":            "profile.2fa_disable",
		"/api/v1/auth/api-keys":               "profile.api_keys",
	},
	"PATCH": {
		"/api/v1/auth/profile":          "profile.update",
		"/api/v1/auth/password":         "profile.password_change",
		"/api/v1/servers/*/name":        "server.name.update",
		"/api/v1/servers/*/resources":   "server.resources.update",
		"/api/v1/servers/*/variables":   "server.variables.update",
		"/api/v1/servers/*/subusers/*":  "server.subuser.update",
	},
	"PUT": {
		"/api/v1/servers/*/allocations/primary": "server.allocation.set_primary",
	},
	"DELETE": {
		"/api/v1/servers/*":                       "server.delete",
		"/api/v1/servers/*/backups/*":              "server.backup.delete",
		"/api/v1/servers/*/subusers/*":             "server.subuser.remove",
		"/api/v1/servers/*/files":                  "server.file.delete",
		"/api/v1/servers/*/allocations":            "server.allocation.delete",
		"/api/v1/servers/*/databases/*":            "server.database.delete",
	},
}

func matchRoute(pattern, path string) bool {
	pi, pj := 0, 0
	pp := []byte(pattern)
	pa := []byte(path)
	for pi < len(pp) && pj < len(pa) {
		if pp[pi] == '*' {
			pi++
			for pj < len(pa) && pa[pj] != '/' {
				pj++
			}
		} else if pp[pi] == pa[pj] {
			pi++
			pj++
		} else {
			return false
		}
	}
	return pi == len(pp) && pj == len(pa)
}

func RequireEmailVerification() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Next()
		}

		if user.EmailVerified || user.IsAdmin {
			return c.Next()
		}

		if !services.IsEmailVerificationEnabled() {
			return c.Next()
		}

		methodMap, exists := routeActionMap[c.Method()]
		if !exists {
			return c.Next()
		}

		path := c.Path()
		for pattern, action := range methodMap {
			if matchRoute(pattern, path) {
				if err := services.CheckEmailVerification(user, action); err != nil {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"success": false,
						"error":   "Email verification required to perform this action",
						"code":    "EMAIL_NOT_VERIFIED",
					})
				}
				break
			}
		}

		return c.Next()
	}
}
