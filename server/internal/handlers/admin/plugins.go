package admin

import (
	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/plugins"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ListPluginsPublic(c *fiber.Ctx) error {
	result := make([]fiber.Map, 0)

	for _, p := range plugins.GetRegistry().All() {
		if !p.Online {
			continue
		}
		item := fiber.Map{
			"id":      p.Config.ID,
			"name":    p.Config.Name,
			"version": "",
		}
		if p.Info != nil {
			item["version"] = p.Info.Version
			if p.Info.Name != "" {
				item["name"] = p.Info.Name
			}
		}
		result = append(result, item)
	}

	for _, ps := range plugins.GetStreamRegistry().All() {
		item := fiber.Map{
			"id":      ps.ID,
			"name":    ps.Info.Name,
			"version": ps.Info.Version,
		}
		if ps.Info.Ui != nil {
			sidebarItems := make([]fiber.Map, 0)
			for _, s := range ps.Info.Ui.SidebarItems {
				sidebarItems = append(sidebarItems, fiber.Map{
					"id":      s.Id,
					"label":   s.Label,
					"icon":    s.Icon,
					"href":    s.Href,
					"section": s.Section,
					"order":   s.Order,
				})
			}
			item["sidebarItems"] = sidebarItems

			pages := make([]fiber.Map, 0)
			for _, p := range ps.Info.Ui.Pages {
				pages = append(pages, fiber.Map{
					"path":      p.Path,
					"component": p.Component,
					"title":     p.Title,
					"icon":      p.Icon,
				})
			}
			item["pages"] = pages
		}
		result = append(result, item)
	}

	return c.JSON(result)
}

func AdminListPlugins(c *fiber.Ctx) error {
	result := make([]fiber.Map, 0)

	for _, p := range plugins.GetRegistry().All() {
		result = append(result, fiber.Map{
			"id":      p.Config.ID,
			"name":    p.Config.Name,
			"address": p.Config.Address,
			"online":  p.Online,
			"mode":    "legacy",
		})
	}

	for _, ps := range plugins.GetStreamRegistry().All() {
		result = append(result, fiber.Map{
			"id":      ps.ID,
			"name":    ps.Info.Name,
			"address": "stream",
			"online":  true,
			"mode":    "stream",
		})
	}

	return c.JSON(fiber.Map{"success": true, "plugins": result})
}

func AdminLoadPlugin(c *fiber.Ctx) error {
	var req plugins.PluginConfig
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request"})
	}
	if err := plugins.LoadPlugin(req); err != nil {
		status := fiber.StatusInternalServerError
		if err == plugins.ErrDynamicDisabled {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func AdminUnloadPlugin(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "missing plugin id"})
	}
	if err := plugins.UnloadPlugin(id); err != nil {
		status := fiber.StatusInternalServerError
		if err == plugins.ErrDynamicDisabled {
			status = fiber.StatusForbidden
		} else if err == plugins.ErrPluginNotFound {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func AdminReloadPlugin(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "missing plugin id"})
	}
	cfg := config.Get()
	if !cfg.Plugins.AllowDynamic {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": plugins.ErrDynamicDisabled.Error()})
	}
	plugin := plugins.GetRegistry().Get(id)
	if plugin == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "plugin not found"})
	}
	pluginCfg := plugin.Config
	if err := plugins.UnloadPlugin(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	if err := plugins.LoadPlugin(pluginCfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func AdminGetPluginConfig(c *fiber.Ctx) error {
	cfg := config.Get()

	_, mvnErr := exec.LookPath("mvn")
	_, goErr := exec.LookPath("go")

	return c.JSON(fiber.Map{
		"success": true,
		"config": fiber.Map{
			"load_mode":     cfg.Plugins.LoadMode,
			"allow_dynamic": cfg.Plugins.AllowDynamic,
			"address":       cfg.Plugins.Address,
			"directory":     cfg.Plugins.Directory,
		},
		"data": fiber.Map{
			"maven_available": mvnErr == nil,
			"go_available":    goErr == nil,
		},
	})
}

func AdminListPluginFiles(c *fiber.Ctx) error {
	cfg := config.Get()
	pluginsDir := cfg.Plugins.Directory
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"files": []fiber.Map{}}})
	}

	cacheDir := filepath.Join(pluginsDir, ".cache")
	files := make([]fiber.Map, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		isJar := strings.HasSuffix(name, ".jar")
		isGoBinary := !strings.Contains(name, ".") && name != ".cache"
		if !isJar && !isGoBinary {
			continue
		}
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		file := fiber.Map{"name": name, "size": size, "type": "java"}
		if isGoBinary {
			file["type"] = "go"
		}
		baseName := strings.TrimSuffix(name, ".jar")
		metaFile := filepath.Join(cacheDir, baseName+".json")
		if metaData, err := os.ReadFile(metaFile); err == nil {
			var meta map[string]string
			if json.Unmarshal(metaData, &meta) == nil {
				file["repo"] = meta["repo"]
				file["owner_name"] = meta["owner_name"]
				file["owner_avatar"] = meta["owner_avatar"]
				file["description"] = meta["description"]
			}
		}
		files = append(files, file)
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"files": files}})
}

