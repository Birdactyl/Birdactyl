package handlers

import (
	"encoding/json"
	"errors"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var errScheduleHandled = errors.New("handled")

func checkSchedulePerm(c *fiber.Ctx, perm string) (uuid.UUID, error) {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
		return uuid.Nil, errScheduleHandled
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
		return uuid.Nil, errScheduleHandled
	}

	if user.IsAdmin || server.UserID == user.ID {
		return serverID, nil
	}

	if !services.HasServerPermission(user.ID, serverID, false, perm) {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Permission denied"})
		return uuid.Nil, errScheduleHandled
	}

	return serverID, nil
}

func GetServerSchedules(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleList)
	if err != nil {
		return nil
	}

	schedules, err := services.GetSchedulesByServer(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": schedules})
}

func GetSchedule(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleList)
	if err != nil {
		return nil
	}

	scheduleID, err := uuid.Parse(c.Params("scheduleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid schedule ID"})
	}

	schedule, err := services.GetScheduleByID(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Schedule not found"})
	}

	if schedule.ServerID != serverID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	return c.JSON(fiber.Map{"success": true, "data": schedule})
}

type CreateScheduleRequest struct {
	Name           string                `json:"name"`
	CronExpression string                `json:"cron_expression"`
	IsActive       bool                  `json:"is_active"`
	OnlyWhenOnline bool                  `json:"only_when_online"`
	Tasks          []models.ScheduleTask `json:"tasks"`
}

func CreateSchedule(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleCreate)
	if err != nil {
		return nil
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	if req.Name == "" || req.CronExpression == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Name and cron_expression are required"})
	}

	tasksJSON, _ := json.Marshal(req.Tasks)

	schedule := &models.Schedule{
		ServerID:       serverID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		IsActive:       req.IsActive,
		OnlyWhenOnline: req.OnlyWhenOnline,
		Tasks:          tasksJSON,
	}

	if err := services.CreateSchedule(schedule); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": schedule})
}

func UpdateSchedule(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleUpdate)
	if err != nil {
		return nil
	}

	scheduleID, err := uuid.Parse(c.Params("scheduleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid schedule ID"})
	}

	existing, err := services.GetScheduleByID(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Schedule not found"})
	}

	if existing.ServerID != serverID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	tasksJSON, _ := json.Marshal(req.Tasks)

	updates := map[string]interface{}{
		"name":             req.Name,
		"cron_expression":  req.CronExpression,
		"is_active":        req.IsActive,
		"only_when_online": req.OnlyWhenOnline,
		"tasks":            tasksJSON,
	}

	schedule, err := services.UpdateSchedule(scheduleID, updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": schedule})
}

func DeleteSchedule(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleDelete)
	if err != nil {
		return nil
	}

	scheduleID, err := uuid.Parse(c.Params("scheduleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid schedule ID"})
	}

	existing, err := services.GetScheduleByID(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Schedule not found"})
	}

	if existing.ServerID != serverID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	if err := services.DeleteSchedule(scheduleID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Schedule deleted"})
}

func RunScheduleNow(c *fiber.Ctx) error {
	serverID, err := checkSchedulePerm(c, models.PermScheduleRun)
	if err != nil {
		return nil
	}

	scheduleID, err := uuid.Parse(c.Params("scheduleId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid schedule ID"})
	}

	existing, err := services.GetScheduleByID(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Schedule not found"})
	}

	if existing.ServerID != serverID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Access denied"})
	}

	if err := services.RunScheduleNow(scheduleID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Schedule execution started"})
}
