package plugins

import (
	"context"
	"time"

	"birdactyl-panel-backend/internal/models"
	pb "birdactyl-panel-backend/internal/plugins/proto"

	"github.com/gofiber/fiber/v2"
)

type Route struct {
	PluginID string
	Method   string
	Path     string
}

var routes []Route

func RegisterPluginRoutes(app *fiber.App) {
	app.All("/api/v1/plugins/:pluginId/*", handlePluginRoute)
}

func handlePluginRoute(c *fiber.Ctx) error {
	pluginID := c.Params("pluginId")

	path := c.Params("*")
	if path == "" {
		path = "/"
	} else {
		path = "/" + path
	}

	if pluginID == "ui" && path == "/manifests" {
		return handleUIManifests(c)
	}

	if path == "/ui/bundle.js" {
		return servePluginBundle(c, pluginID)
	}

	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	query := make(map[string]string)
	c.Request().URI().QueryArgs().VisitAll(func(k, v []byte) {
		query[string(k)] = string(v)
	})

	userID := ""
	if user, ok := c.Locals("user").(*models.User); ok && user != nil {
		userID = user.ID.String()
	}

	req := &pb.HTTPRequest{
		Method:  c.Method(),
		Path:    path,
		Headers: headers,
		Query:   query,
		Body:    c.Body(),
		UserId:  userID,
	}

	if ps := GetStreamRegistry().Get(pluginID); ps != nil {
		resp, err := ps.SendHTTP(req)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "plugin error: " + err.Error()})
		}
		if resp == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "route not found"})
		}
		for k, v := range resp.Headers {
			c.Set(k, v)
		}
		return c.Status(int(resp.Status)).Send(resp.Body)
	}

	plugin := GetRegistry().Get(pluginID)
	if plugin == nil || !plugin.Online {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "plugin not found"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := plugin.Client.OnHTTP(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "plugin error: " + err.Error()})
	}

	for k, v := range resp.Headers {
		c.Set(k, v)
	}

	return c.Status(int(resp.Status)).Send(resp.Body)
}


func servePluginBundle(c *fiber.Ctx, pluginID string) error {
	manifest := GetUIRegistry().Get(pluginID)
	if manifest == nil || !manifest.HasBundle {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "plugin bundle not found",
		})
	}

	if len(manifest.BundleData) > 0 {
		c.Set("Content-Type", "application/javascript")
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Send(manifest.BundleData)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   "bundle not embedded",
	})
}

func handleUIManifests(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	manifests := GetUIRegistry().All()
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"plugins": manifests,
		},
	})
}