func AdminDeletePluginFile(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid filename"})
	}
	cfg := config.Get()
	pluginsDir := cfg.Plugins.Directory
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}
	path := filepath.Join(pluginsDir, filename)

	absPath, _ := filepath.Abs(path)
	for _, p := range plugins.GetRegistry().All() {
		pAbs, _ := filepath.Abs(p.Config.Binary)
		if pAbs == absPath || filepath.Base(p.Config.Binary) == filename {
			plugins.UnloadPlugin(p.Config.ID)
			break
		}
	}

	if err := os.Remove(path); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to delete file"})
	}
	cacheDir := filepath.Join(pluginsDir, ".cache")
	metaFile := filepath.Join(cacheDir, strings.TrimSuffix(filename, ".jar")+".json")
	os.Remove(metaFile)
	return c.JSON(fiber.Map{"success": true})
}

func AdminInstallPluginFromSource(c *fiber.Ctx) error {
	var req struct {
		Repo        string `json:"repo"`
		OwnerName   string `json:"owner_name"`
		OwnerAvatar string `json:"owner_avatar"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil || req.Repo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request"})
	}
	cfg := config.Get()
	pluginsDir := cfg.Plugins.Directory
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}
	tmpDir, err := os.MkdirTemp("", "birdactyl-plugin-*")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to create temp dir"})
	}
	defer os.RemoveAll(tmpDir)

	cloneURL := fmt.Sprintf("https://github.com/%s.git", req.Repo)
	cloneCmd := exec.Command("git", "clone", "--depth", "1", cloneURL, tmpDir)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": fmt.Sprintf("git clone failed: %s", string(out))})
	}

	var outputFile string

	if _, err := os.Stat(filepath.Join(tmpDir, "pom.xml")); err == nil {
		if _, err := exec.LookPath("mvn"); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Maven is not installed on this system"})
		}
		buildCmd := exec.Command("mvn", "clean", "package", "-q", "-DskipTests")
		buildCmd.Dir = tmpDir
		if out, err := buildCmd.CombinedOutput(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": fmt.Sprintf("maven build failed: %s", string(out))})
		}
		targetDir := filepath.Join(tmpDir, "target")
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "no target directory"})
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".jar") && !strings.Contains(e.Name(), "original") {
				outputFile = filepath.Join(targetDir, e.Name())
				break
			}
		}
	} else if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); err == nil {
		if _, err := exec.LookPath("go"); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Go is not installed on this system"})
		}
		repoName := filepath.Base(req.Repo)
		outputPath := filepath.Join(tmpDir, repoName)
		buildCmd := exec.Command("go", "build", "-o", outputPath, ".")
		buildCmd.Dir = tmpDir
		if out, err := buildCmd.CombinedOutput(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": fmt.Sprintf("go build failed: %s", string(out))})
		}
		os.Chmod(outputPath, 0755)
		outputFile = outputPath
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "unknown project type (no pom.xml or go.mod found)"})
	}

	if outputFile == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "no output file found"})
	}

	fileName := filepath.Base(outputFile)
	destPath := filepath.Join(pluginsDir, fileName)
	if err := copyFile(outputFile, destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to copy output"})
	}
	if !strings.HasSuffix(fileName, ".jar") {
		os.Chmod(destPath, 0755)
	}
	savePluginMeta(pluginsDir, fileName, req.Repo, req.OwnerName, req.OwnerAvatar, req.Description)

	plugins.LoadPlugin(plugins.PluginConfig{Binary: destPath})

	return c.JSON(fiber.Map{"success": true, "file": fileName})
}

func AdminUploadPlugin(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "no file provided"})
	}

	cfg := config.Get()
	pluginsDir := cfg.Plugins.Directory
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}

	destPath := filepath.Join(pluginsDir, file.Filename)
	if err := c.SaveFile(file, destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to save file"})
	}

	if !strings.HasSuffix(file.Filename, ".jar") {
		os.Chmod(destPath, 0755)
	}

	description := c.FormValue("description", "")
	savePluginMeta(pluginsDir, file.Filename, "", "Manual Upload", "", description)

	plugins.LoadPlugin(plugins.PluginConfig{Binary: destPath})

	return c.JSON(fiber.Map{"success": true, "file": file.Filename})
}

func AdminInstallPluginFromRelease(c *fiber.Ctx) error {
	var req struct {
		URL         string `json:"url"`
		Filename    string `json:"filename"`
		Repo        string `json:"repo"`
		OwnerName   string `json:"owner_name"`
		OwnerAvatar string `json:"owner_avatar"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil || req.URL == "" || req.Filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request"})
	}
	cfg := config.Get()
	pluginsDir := cfg.Plugins.Directory
	if pluginsDir == "" {
		pluginsDir = "plugins"
	}

	resp, err := http.Get(req.URL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to download"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": fmt.Sprintf("download returned %d", resp.StatusCode)})
	}

	destPath := filepath.Join(pluginsDir, req.Filename)
	out, err := os.Create(destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to create file"})
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(destPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "failed to write file"})
	}

	if !strings.HasSuffix(req.Filename, ".jar") {
		os.Chmod(destPath, 0755)
	}
	savePluginMeta(pluginsDir, req.Filename, req.Repo, req.OwnerName, req.OwnerAvatar, req.Description)

	plugins.LoadPlugin(plugins.PluginConfig{Binary: destPath})

	return c.JSON(fiber.Map{"success": true, "file": req.Filename})
}

func savePluginMeta(pluginsDir, filename, repo, ownerName, ownerAvatar, description string) {
	cacheDir := filepath.Join(pluginsDir, ".cache")
	os.MkdirAll(cacheDir, 0755)
	meta := map[string]string{"repo": repo, "owner_name": ownerName, "owner_avatar": ownerAvatar, "description": description}
	data, _ := json.Marshal(meta)
	metaFile := filepath.Join(cacheDir, strings.TrimSuffix(filename, ".jar")+".json")
	os.WriteFile(metaFile, data, 0644)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
