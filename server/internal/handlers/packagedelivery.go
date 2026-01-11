package handlers

import (
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CreatePackageRequest struct {
	Name                string                     `json:"name"`
	Version             string                     `json:"version"`
	Author              string                     `json:"author"`
	Description         string                     `json:"description"`
	Icon                string                     `json:"icon"`
	DockerImage         string                     `json:"docker_image"`
	InstallImage        string                     `json:"install_image"`
	Startup             string                     `json:"startup"`
	InstallScript       string                     `json:"install_script"`
	StopSignal          string                     `json:"stop_signal"`
	StopCommand         string                     `json:"stop_command"`
	StopTimeout         int                        `json:"stop_timeout"`
	StartupEditable     bool                       `json:"startup_editable"`
	DockerImageEditable bool                       `json:"docker_image_editable"`
	Ports               []models.PackagePort       `json:"ports"`
	Variables           []models.PackageVariable   `json:"variables"`
	ConfigFiles         []models.PackageConfigFile `json:"config_files"`
	AddonSources        []models.AddonSource       `json:"addon_sources"`
}

func AdminGetPackages(c *fiber.Ctx) error {
	packages, err := services.GetPackages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinPackageList), map[string]interface{}{"packages": packages}, func(input map[string]interface{}) (interface{}, error) {
		return input["packages"], nil
	})
	if result != nil {
		packages = result.([]models.Package)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    packages,
	})
}

func AdminCreatePackage(c *fiber.Ctx) error {
	var req CreatePackageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Name == "" || req.DockerImage == "" || req.Startup == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name, docker_image, and startup are required",
		})
	}

	if req.StopSignal == "" {
		req.StopSignal = "SIGTERM"
	}
	if req.StopTimeout == 0 {
		req.StopTimeout = 30
	}

	mixinInput := map[string]interface{}{
		"name":         req.Name,
		"docker_image": req.DockerImage,
	}

	portsJSON, _ := datatypes.NewJSONType(req.Ports).MarshalJSON()
	varsJSON, _ := datatypes.NewJSONType(req.Variables).MarshalJSON()
	configJSON, _ := datatypes.NewJSONType(req.ConfigFiles).MarshalJSON()
	addonJSON, _ := datatypes.NewJSONType(req.AddonSources).MarshalJSON()

	pkg := &models.Package{
		Name:                req.Name,
		Version:             req.Version,
		Author:              req.Author,
		Description:         req.Description,
		Icon:                req.Icon,
		DockerImage:         req.DockerImage,
		InstallImage:        req.InstallImage,
		Startup:             req.Startup,
		InstallScript:       req.InstallScript,
		StopSignal:          req.StopSignal,
		StopCommand:         req.StopCommand,
		StopTimeout:         req.StopTimeout,
		StartupEditable:     req.StartupEditable,
		DockerImageEditable: req.DockerImageEditable,
		Ports:               portsJSON,
		Variables:           varsJSON,
		ConfigFiles:         configJSON,
		AddonSources:        addonJSON,
	}

	_, err := plugins.ExecuteMixin(string(plugins.MixinPackageCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return pkg, services.CreatePackage(pkg)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		status := fiber.StatusInternalServerError
		if err == services.ErrPackageNameTaken {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminPackageCreate, "Created package: "+req.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"package_name": req.Name})

	plugins.Emit(plugins.EventPackageCreated, map[string]string{"package_id": pkg.ID.String(), "name": pkg.Name})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    pkg,
	})
}

func AdminGetPackage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid package ID",
		})
	}

	pkg, err := services.GetPackageByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinPackageGet), map[string]interface{}{"package_id": id.String(), "package": pkg}, func(input map[string]interface{}) (interface{}, error) {
		return input["package"], nil
	})
	if result != nil {
		pkg = result.(*models.Package)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    pkg,
	})
}

func AdminUpdatePackage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid package ID",
		})
	}

	var req CreatePackageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	portsJSON, _ := datatypes.NewJSONType(req.Ports).MarshalJSON()
	varsJSON, _ := datatypes.NewJSONType(req.Variables).MarshalJSON()
	configJSON, _ := datatypes.NewJSONType(req.ConfigFiles).MarshalJSON()
	addonJSON, _ := datatypes.NewJSONType(req.AddonSources).MarshalJSON()

	updates := map[string]interface{}{
		"name":                  req.Name,
		"version":               req.Version,
		"author":                req.Author,
		"description":           req.Description,
		"icon":                  req.Icon,
		"docker_image":          req.DockerImage,
		"install_image":         req.InstallImage,
		"startup":               req.Startup,
		"install_script":        req.InstallScript,
		"stop_signal":           req.StopSignal,
		"stop_command":          req.StopCommand,
		"stop_timeout":          req.StopTimeout,
		"startup_editable":      req.StartupEditable,
		"docker_image_editable": req.DockerImageEditable,
		"ports":                 portsJSON,
		"variables":             varsJSON,
		"config_files":          configJSON,
		"addon_sources":         addonJSON,
	}

	mixinInput := map[string]interface{}{
		"package_id": id.String(),
		"name":       req.Name,
	}

	var pkg *models.Package
	_, err = plugins.ExecuteMixin(string(plugins.MixinPackageUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var updateErr error
		pkg, updateErr = services.UpdatePackage(id, updates)
		return pkg, updateErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminPackageUpdate, "Updated package: "+req.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"package_id": id})

	plugins.Emit(plugins.EventPackageUpdated, map[string]string{"package_id": id.String(), "name": req.Name})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    pkg,
	})
}

func AdminDeletePackage(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid package ID",
		})
	}

	mixinInput := map[string]interface{}{
		"package_id": id.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinPackageDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.DeletePackage(id)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminPackageDelete, "Deleted package", c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"package_id": id})

	plugins.Emit(plugins.EventPackageDeleted, map[string]string{"package_id": id.String()})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Package deleted",
	})
}

func GetAvailablePackages(c *fiber.Ctx) error {
	packages, err := services.GetPackages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    packages,
	})
}
