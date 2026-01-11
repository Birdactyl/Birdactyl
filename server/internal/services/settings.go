package services

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
)

func GetSetting(key string) string {
	var setting models.Setting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

func SetSetting(key, value string) error {
	return database.DB.Save(&models.Setting{Key: key, Value: value}).Error
}

func IsRegistrationEnabled() bool {
	val := GetSetting("registration_enabled")
	return val != "false"
}

func IsServerCreationEnabled() bool {
	val := GetSetting("server_creation_enabled")
	return val != "false"
}

func SetRegistrationEnabled(enabled bool) error {
	val := "true"
	if !enabled {
		val = "false"
	}
	return SetSetting("registration_enabled", val)
}

func SetServerCreationEnabled(enabled bool) error {
	val := "true"
	if !enabled {
		val = "false"
	}
	return SetSetting("server_creation_enabled", val)
}
