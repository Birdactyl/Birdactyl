package api

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"cauthon-axis/internal/logger"
	"cauthon-axis/internal/server"

	"github.com/gofiber/fiber/v2"
)

func handleCreateArchive(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Transfer("Creating archive for server %s", id)

	archivePath, err := server.ArchiveServer(id)
	if err != nil {
		logger.Error("Failed to create archive for %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	info, _ := os.Stat(archivePath)
	logger.Success("Archive created for %s: %d bytes", id, info.Size())
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"size": info.Size(),
		},
	})
}

func handleDownloadArchive(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Transfer("Download archive requested for server %s", id)

	path, err := server.GetArchivePath(id)
	if err != nil {
		logger.Error("Archive not found for %s: %v", id, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	c.Set("Content-Disposition", "attachment; filename=\""+id+"-transfer.tar.gz\"")
	c.Set("Content-Type", "application/gzip")
	return c.SendFile(path)
}

func handleDeleteArchive(c *fiber.Ctx) error {
	id := c.Params("id")
	logger.Transfer("Deleting archive for server %s", id)

	if err := server.DeleteArchive(id); err != nil {
		logger.Error("Failed to delete archive for %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Success("Archive deleted for %s", id)
	return c.JSON(fiber.Map{"success": true})
}

func handleImportServer(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if err := c.BodyParser(&req); err != nil || req.URL == "" {
		file, err := c.FormFile("archive")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false, "error": "url or archive file required",
			})
		}
		return handleImportFromFile(c, id, file)
	}

	return handleImportFromURL(c, id, req.URL, req.Token)
}

func handleImportFromURL(c *fiber.Ctx, id, url, token string) error {
	logger.Transfer("Import from URL started for server %s", id)
	logger.Transfer("Fetching archive from %s", url)
	logger.Transfer("Using token: %s...", token[:min(len(token), 10)])

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to fetch archive for %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "failed to fetch archive: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Error("Source returned status %d for %s", resp.StatusCode, id)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "source node returned status " + resp.Status,
		})
	}

	logger.Transfer("Downloading archive for %s: %d bytes", id, resp.ContentLength)

	tmpPath := filepath.Join(os.TempDir(), id+"-import.tar.gz")
	dst, err := os.Create(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	written, err := io.Copy(dst, resp.Body)
	dst.Close()
	if err != nil {
		os.Remove(tmpPath)
		logger.Error("Failed to download archive for %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	logger.Transfer("Downloaded %d bytes for %s", written, id)

	logger.Transfer("Extracting archive for %s", id)
	if err := server.ImportServer(id, tmpPath); err != nil {
		os.Remove(tmpPath)
		logger.Error("Failed to import server %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	os.Remove(tmpPath)
	logger.Success("Import complete for server %s", id)
	return c.JSON(fiber.Map{"success": true})
}

func handleImportFromFile(c *fiber.Ctx, id string, file *multipart.FileHeader) error {
	logger.Transfer("Import from file started for server %s", id)
	logger.Transfer("Received archive: %d bytes", file.Size)

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	defer src.Close()

	tmpPath := filepath.Join(os.TempDir(), id+"-import.tar.gz")
	dst, err := os.Create(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	dst.Close()

	logger.Transfer("Extracting archive for %s", id)
	if err := server.ImportServer(id, tmpPath); err != nil {
		os.Remove(tmpPath)
		logger.Error("Failed to import server %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}

	os.Remove(tmpPath)
	logger.Success("Import complete for server %s", id)
	return c.JSON(fiber.Map{"success": true})
}
