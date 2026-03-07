package services

import (
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

var ErrEmailNotVerified = errors.New("email verification required")

func GetSetting(key string) string {
	var setting models.Setting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ""
		}
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

func IsEmailVerificationEnabled() bool {
	val := GetSetting("email_verification_enabled")
	return val == "true"
}

func SetEmailVerificationEnabled(enabled bool) error {
	val := "true"
	if !enabled {
		val = "false"
	}
	return SetSetting("email_verification_enabled", val)
}

func GetEmailVerificationRestrictions() []string {
	val := GetSetting("email_verification_restrictions")
	if val == "" {
		return []string{}
	}
	var restrictions []string
	if err := json.Unmarshal([]byte(val), &restrictions); err != nil {
		return []string{}
	}
	return restrictions
}

func SetEmailVerificationRestrictions(restrictions []string) error {
	data, err := json.Marshal(restrictions)
	if err != nil {
		return err
	}
	return SetSetting("email_verification_restrictions", string(data))
}

func CheckEmailVerification(user *models.User, action string) error {
	if user.EmailVerified || user.IsAdmin {
		return nil
	}
	if !IsEmailVerificationEnabled() {
		return nil
	}
	restrictions := GetEmailVerificationRestrictions()
	for _, r := range restrictions {
		if r == action {
			return ErrEmailNotVerified
		}
	}
	return nil
}
