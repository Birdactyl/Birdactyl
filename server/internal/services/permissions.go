package services

import (
	"encoding/json"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
)

func GetUserServerPermissions(userID, serverID uuid.UUID) ([]string, error) {
	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return nil, err
	}

	if server.UserID == userID {
		return models.OwnerPermissions(), nil
	}

	var subuser models.Subuser
	if err := database.DB.Where("server_id = ? AND user_id = ?", serverID, userID).First(&subuser).Error; err != nil {
		return nil, err
	}

	return subuser.GetPermissions(), nil
}

func CanAccessServer(userID, serverID uuid.UUID, isAdmin bool) bool {
	if isAdmin {
		return true
	}

	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return false
	}

	if server.UserID == userID {
		return true
	}

	var count int64
	database.DB.Model(&models.Subuser{}).Where("server_id = ? AND user_id = ?", serverID, userID).Count(&count)
	return count > 0
}

func HasServerPermission(userID, serverID uuid.UUID, isAdmin bool, permission string) bool {
	if isAdmin {
		return true
	}

	perms, err := GetUserServerPermissions(userID, serverID)
	if err != nil {
		return false
	}

	return models.HasPermission(perms, permission)
}

func GetSubusers(serverID uuid.UUID) ([]models.Subuser, error) {
	var subusers []models.Subuser
	err := database.DB.Preload("User").Where("server_id = ?", serverID).Find(&subusers).Error
	return subusers, err
}

func AddSubuser(serverID, userID uuid.UUID, permissions []string) (*models.Subuser, error) {
	permsJSON, _ := json.Marshal(permissions)
	subuser := &models.Subuser{
		ServerID:    serverID,
		UserID:      userID,
		Permissions: permsJSON,
	}
	err := database.DB.Create(subuser).Error
	return subuser, err
}

func UpdateSubuserPermissions(subuserID uuid.UUID, permissions []string) error {
	permsJSON, _ := json.Marshal(permissions)
	return database.DB.Model(&models.Subuser{}).Where("id = ?", subuserID).Update("permissions", permsJSON).Error
}

func RemoveSubuser(subuserID uuid.UUID) error {
	return database.DB.Where("id = ?", subuserID).Delete(&models.Subuser{}).Error
}
