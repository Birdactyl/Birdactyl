package handlers

import (
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateNodeRequest struct {
	Name string `json:"name"`
	FQDN string `json:"fqdn"`
	Port int    `json:"port"`
	Icon string `json:"icon"`
}

type UpdateNodeRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type PairNodeRequest struct {
	Name string `json:"name"`
	FQDN string `json:"fqdn"`
	Port int    `json:"port"`
	Code string `json:"code"`
	Icon string `json:"icon"`
}

func AdminGetNodes(c *fiber.Ctx) error {
	nodes, err := services.GetNodes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinNodeList), map[string]interface{}{"nodes": nodes}, func(input map[string]interface{}) (interface{}, error) {
		return input["nodes"], nil
	})
	if result != nil {
		nodes = result.([]models.Node)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nodes,
	})
}

func AdminRefreshNodes(c *fiber.Ctx) error {
	nodes, err := services.RefreshNodes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nodes,
	})
}

func AdminCreateNode(c *fiber.Ctx) error {
	var req CreateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Name == "" || req.FQDN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name and FQDN are required",
		})
	}

	if req.Port == 0 {
		req.Port = 8443
	}

	mixinInput := map[string]interface{}{
		"name": req.Name,
		"fqdn": req.FQDN,
		"port": req.Port,
	}

	var node *models.Node
	var nodeToken *services.NodeToken

	_, err := plugins.ExecuteMixin(string(plugins.MixinNodeCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		var createErr error
		node, nodeToken, createErr = services.CreateNode(req.Name, req.FQDN, req.Port)
		return node, createErr
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		status := fiber.StatusInternalServerError
		if err == services.ErrNodeNameTaken {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminNodeCreate, "Created node: "+req.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"node_name": req.Name, "fqdn": req.FQDN})

	plugins.Emit(plugins.EventNodeCreated, map[string]string{"node_id": node.ID.String(), "name": node.Name, "fqdn": node.FQDN})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"node":  node,
			"token": nodeToken,
		},
	})
}

func AdminGetNode(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid node ID",
		})
	}

	node, err := services.GetNodeByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	result, _ := plugins.ExecuteMixin(string(plugins.MixinNodeGet), map[string]interface{}{"node_id": id.String(), "node": node}, func(input map[string]interface{}) (interface{}, error) {
		return input["node"], nil
	})
	if result != nil {
		node = result.(*models.Node)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    node,
	})
}

func AdminUpdateNode(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid node ID",
		})
	}

	var req UpdateNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	node, err := services.UpdateNode(id, req.Name, req.Icon)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminNodeUpdate, "Updated node: "+node.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"node_id": id.String()})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    node,
	})
}

func AdminDeleteNode(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid node ID",
		})
	}

	mixinInput := map[string]interface{}{
		"node_id": id.String(),
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinNodeDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, services.DeleteNode(id)
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
	LogActivity(admin.ID, admin.Username, ActionAdminNodeDelete, "Deleted node", c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"node_id": id.String()})

	plugins.Emit(plugins.EventNodeDeleted, map[string]string{"node_id": id.String()})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Node deleted",
	})
}

func AdminResetNodeToken(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid node ID",
		})
	}

	token, err := services.ResetNodeToken(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminNodeResetToken, "Reset node token", c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"node_id": id.String()})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    token,
	})
}

type NodeHeartbeatRequest struct {
	System    models.SystemInfo `json:"system"`
	DisplayIP string            `json:"display_ip"`
}

func NodeHeartbeat(c *fiber.Ctx) error {
	node := c.Locals("node").(*models.Node)

	var req NodeHeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := services.NodeHeartbeat(node.ID, req.System, req.DisplayIP); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func GetAvailableNodes(c *fiber.Ctx) error {
	nodes, err := services.GetOnlineNodes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "error": "Failed to fetch nodes",
		})
	}
	return c.JSON(fiber.Map{"success": true, "data": nodes})
}

func AdminGeneratePairingCode(c *fiber.Ctx) error {
	code := services.GeneratePairingCode()
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"code": code,
		},
	})
}

func AdminPairNode(c *fiber.Ctx) error {
	var req PairNodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Name == "" || req.FQDN == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name, FQDN, and code are required",
		})
	}

	if req.Port == 0 {
		req.Port = 8443
	}

	panelURL := c.Protocol() + "://" + c.Hostname()

	node, nodeToken, err := services.PairWithNode(req.Name, req.FQDN, req.Port, panelURL, req.Code)
	if err != nil {
		status := fiber.StatusInternalServerError
		msg := err.Error()

		switch err {
		case services.ErrNodeNameTaken:
			status = fiber.StatusConflict
		case services.ErrNodeNotReady:
			status = fiber.StatusServiceUnavailable
			msg = "Node not ready for pairing. Run 'axis pair' on the node first."
		case services.ErrPairingRejected:
			status = fiber.StatusForbidden
			msg = "Pairing was rejected on the node"
		case services.ErrPairingTimeout:
			status = fiber.StatusRequestTimeout
			msg = "Pairing request timed out waiting for confirmation"
		}

		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   msg,
		})
	}

	admin := c.Locals("user").(*models.User)
	LogActivity(admin.ID, admin.Username, ActionAdminNodeCreate, "Paired node: "+req.Name, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"node_name": req.Name, "fqdn": req.FQDN, "method": "pairing"})

	plugins.Emit(plugins.EventNodeCreated, map[string]string{"node_id": node.ID.String(), "name": node.Name, "fqdn": node.FQDN})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"node":  node,
			"token": nodeToken,
		},
	})
}
