package api

import (
	"os"
	"path/filepath"

	"cauthon-axis/internal/server"

	"github.com/gofiber/fiber/v2"
)

func handleListFiles(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.Query("path", "/")

	files, err := server.ListFiles(id, path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": files})
}

func handleListFilesWithHashes(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.Query("path", "/")

	files, err := server.ListFilesWithHashes(id, path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": files})
}

func handleReadFile(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.Query("path")

	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "error": "path required",
		})
	}

	content, err := server.ReadFile(id, path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": string(content)})
}

func handleSearchFiles(c *fiber.Ctx) error {
	id := c.Params("id")
	query := c.Query("q")

	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "query required"})
	}

	results, err := server.SearchFiles(id, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": results})
}

func handleCreateFolder(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Path string `json:"path"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}
	if err := server.CreateFolder(id, body.Path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleWriteFile(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}
	if err := server.WriteFile(id, body.Path, []byte(body.Content)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleUploadFile(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.FormValue("path", "/")

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "file required"})
	}

	if file.Size > 100*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "file too large (max 100MB)"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	defer src.Close()

	filePath := path
	if path == "/" || path == "" {
		filePath = "/" + file.Filename
	} else {
		filePath = path + "/" + file.Filename
	}

	if err := server.WriteFileStream(id, filePath, src); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleDeletePath(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.Query("path")

	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}

	if err := server.DeletePath(id, path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleMovePath(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.BodyParser(&body); err != nil || body.From == "" || body.To == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "from and to required"})
	}

	if err := server.MovePath(id, body.From, body.To); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleCopyPath(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.BodyParser(&body); err != nil || body.From == "" || body.To == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "from and to required"})
	}

	if err := server.CopyPath(id, body.From, body.To); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleCompressPath(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Path   string `json:"path"`
		Dest   string `json:"dest"`
		Format string `json:"format"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" || body.Dest == "" || body.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path, dest and format required"})
	}

	if err := server.CompressPath(id, body.Path, body.Dest, body.Format); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleDecompressPath(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Path string `json:"path"`
		Dest string `json:"dest"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" || body.Dest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path and dest required"})
	}

	if err := server.DecompressPath(id, body.Path, body.Dest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleDownloadFile(c *fiber.Ctx) error {
	id := c.Params("id")
	path := c.Query("path")

	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}

	filePath, err := server.GetFilePath(id, path)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "file not found"})
	}
	if info.IsDir() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "cannot download directory"})
	}

	fileName := filepath.Base(path)
	c.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Set("Content-Type", "application/octet-stream")

	return c.SendFile(filePath)
}

func handleBulkDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Paths []string `json:"paths"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths required"})
	}

	deleted, err := server.BulkDelete(id, body.Paths)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"deleted": deleted}})
}

func handleBulkCopy(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Paths []string `json:"paths"`
		Dest  string   `json:"dest"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 || body.Dest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths and dest required"})
	}

	copied, err := server.BulkCopy(id, body.Paths, body.Dest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"copied": copied}})
}

func handleBulkCompress(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Paths  []string `json:"paths"`
		Dest   string   `json:"dest"`
		Format string   `json:"format"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 || body.Dest == "" || body.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths, dest and format required"})
	}

	if err := server.BulkCompress(id, body.Paths, body.Dest, body.Format); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleDownloadURL(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		URL  string `json:"url"`
		Path string `json:"path"`
	}
	if err := c.BodyParser(&body); err != nil || body.URL == "" || body.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "url and path required"})
	}

	if err := server.DownloadURL(id, body.URL, body.Path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func handleInstallModpack(c *fiber.Ctx) error {
	id := c.Params("id")
	var body server.ModpackInstallRequest
	if err := c.BodyParser(&body); err != nil || body.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "url required"})
	}

	result, err := server.InstallModpack(id, body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": result})
}
