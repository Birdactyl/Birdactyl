package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SessionInfo struct {
	ID        uuid.UUID `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}

func createSession(userID uuid.UUID, ip, userAgent string) (*TokenPair, error) {
	cfg := config.Get()

	var sessionCount int64
	database.DB.Model(&models.Session{}).Where("user_id = ?", userID).Count(&sessionCount)
	if sessionCount >= int64(cfg.Auth.MaxSessionsPerUser) {
		var oldest models.Session
		database.DB.Where("user_id = ?", userID).Order("created_at ASC").First(&oldest)
		database.DB.Delete(&oldest)
	}

	refreshBytes := make([]byte, 48)
	rand.Read(refreshBytes)
	refreshToken := base64.URLEncoding.EncodeToString(refreshBytes)

	session := &models.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IP:           ip,
		ExpiresAt:    time.Now().Add(time.Duration(cfg.Auth.RefreshTokenExpiry) * time.Minute),
	}

	if err := database.DB.Create(session).Error; err != nil {
		return nil, err
	}

	accessExpiry := time.Now().Add(time.Duration(cfg.Auth.AccessTokenExpiry) * time.Minute)
	claims := &AccessClaims{
		UserID:    userID,
		SessionID: session.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := token.SignedString(getJWTSecret())
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func Logout(sessionID uuid.UUID) error {
	return database.DB.Where("id = ?", sessionID).Delete(&models.Session{}).Error
}

func LogoutAll(userID uuid.UUID) error {
	return database.DB.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func CleanExpiredSessions() {
	database.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
}

func GetUserSessions(userID, currentSessionID uuid.UUID) []SessionInfo {
	var sessions []models.Session
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions)

	result := make([]SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = SessionInfo{
			ID:        s.ID,
			IP:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
			IsCurrent: s.ID == currentSessionID,
		}
	}
	return result
}

func RevokeSession(userID uuid.UUID, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return errors.New("invalid session ID")
	}
	result := database.DB.Where("id = ? AND user_id = ?", sid, userID).Delete(&models.Session{})
	if result.RowsAffected == 0 {
		return errors.New("session not found")
	}
	return nil
}

func RevokeOtherSessions(userID, currentSessionID uuid.UUID) {
	database.DB.Where("user_id = ? AND id != ?", userID, currentSessionID).Delete(&models.Session{})
}
