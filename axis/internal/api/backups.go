package api

import (
	"cauthon-axis/internal/server"

	"github.com/gofiber/fiber/v2"
)

func handleListBackups(c *fiber.Ctx) error {
	id := c.Params("id")
	backups, err := server.ListBackups(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": backups})
}

func handleCreateBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Name string `json:"name"`
	}
	c.BodyParser(&body)

	backup, err := server.CreateBackup(id, body.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": backup})
}

func handleDeleteBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	backupID := c.Params("backupId")

	if err := server.DeleteBackup(id, backupID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Backup deleted"})
}

func handleDownloadBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	backupID := c.Params("backupId")

	path, err := server.GetBackupPath(id, backupID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	c.Set("Content-Disposition", "attachment; filename=\""+backupID+".tar.gz\"")
	return c.SendFile(path)
}

func handleRestoreBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	backupID := c.Params("backupId")

	status, _ := server.GetStatus(id)
	if status == "running" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false, "error": "Server must be stopped before restoring a backup",
		})
	}

	if err := server.RestoreBackup(id, backupID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Backup restored"})
}
