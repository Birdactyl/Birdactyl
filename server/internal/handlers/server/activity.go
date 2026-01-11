package server

import (
	"math"
	"strconv"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PaginatedServerLogs struct {
	Logs       []models.ActivityLog `json:"logs"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

func GetServerActivity(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	serverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid server ID"})
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Server not found"})
	}

	if !user.IsAdmin && server.UserID != user.ID {
		if !services.HasServerPermission(user.ID, serverID, false, models.PermActivityView) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Permission denied"})
		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")
	from := c.Query("from", "")
	to := c.Query("to", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	serverIDStr := serverID.String()
	query := database.DB.Model(&models.ActivityLog{}).Where(
		"action LIKE ? AND metadata LIKE ?",
		"server.%",
		"%"+serverIDStr+"%",
	)

	if search != "" {
		val := database.ILikeValue(search)
		query = query.Where(
			database.ILike("username", val)+" OR "+database.ILike("action", val)+" OR "+database.ILike("description", val),
			val, val, val,
		)
	}

	if from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	var total int64
	query.Count(&total)

	var logs []models.ActivityLog
	offset := (page - 1) * perPage
	query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&logs)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return c.JSON(fiber.Map{
		"success": true,
		"data": PaginatedServerLogs{
			Logs:       logs,
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
