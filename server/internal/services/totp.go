package services

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTOTPAlreadyEnabled  = errors.New("2FA is already enabled")
	ErrTOTPNotEnabled      = errors.New("2FA is not enabled")
	ErrTOTPNotSetup        = errors.New("2FA has not been set up yet")
	ErrTOTPInvalidCode     = errors.New("invalid 2FA code")
	ErrTOTPInvalidPassword = errors.New("invalid password")
	Err2FARequired         = errors.New("2fa_required")
)

type TwoFactorSetupResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

type TwoFactorChallengeClaims struct {
	UserID uuid.UUID `json:"uid"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

func SetupTOTP(userID uuid.UUID) (*TwoFactorSetupResponse, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if user.TOTPEnabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Birdactyl",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, err
	}

	database.DB.Model(&user).Update("totp_secret", key.Secret())

	return &TwoFactorSetupResponse{
		Secret: key.Secret(),
		URL:    key.URL(),
	}, nil
}

func EnableTOTP(userID uuid.UUID, code string) ([]string, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if user.TOTPEnabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	if user.TOTPSecret == "" {
		return nil, ErrTOTPNotSetup
	}

	if !totp.Validate(code, user.TOTPSecret) {
		return nil, ErrTOTPInvalidCode
	}

	codes, err := generateBackupCodes()
	if err != nil {
		return nil, err
	}

	hashedCodes, err := hashBackupCodes(codes)
	if err != nil {
		return nil, err
	}

	codesJSON, _ := json.Marshal(hashedCodes)
	database.DB.Model(&user).Updates(map[string]interface{}{
		"totp_enabled": true,
		"backup_codes": string(codesJSON),
	})

	return codes, nil
}

func DisableTOTP(userID uuid.UUID, password string) error {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if !user.TOTPEnabled {
		return ErrTOTPNotEnabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrTOTPInvalidPassword
	}

	database.DB.Model(&user).Updates(map[string]interface{}{
		"totp_enabled": false,
		"totp_secret":  "",
		"backup_codes": "",
	})

	return nil
}

func RegenerateBackupCodes(userID uuid.UUID, password string) ([]string, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if !user.TOTPEnabled {
		return nil, ErrTOTPNotEnabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrTOTPInvalidPassword
	}

	codes, err := generateBackupCodes()
	if err != nil {
		return nil, err
	}

	hashedCodes, err := hashBackupCodes(codes)
	if err != nil {
		return nil, err
	}

	codesJSON, _ := json.Marshal(hashedCodes)
	database.DB.Model(&user).Update("backup_codes", string(codesJSON))

	return codes, nil
}

func VerifyTOTP(userID uuid.UUID, code string) bool {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	if !user.TOTPEnabled {
		return false
	}

	if totp.Validate(code, user.TOTPSecret) {
		return true
	}

	return consumeBackupCode(&user, code)
}

func GenerateChallengeToken(userID uuid.UUID) (string, error) {
	claims := &TwoFactorChallengeClaims{
		UserID: userID,
		Type:   "2fa_challenge",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(getJWTSecret())
}

func ValidateChallengeToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TwoFactorChallengeClaims{}, func(t *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	claims := token.Claims.(*TwoFactorChallengeClaims)
	if claims.Type != "2fa_challenge" {
		return uuid.Nil, ErrInvalidToken
	}

	return claims.UserID, nil
}

func VerifyTwoFactor(challengeToken, code, ip, userAgent string) (*models.User, *TokenPair, error) {
	userID, err := ValidateChallengeToken(challengeToken)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	if !VerifyTOTP(userID, code) {
		return nil, nil, ErrTOTPInvalidCode
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, nil, errors.New("user not found")
	}

	tokens, err := createSession(userID, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func generateBackupCodes() ([]string, error) {
	codes := make([]string, 8)
	for i := range codes {
		n, err := rand.Int(rand.Reader, big.NewInt(99999999))
		if err != nil {
			return nil, err
		}
		codes[i] = fmt.Sprintf("%08d", n.Int64())
	}
	return codes, nil
}

func hashBackupCodes(codes []string) ([]string, error) {
	cfg := config.Get()
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), cfg.Auth.BcryptCost)
		if err != nil {
			return nil, err
		}
		hashed[i] = string(hash)
	}
	return hashed, nil
}

func consumeBackupCode(user *models.User, code string) bool {
	if user.BackupCodes == "" {
		return false
	}

	var hashes []string
	if err := json.Unmarshal([]byte(user.BackupCodes), &hashes); err != nil {
		return false
	}

	for i, hash := range hashes {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			hashes = append(hashes[:i], hashes[i+1:]...)
			updated, _ := json.Marshal(hashes)
			database.DB.Model(user).Update("backup_codes", string(updated))
			return true
		}
	}

	return false
}
