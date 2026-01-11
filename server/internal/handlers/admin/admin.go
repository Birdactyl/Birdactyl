package admin

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/handlers"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/plugins"
	"birdactyl-panel-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type PaginatedUsers struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
	AdminCount int64          `json:"admin_count"`
}

type UserResponse struct {
	models.User
	IsRootAdmin bool `json:"is_root_admin"`
}

func AdminGetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	search := c.Query("search", "")
	filter := c.Query("filter", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	query := database.DB.Model(&models.User{})
	if search != "" {
		val := database.ILikeValue(search)
		query = query.Where(database.ILike("username", val)+" OR "+database.ILike("email", val), val, val)
	}
	if filter == "admin" {
		query = query.Where("is_admin = ?", true)
	} else if filter == "banned" {
		query = query.Where("is_banned = ?", true)
	}

	var total int64
	query.Count(&total)

	var adminCount int64
	database.DB.Model(&models.User{}).Where("is_admin = ?", true).Count(&adminCount)

	var users []models.User
	offset := (page - 1) * perPage
	query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&users)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	result, _ := plugins.ExecuteMixin(string(plugins.MixinUserList), map[string]interface{}{"users": users, "page": page, "per_page": perPage, "search": search}, func(input map[string]interface{}) (interface{}, error) {
		return input["users"], nil
	})
	if result != nil {
		users = result.([]models.User)
	}

	userResponses := make([]UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = UserResponse{
			User:        u,
			IsRootAdmin: config.IsRootAdmin(u.ID.String()),
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": PaginatedUsers{
			Users:      userResponses,
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
			AdminCount: adminCount,
		},
	})
}

type AdminCreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func AdminCreateUser(c *fiber.Ctx) error {
	var req AdminCreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email, username and password are required",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Password must be at least 8 characters",
		})
	}

	mixinInput := map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	}

	result, err := plugins.ExecuteMixin(string(plugins.MixinUserCreate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return services.AdminCreateUser(req.Email, req.Username, req.Password)
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
		status := fiber.StatusInternalServerError
		if err == services.ErrEmailTaken || err == services.ErrUsernameTaken {
			status = fiber.StatusConflict
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	user := result.(*models.User)
	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminUserCreate, "Created user: "+req.Username, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"target_user": req.Username})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

type BulkUserIDsRequest struct {
	UserIDs []string `json:"user_ids"`
}

func AdminBanUsers(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	var skippedSelf, skippedRoot bool
	for _, id := range req.UserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		if uid == currentUser.ID {
			skippedSelf = true
			continue
		}
		if config.IsRootAdmin(id) {
			skippedRoot = true
			continue
		}
		uuids = append(uuids, uid)
	}

	if len(uuids) == 0 {
		if skippedRoot {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot ban root admins"})
		}
		if skippedSelf {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Cannot ban yourself"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No valid users to ban"})
	}

	var users []models.User
	database.DB.Where("id IN ?", uuids).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	affected := int64(0)
	for i, uid := range uuids {
		mixinInput := map[string]interface{}{
			"user_id":  uid.String(),
			"username": usernames[i],
		}

		_, err := plugins.ExecuteMixin(string(plugins.MixinUserBan), mixinInput, func(input map[string]interface{}) (interface{}, error) {
			res := database.DB.Model(&models.User{}).Where("id = ?", uid).Update("is_banned", true)
			database.DB.Where("user_id = ?", uid).Delete(&models.Session{})
			return nil, res.Error
		})

		if err == nil {
			affected++
			plugins.Emit(plugins.EventUserBanned, map[string]string{"user_id": uid.String(), "username": usernames[i]})
		}
	}

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserBan, "Banned users: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}

func AdminUnbanUsers(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	for _, id := range req.UserIDs {
		if uid, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, uid)
		}
	}

	var users []models.User
	database.DB.Where("id IN ?", uuids).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	affected := int64(0)
	for i, uid := range uuids {
		mixinInput := map[string]interface{}{
			"user_id":  uid.String(),
			"username": usernames[i],
		}

		_, err := plugins.ExecuteMixin(string(plugins.MixinUserUnban), mixinInput, func(input map[string]interface{}) (interface{}, error) {
			res := database.DB.Model(&models.User{}).Where("id = ?", uid).Update("is_banned", false)
			return nil, res.Error
		})

		if err == nil {
			affected++
			plugins.Emit(plugins.EventUserUnbanned, map[string]string{"user_id": uid.String()})
		}
	}

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserUnban, "Unbanned users: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}

func AdminSetAdmin(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	for _, id := range req.UserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		if uid != currentUser.ID {
			uuids = append(uuids, uid)
		}
	}

	if len(uuids) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No valid users to modify"})
	}

	var users []models.User
	database.DB.Where("id IN ? AND is_banned = ?", uuids, false).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	result := database.DB.Model(&models.User{}).Where("id IN ? AND is_banned = ?", uuids, false).Update("is_admin", true)

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserSetAdmin, "Granted admin to: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": result.RowsAffected}})
}

func AdminRevokeAdmin(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	var skippedRoot []string
	for _, id := range req.UserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		if uid == currentUser.ID {
			continue
		}
		if config.IsRootAdmin(id) {
			skippedRoot = append(skippedRoot, id)
			continue
		}
		uuids = append(uuids, uid)
	}

	if len(uuids) == 0 {
		if len(skippedRoot) > 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot revoke admin from root admins"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Cannot revoke your own admin"})
	}

	var users []models.User
	database.DB.Where("id IN ?", uuids).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	result := database.DB.Model(&models.User{}).Where("id IN ?", uuids).Update("is_admin", false)

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserRevokeAdm, "Revoked admin from: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": result.RowsAffected}})
}

type AdminUpdateUserRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RAMLimit    *int   `json:"ram_limit"`
	CPULimit    *int   `json:"cpu_limit"`
	DiskLimit   *int   `json:"disk_limit"`
	ServerLimit *int   `json:"server_limit"`
}

func AdminUpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid user ID"})
	}

	var req AdminUpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request body"})
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "User not found"})
	}

	updates := map[string]interface{}{}

	if req.Email != "" && req.Email != user.Email {
		if !emailRegex.MatchString(req.Email) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid email format"})
		}
		var existing models.User
		if database.DB.Where("email = ? AND id != ?", req.Email, id).First(&existing).Error == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Email already in use"})
		}
		updates["email"] = req.Email
	}

	if req.Username != "" && req.Username != user.Username {
		var existing models.User
		if database.DB.Where("username = ? AND id != ?", req.Username, id).First(&existing).Error == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Username already in use"})
		}
		updates["username"] = req.Username
	}

	if req.Password != "" {
		if len(req.Password) < 8 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Password must be at least 8 characters"})
		}
		hash, err := services.HashPassword(req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": "Failed to hash password"})
		}
		updates["password_hash"] = hash
	}

	if req.RAMLimit != nil {
		if *req.RAMLimit == 0 {
			updates["ram_limit"] = nil
		} else {
			updates["ram_limit"] = *req.RAMLimit
		}
	}
	if req.CPULimit != nil {
		if *req.CPULimit == 0 {
			updates["cpu_limit"] = nil
		} else {
			updates["cpu_limit"] = *req.CPULimit
		}
	}
	if req.DiskLimit != nil {
		if *req.DiskLimit == 0 {
			updates["disk_limit"] = nil
		} else {
			updates["disk_limit"] = *req.DiskLimit
		}
	}
	if req.ServerLimit != nil {
		if *req.ServerLimit == 0 {
			updates["server_limit"] = nil
		} else {
			updates["server_limit"] = *req.ServerLimit
		}
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No changes provided"})
	}

	mixinInput := map[string]interface{}{
		"user_id":  id.String(),
		"username": user.Username,
		"updates":  updates,
	}

	_, err = plugins.ExecuteMixin(string(plugins.MixinUserUpdate), mixinInput, func(input map[string]interface{}) (interface{}, error) {
		return nil, database.DB.Model(&user).Updates(updates).Error
	})

	if err != nil {
		if mixinErr, ok := err.(*plugins.MixinError); ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": mixinErr.Message})
		}
	}

	admin := c.Locals("user").(*models.User)
	handlers.LogActivity(admin.ID, admin.Username, handlers.ActionAdminUserUpdate, "Updated user: "+user.Username, c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"target_user_id": id})

	plugins.Emit(plugins.EventUserUpdated, map[string]string{"user_id": id.String(), "username": user.Username})

	return c.JSON(fiber.Map{"success": true, "message": "User updated"})
}

func AdminForcePasswordReset(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	for _, id := range req.UserIDs {
		if uid, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, uid)
		}
	}

	var users []models.User
	database.DB.Where("id IN ?", uuids).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	result := database.DB.Model(&models.User{}).Where("id IN ?", uuids).Update("force_password_reset", true)
	database.DB.Where("user_id IN ?", uuids).Delete(&models.Session{})

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserForceReset, "Forced password reset for: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": result.RowsAffected}})
}

func AdminDeleteUsers(c *fiber.Ctx) error {
	currentUser := c.Locals("user").(*models.User)
	var req BulkUserIDsRequest
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Invalid request"})
	}

	var uuids []uuid.UUID
	var skippedSelf, skippedRoot bool
	for _, id := range req.UserIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		if uid == currentUser.ID {
			skippedSelf = true
			continue
		}
		if config.IsRootAdmin(id) {
			skippedRoot = true
			continue
		}
		uuids = append(uuids, uid)
	}

	if len(uuids) == 0 {
		if skippedRoot {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "error": "Cannot delete root admins"})
		}
		if skippedSelf {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "Cannot delete yourself"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "No valid users to delete"})
	}

	var serverCount int64
	database.DB.Model(&models.Server{}).Where("user_id IN ?", uuids).Count(&serverCount)
	if serverCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "error": "Cannot delete users that own servers"})
	}

	var users []models.User
	database.DB.Where("id IN ?", uuids).Find(&users)
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	affected := int64(0)
	for i, uid := range uuids {
		mixinInput := map[string]interface{}{
			"user_id":  uid.String(),
			"username": usernames[i],
		}

		_, err := plugins.ExecuteMixin(string(plugins.MixinUserDelete), mixinInput, func(input map[string]interface{}) (interface{}, error) {
			database.DB.Where("user_id = ?", uid).Delete(&models.Session{})
			database.DB.Where("user_id = ?", uid).Delete(&models.ActivityLog{})
			return nil, database.DB.Where("id = ?", uid).Delete(&models.User{}).Error
		})

		if err == nil {
			affected++
			plugins.Emit(plugins.EventUserDeleted, map[string]string{"user_id": uid.String(), "username": usernames[i]})
		}
	}

	handlers.LogActivity(currentUser.ID, currentUser.Username, handlers.ActionAdminUserDelete, "Deleted users: "+strings.Join(usernames, ", "), c.IP(), c.Get("User-Agent"), true, map[string]interface{}{"users": usernames})

	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"affected": affected}})
}
