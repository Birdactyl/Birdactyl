package server

import (
	"bufio"
	"bytes"
	"io"
	"net/url"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func getServerForFiles(c *fiber.Ctx) (*models.Server, error) {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}
	server, err := services.GetServerByID(serverID, user.ID, user.IsAdmin)
	if err != nil {
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}
	return server, nil
}

func getServerWithFilePerm(c *fiber.Ctx, perm string) (*models.Server, error) {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
		return nil, errHandled
	}

	var server models.Server
	if err := database.DB.Preload("Node").Preload("Package").Where("id = ?", serverID).First(&server).Error; err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
		return nil, errHandled
	}

	if user.IsAdmin || server.UserID == user.ID {
		return &server, nil
	}

	if !services.HasServerPermission(user.ID, serverID, false, perm) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Permission denied"})
		return nil, errHandled
	}

	return &server, nil
}

func proxyGetWithQuery(c *fiber.Ctx, server *models.Server, endpoint, queryKey, queryVal string) error {
	resp, err := services.ProxyToNode(server, "GET", "/api/servers/"+server.ID.String()+endpoint+"?"+queryKey+"="+url.QueryEscape(queryVal), nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func proxyPost(c *fiber.Ctx, server *models.Server, endpoint string, body interface{}) error {
	resp, err := services.ProxyToNode(server, "POST", "/api/servers/"+server.ID.String()+endpoint, body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func ListFiles(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}

	plugins.ExecuteMixin(string(plugins.MixinFileList), map[string]interface{}{"server_id": server.ID.String(), "path": c.Query("path", "/")}, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	return proxyGetWithQuery(c, server, "/files", "path", c.Query("path", "/"))
}

func ReadFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileRead)
	if err != nil {
		return nil
	}
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}

	user := c.Locals("user").(*models.User)
	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      path,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileRead), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	return proxyGetWithQuery(c, server, "/files/read", "path", path)
}

func SearchFiles(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileList)
	if err != nil {
		return nil
	}
	q := c.Query("q")
	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "query required"})
	}
	return proxyGetWithQuery(c, server, "/files/search", "q", q)
}

func CreateFolder(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileCreate)
	if err != nil {
		return nil
	}
	var body struct{ Path string `json:"path"` }
	if err := c.BodyParser(&body); err != nil || body.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}
	user := c.Locals("user").(*models.User)
	handlers.Log(c, user, handlers.ActionFileCreateFolder, "Created folder", map[string]interface{}{"server_id": server.ID, "path": body.Path})
	return proxyPost(c, server, "/files/folder", body)
}

func WriteFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileWrite)
	if err != nil {
		return nil
	}
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path and content required"})
	}
	user := c.Locals("user").(*models.User)
	if allow, msg := plugins.Emit(plugins.EventFileWriting, map[string]string{"server_id": server.ID.String(), "path": body.Path, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      body.Path,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileWrite), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileWrite, "Wrote file", map[string]interface{}{"server_id": server.ID, "path": body.Path})
	plugins.Emit(plugins.EventFileWritten, map[string]string{"server_id": server.ID.String(), "path": body.Path})
	return proxyPost(c, server, "/files/write", body)
}

func UploadFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileUpload)
	if err != nil {
		return nil
	}
	path := c.FormValue("path", "/")
	user := c.Locals("user").(*models.User)
	if allow, msg := plugins.Emit(plugins.EventFileUploading, map[string]string{"server_id": server.ID.String(), "path": path, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      path,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileUpload), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileUpload, "Uploaded file", map[string]interface{}{"server_id": server.ID, "path": path})
	resp, err := services.ProxyUploadToNode(server, "/api/servers/"+server.ID.String()+"/files/upload?path="+path, bytes.NewReader(c.Body()), c.Get("Content-Type"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	plugins.Emit(plugins.EventFileUploaded, map[string]string{"server_id": server.ID.String(), "path": path})
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func DeleteFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileDelete)
	if err != nil {
		return nil
	}
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}
	user := c.Locals("user").(*models.User)
	if allow, msg := plugins.Emit(plugins.EventFileDeleting, map[string]string{"server_id": server.ID.String(), "path": path, "user_id": user.ID.String()}); !allow {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": msg})
	}

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      path,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileDelete, "Deleted file", map[string]interface{}{"server_id": server.ID, "path": path})
	resp, err := services.ProxyToNode(server, "DELETE", "/api/servers/"+server.ID.String()+"/files?path="+url.QueryEscape(path), nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	plugins.Emit(plugins.EventFileDeleted, map[string]string{"server_id": server.ID.String(), "path": path})
	return c.Status(resp.StatusCode).Send(resp.Body)
}

func MoveFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileMove)
	if err != nil {
		return nil
	}
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.BodyParser(&body); err != nil || body.From == "" || body.To == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "from and to required"})
	}
	user := c.Locals("user").(*models.User)

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"from":      body.From,
		"to":        body.To,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileMove), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileMove, "Moved file", map[string]interface{}{"server_id": server.ID, "from": body.From, "to": body.To})
	return proxyPost(c, server, "/files/move", body)
}

func CopyFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileCopy)
	if err != nil {
		return nil
	}
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.BodyParser(&body); err != nil || body.From == "" || body.To == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "from and to required"})
	}
	user := c.Locals("user").(*models.User)

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"from":      body.From,
		"to":        body.To,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileCopy), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileCopy, "Copied file", map[string]interface{}{"server_id": server.ID, "from": body.From, "to": body.To})
	return proxyPost(c, server, "/files/copy", body)
}

func CompressFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileCompress)
	if err != nil {
		return nil
	}
	var body struct {
		Path   string `json:"path"`
		Dest   string `json:"dest"`
		Format string `json:"format"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" || body.Dest == "" || body.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path, dest and format required"})
	}
	user := c.Locals("user").(*models.User)

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      body.Path,
		"dest":      body.Dest,
		"format":    body.Format,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileCompress), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileCompress, "Compressed file", map[string]interface{}{"server_id": server.ID, "path": body.Path, "dest": body.Dest})
	return proxyPost(c, server, "/files/compress", body)
}

func DecompressFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileDecompress)
	if err != nil {
		return nil
	}
	var body struct {
		Path string `json:"path"`
		Dest string `json:"dest"`
	}
	if err := c.BodyParser(&body); err != nil || body.Path == "" || body.Dest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path and dest required"})
	}
	user := c.Locals("user").(*models.User)

	mixinInput := map[string]interface{}{
		"server_id": server.ID.String(),
		"path":      body.Path,
		"dest":      body.Dest,
		"user_id":   user.ID.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinFileDecompress), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	handlers.Log(c, user, handlers.ActionFileDecompress, "Decompressed file", map[string]interface{}{"server_id": server.ID, "path": body.Path, "dest": body.Dest})
	return proxyPost(c, server, "/files/decompress", body)
}

func DownloadFile(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileDownload)
	if err != nil {
		return nil
	}
	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "path required"})
	}
	resp, err := services.StreamDownloadFromNode(server, "/api/servers/"+server.ID.String()+"/files/download?path="+url.QueryEscape(path))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"success": false, "error": "download failed"})
	}
	c.Set("Content-Disposition", resp.Header.Get("Content-Disposition"))
	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		c.Set("Content-Length", cl)
	}
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		io.Copy(w, resp.Body)
	})
	return nil
}

func BulkDeleteFiles(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileDelete)
	if err != nil {
		return nil
	}
	var body struct{ Paths []string `json:"paths"` }
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths required"})
	}
	user := c.Locals("user").(*models.User)
	handlers.Log(c, user, handlers.ActionFileBulkDelete, "Bulk deleted files", map[string]interface{}{"server_id": server.ID, "count": len(body.Paths)})
	return proxyPost(c, server, "/files/bulk-delete", body)
}

func BulkCopyFiles(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileCopy)
	if err != nil {
		return nil
	}
	var body struct {
		Paths []string `json:"paths"`
		Dest  string   `json:"dest"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 || body.Dest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths and dest required"})
	}
	user := c.Locals("user").(*models.User)
	handlers.Log(c, user, handlers.ActionFileBulkCopy, "Bulk copied files", map[string]interface{}{"server_id": server.ID, "count": len(body.Paths), "dest": body.Dest})
	return proxyPost(c, server, "/files/bulk-copy", body)
}

func BulkCompressFiles(c *fiber.Ctx) error {
	server, err := getServerWithFilePerm(c, models.PermFileCompress)
	if err != nil {
		return nil
	}
	var body struct {
		Paths  []string `json:"paths"`
		Dest   string   `json:"dest"`
		Format string   `json:"format"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Paths) == 0 || body.Dest == "" || body.Format == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "paths, dest and format required"})
	}
	user := c.Locals("user").(*models.User)
	handlers.Log(c, user, handlers.ActionFileBulkCompress, "Bulk compressed files", map[string]interface{}{"server_id": server.ID, "count": len(body.Paths), "dest": body.Dest})
	return proxyPost(c, server, "/files/bulk-compress", body)
}
