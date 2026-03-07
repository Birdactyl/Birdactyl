package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"birdactyl-panel-backend/internal/config"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrResetTokenInvalid = errors.New("invalid or expired reset token")
	ErrResetTokenUsed    = errors.New("reset token has already been used")
)

type PasswordResetClaims struct {
	UserID uuid.UUID `json:"uid"`
	Nonce  string    `json:"nonce"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

func RequestPasswordReset(email, baseURL string) error {
	cfg := config.Get()
	if !cfg.SMTPEnabled() {
		return nil
	}

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}

	if user.IsBanned {
		return nil
	}

	nonce := generateNonce()
	database.DB.Model(&user).Update("reset_nonce", nonce)

	token, err := generateResetToken(user.ID, nonce)
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/auth?reset=%s", baseURL, token)

	html := buildResetEmail(user.Username, resetURL)

	go SendEmail(user.Email, "Password Reset - Birdactyl", html)

	return nil
}

func ValidateResetToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(t *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrResetTokenInvalid
	}

	claims := token.Claims.(*PasswordResetClaims)
	if claims.Type != "password_reset" {
		return uuid.Nil, ErrResetTokenInvalid
	}

	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return uuid.Nil, ErrResetTokenInvalid
	}

	if user.ResetNonce != claims.Nonce || user.ResetNonce == "" {
		return uuid.Nil, ErrResetTokenUsed
	}

	return claims.UserID, nil
}

func ResetPassword(tokenString, newPassword string) error {
	userID, err := ValidateResetToken(tokenString)
	if err != nil {
		return err
	}

	cfg := config.Get()
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), cfg.Auth.BcryptCost)
	if err != nil {
		return err
	}

	database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_hash":        string(hash),
		"reset_nonce":          "",
		"force_password_reset": false,
	})

	database.DB.Where("user_id = ?", userID).Delete(&models.Session{})

	return nil
}

func generateResetToken(userID uuid.UUID, nonce string) (string, error) {
	claims := &PasswordResetClaims{
		UserID: userID,
		Nonce:  nonce,
		Type:   "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(getJWTSecret())
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func buildResetEmail(username, resetURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Password Reset</title></head>
<body style="margin:0;padding:0;background-color:#ffffff;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;-webkit-font-smoothing:antialiased;">
<div style="display:none;max-height:0;overflow:hidden;">Reset your Birdactyl password. This link expires in 15 minutes.</div>
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#ffffff;">
<tr><td align="center" style="padding:48px 24px;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width:460px;text-align:left;">
<tr><td style="padding:0 0 40px;">
<span style="font-size:16px;font-weight:700;color:#18181b;letter-spacing:-0.3px;">Birdactyl</span>
</td></tr>
<tr><td>
<h1 style="margin:0 0 20px;font-size:24px;font-weight:600;color:#18181b;letter-spacing:-0.5px;">Reset your password</h1>
<p style="margin:0 0 24px;font-size:15px;color:#52525b;line-height:1.7;">Hey <strong style="color:#18181b;">%s</strong>, we received a request to reset your password. Click the button below to choose a new one.</p>
</td></tr>
<tr><td style="padding:8px 0 32px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%">
<tr><td align="center" style="border-radius:10px;background-color:#18181b;padding:0;">
<a href="%s" target="_blank" style="display:block;padding:14px 0;font-size:14px;font-weight:600;color:#ffffff;text-decoration:none;text-align:center;line-height:1;">Reset password</a>
</td></tr>
</table>
</td></tr>
<tr><td style="border-top:1px solid #e4e4e7;padding:24px 0 0;">
<p style="margin:0 0 16px;font-size:13px;color:#a1a1aa;line-height:1.6;">This link expires in 15 minutes. If you didn't request this, ignore this email.</p>
<p style="margin:0;font-size:12px;color:#a1a1aa;line-height:1.5;word-break:break-all;font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,Courier,monospace;">%s</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, username, resetURL, resetURL)
}
