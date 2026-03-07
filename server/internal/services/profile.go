package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SendEmailChangeCode(userID uuid.UUID) error {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())

	Cache.Set("email_change_code_"+userID.String(), code, 10*time.Minute)

	subject := "Confirm your email change"
	htmlBody := fmt.Sprintf(`
		<h2>Email Change Verification</h2>
		<p>You have requested to change the email address for your Birdactyl account.</p>
		<p>Please enter the following verification code to confirm this change:</p>
		<h3 style="background:#f4f4f5;padding:12px;display:inline-block;border-radius:6px;letter-spacing:4px;">%s</h3>
		<p>If you did not request this change, please change your password immediately.</p>
	`, code)

	return SendEmail(user.Email, subject, htmlBody)
}

func GetUserByID(id uuid.UUID) (*models.User, error) {
	cacheKey := "user_" + id.String()
	if cached, found := Cache.Get(cacheKey); found {
		return cached.(*models.User), nil
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	Cache.Set(cacheKey, &user, 30*time.Duration(1000000000)) // 30 seconds
	return &user, nil
}

func AdminCreateUser(email, username, password string) (*models.User, error) {
	cfg := config.Get()

	var existing models.User
	if err := database.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	}
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Auth.BcryptCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		RegisterIP:   "admin-created",
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateProfile(userID uuid.UUID, username, email, securityCode string) (*models.User, bool, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, false, err
	}

	emailChanged := false

	if username != "" && username != user.Username {
		var existing models.User
		if err := database.DB.Where("username = ? AND id != ?", username, userID).First(&existing).Error; err == nil {
			return nil, false, ErrUsernameTaken
		}
		user.Username = username
	}

	if email != "" && email != user.Email {
		if user.TOTPEnabled {
			if securityCode == "" || !VerifyTOTP(user.ID, securityCode) {
				return nil, false, ErrTOTPInvalidCode
			}
		} else {
			cachedCode, found := Cache.Get("email_change_code_" + user.ID.String())
			if !found || cachedCode.(string) != securityCode {
				return nil, false, errors.New("invalid or expired email verification code")
			}
			Cache.Delete("email_change_code_" + user.ID.String())
		}

		var existing models.User
		if err := database.DB.Where("email = ? AND id != ?", email, userID).First(&existing).Error; err == nil {
			return nil, false, ErrEmailTaken
		}
		user.Email = email
		user.EmailVerified = false
		emailChanged = true
	}

	if err := database.DB.Save(&user).Error; err != nil {
		return nil, false, err
	}

	Cache.Delete("user_" + userID.String())

	return &user, emailChanged, nil
}

func UpdatePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	cfg := config.Get()

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), cfg.Auth.BcryptCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ForcePasswordReset = false
	err = database.DB.Save(&user).Error
	if err == nil {
		Cache.Delete("user_" + userID.String())
	}
	return err
}
