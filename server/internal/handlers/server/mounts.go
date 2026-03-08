package server

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ServerMountResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	ReadOnly    bool      `json:"read_only"`
	IsMounted   bool      `json:"is_mounted"`
	Navigable   bool      `json:"navigable"`
}

func GetServerMounts(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermMountRead)
	if err != nil {
		return nil
	}

	var allMounts []models.Mount
	err = database.DB.Preload("Nodes").Preload("Packages").Where("user_mountable = ?", true).Find(&allMounts).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to load mounts"})
	}

	var serverWithMounts models.Server
	database.DB.Preload("Mounts").First(&serverWithMounts, serverID)

	var response []ServerMountResponse
	for _, m := range allMounts {
		var nodeMatch bool
		for _, n := range m.Nodes {
			if n.ID == server.NodeID {
				nodeMatch = true
				break
			}
		}
		if !nodeMatch {
			continue
		}

		packageMatch := len(m.Packages) == 0
		if !packageMatch {
			for _, p := range m.Packages {
				if p.ID == server.PackageID {
					packageMatch = true
					break
				}
			}
		}

		if !packageMatch {
			continue
		}

		isMounted := false
		for _, sm := range serverWithMounts.Mounts {
			if sm.ID == m.ID {
				isMounted = true
				break
			}
		}

		response = append(response, ServerMountResponse{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			Source:      m.Source,
			Target:      m.Target,
			ReadOnly:    m.ReadOnly,
			IsMounted:   isMounted,
			Navigable:   m.Navigable,
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": response})
}

func MountServerMount(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}
	mountID, err := uuid.Parse(c.Params("mountId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid mount ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermMountUpdate)
	if err != nil {
		return nil
	}

	var m models.Mount
	err = database.DB.Preload("Nodes").Preload("Packages").Where("id = ? AND user_mountable = ?", mountID, true).First(&m).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Mount not found or not user mountable"})
	}

	var nodeMatch bool
	for _, n := range m.Nodes {
		if n.ID == server.NodeID {
			nodeMatch = true
			break
		}
	}
	if !nodeMatch {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Mount not available for this node"})
	}

	packageMatch := len(m.Packages) == 0
	if !packageMatch {
		for _, p := range m.Packages {
			if p.ID == server.PackageID {
				packageMatch = true
				break
			}
		}
	}

	if !packageMatch {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Mount not available for this package"})
	}

	err = database.DB.Model(&server).Association("Mounts").Append(&m)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to attach mount"})
	}

	handlers.Log(c, user, "server.mount.add", "Attached mount: "+m.Name, map[string]interface{}{"server_id": serverID, "mount_id": mountID})

	return c.JSON(fiber.Map{"success": true, "message": "Mount attached successfully"})
}

func UnmountServerMount(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}
	mountID, err := uuid.Parse(c.Params("mountId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid mount ID"})
	}

	server, err := checkServerPerm(c, serverID, models.PermMountUpdate)
	if err != nil {
		return nil
	}

	var m models.Mount
	err = database.DB.Where("id = ? AND user_mountable = ?", mountID, true).First(&m).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Mount not found or not user mountable"})
	}

	err = database.DB.Model(&server).Association("Mounts").Delete(&m)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to detach mount"})
	}

	handlers.Log(c, user, "server.mount.remove", "Detached mount: "+m.Name, map[string]interface{}{"server_id": serverID, "mount_id": mountID})

	return c.JSON(fiber.Map{"success": true, "message": "Mount detached successfully"})
}
