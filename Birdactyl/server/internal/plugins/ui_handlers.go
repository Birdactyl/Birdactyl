package plugins

import (
	"os"

	"birdactyl-panel-backend/internal/config"

	"github.com/gofiber/fiber/v2"
)

func RegisterUIRoutes(app *fiber.App) {
	app.Get("/api/v1/plugins/ui/manifests", handleGetManifests)
	app.Get("/api/v1/plugins/:pluginId/ui/bundle.js", handleGetBundle)
}

func handleGetManifests(c *fiber.Ctx) error {
	manifests := GetUIRegistry().All()
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"plugins": manifests,
		},
	})
}

func handleGetBundle(c *fiber.Ctx) error {
	pluginID := c.Params("pluginId")

	manifest := GetUIRegistry().Get(pluginID)
	if manifest == nil || !manifest.HasBundle {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "plugin bundle not found",
		})
	}

	if len(manifest.BundleData) > 0 {
		c.Set("Content-Type", "application/javascript")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.Send(manifest.BundleData)
	}

	cfg := config.Get()
	bundlePath := GetPluginBundlePath(pluginID, cfg.Plugins.Directory)

	if bundlePath == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "bundle file not found",
		})
	}

	content, err := os.ReadFile(bundlePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to read bundle",
		})
	}

	c.Set("Content-Type", "application/javascript")
	c.Set("Cache-Control", "public, max-age=3600")
	return c.Send(content)
}
