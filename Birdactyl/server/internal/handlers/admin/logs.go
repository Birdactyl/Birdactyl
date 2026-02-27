package admin

import (
	"math"
	"strconv"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"

	"github.com/gofiber/fiber/v2"
)

type PaginatedLogs struct {
	Logs       []models.ActivityLog `json:"logs"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

func AdminGetLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")
	filter := c.Query("filter", "")
	from := c.Query("from", "")
	to := c.Query("to", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := database.DB.Model(&models.ActivityLog{})

	if search != "" {
		val := database.ILikeValue(search)
		query = query.Where(database.ILike("username", val)+" OR "+database.ILike("action", val)+" OR "+database.ILike("description", val), val, val, val)
	}

	if filter == "admin" {
		query = query.Where("is_admin = ?", true)
	} else if filter == "user" {
		query = query.Where("is_admin = ?", false)
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

	plugins.ExecuteMixin(string(plugins.MixinActivityLogList), map[string]interface{}{"logs": logs, "page": page, "per_page": perPage}, func(input map[string]interface{}) (interface{}, error) {
		return input["logs"], nil
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data": PaginatedLogs{
			Logs:       logs,
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}