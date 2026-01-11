package services

import (
	"errors"
	"fmt"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailTaken          = errors.New("email already taken")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrIPLimitReached      = errors.New("account limit reached for this IP")
	ErrSessionLimitReached = errors.New("maximum sessions reached")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrSessionExpired      = errors.New("session expired")
	ErrUserBanned          = errors.New("account is banned")
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AccessClaims struct {
	UserID    uuid.UUID `json:"uid"`
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	return []byte(config.Get().Auth.JWTSecret)
}

func HashPassword(password string) (string, error) {
	cfg := config.Get()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Auth.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Register(email, username, password, ip, userAgent string) (*models.User, *TokenPair, error) {
	cfg := config.Get()

	var ipCount int64
	database.DB.Model(&models.IPRegistration{}).Where("ip = ?", ip).Count(&ipCount)
	if ipCount >= int64(cfg.Auth.AccountsPerIP) {
		return nil, nil, ErrIPLimitReached
	}

	var existing models.User
	if err := database.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, nil, ErrEmailTaken
	}
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Auth.BcryptCost)
	if err != nil {
		return nil, nil, err
	}

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	isFirstUser := userCount == 0

	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		RegisterIP:   ip,
		IsAdmin:      isFirstUser,
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&models.IPRegistration{IP: ip, UserID: user.ID}).Error
	})
	if err != nil {
		return nil, nil, err
	}

	if isFirstUser {
		config.AddRootAdmin(user.ID.String())
	}

	tokens, err := createSession(user.ID, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func Login(email, password, ip, userAgent string) (*models.User, *TokenPair, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if user.IsBanned {
		return nil, nil, ErrUserBanned
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := createSession(user.ID, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func RefreshTokens(refreshToken, ip, userAgent string) (*TokenPair, error) {
	var session models.Session
	err := database.DB.Where("refresh_token = ?", refreshToken).First(&session).Error
	if err != nil {
		if err := database.DB.Where("previous_token = ?", refreshToken).First(&session).Error; err != nil {
			var count int64
			database.DB.Model(&models.Session{}).Count(&count)
			fmt.Printf("[DEBUG] RefreshTokens failed - token not found. Total sessions in DB: %d\n", count)
			return nil, ErrInvalidToken
		}
		if session.LastRefreshAt != nil && time.Since(*session.LastRefreshAt) < 10*time.Second {
			return generateTokensForSession(&session)
		}
		return nil, ErrInvalidToken
	}

	if time.Now().After(session.ExpiresAt) {
		database.DB.Delete(&session)
		return nil, ErrSessionExpired
	}

	return refreshSession(&session, ip, userAgent)
}

func RefreshBySessionID(sessionID uuid.UUID, ip, userAgent string) (*TokenPair, error) {
	var session models.Session
	if err := database.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(session.ExpiresAt) {
		database.DB.Delete(&session)
		return nil, ErrSessionExpired
	}

	return refreshSession(&session, ip, userAgent)
}

func refreshSession(session *models.Session, ip, userAgent string) (*TokenPair, error) {
	cfg := config.Get()

	session.IP = ip
	session.UserAgent = userAgent
	session.ExpiresAt = time.Now().Add(time.Duration(cfg.Auth.RefreshTokenExpiry) * time.Minute)
	now := time.Now()
	session.LastRefreshAt = &now

	if err := database.DB.Save(session).Error; err != nil {
		return nil, err
	}

	return generateTokensForSession(session)
}

func generateTokensForSession(session *models.Session) (*TokenPair, error) {
	cfg := config.Get()
	accessExpiry := time.Now().Add(time.Duration(cfg.Auth.AccessTokenExpiry) * time.Minute)
	claims := &AccessClaims{
		UserID:    session.UserID,
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
		RefreshToken: session.RefreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func ValidateAccessToken(tokenString string) (*AccessClaims, bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		return nil, false, ErrInvalidToken
	}

	claims := token.Claims.(*AccessClaims)

	var session models.Session
	if err := database.DB.Where("id = ?", claims.SessionID).First(&session).Error; err != nil {
		return nil, false, ErrInvalidToken
	}

	cfg := config.Get()
	needsRefresh := time.Until(claims.ExpiresAt.Time) <= time.Duration(cfg.Auth.TokenRefreshThreshold)*time.Minute

	return claims, needsRefresh, nil
}
