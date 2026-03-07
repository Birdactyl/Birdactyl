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
)

var (
	ErrAlreadyVerified        = errors.New("email is already verified")
	ErrVerificationTokenInval = errors.New("invalid or expired verification token")
)

type EmailVerificationClaims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

func SendVerificationEmail(userID uuid.UUID, baseURL string) error {
	cfg := config.Get()
	if !cfg.SMTPEnabled() {
		return nil
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	if user.EmailVerified {
		return ErrAlreadyVerified
	}

	token, err := generateVerificationToken(user.ID, user.Email)
	if err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth?verify=%s", baseURL, token)
	html := buildVerificationEmail(user.Username, verifyURL)

	go SendEmail(user.Email, "Verify your email - Birdactyl", html)

	return nil
}

func VerifyEmail(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &EmailVerificationClaims{}, func(t *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		return ErrVerificationTokenInval
	}

	claims := token.Claims.(*EmailVerificationClaims)
	if claims.Type != "email_verification" {
		return ErrVerificationTokenInval
	}

	var user models.User
	if err := database.DB.Where("id = ? AND email = ?", claims.UserID, claims.Email).First(&user).Error; err != nil {
		return ErrVerificationTokenInval
	}

	if user.EmailVerified {
		return nil
	}

	database.DB.Model(&user).Update("email_verified", true)
	return nil
}

func generateVerificationToken(userID uuid.UUID, email string) (string, error) {
	claims := &EmailVerificationClaims{
		UserID: userID,
		Email:  email,
		Type:   "email_verification",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(getJWTSecret())
}

func buildVerificationEmail(username, verifyURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Verify Email</title></head>
<body style="margin:0;padding:0;background-color:#ffffff;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;-webkit-font-smoothing:antialiased;">
<div style="display:none;max-height:0;overflow:hidden;">Verify your Birdactyl email address.</div>
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#ffffff;">
<tr><td align="center" style="padding:48px 24px;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width:460px;text-align:left;">
<tr><td style="padding:0 0 40px;">
<span style="font-size:16px;font-weight:700;color:#18181b;letter-spacing:-0.3px;">Birdactyl</span>
</td></tr>
<tr><td>
<h1 style="margin:0 0 20px;font-size:24px;font-weight:600;color:#18181b;letter-spacing:-0.5px;">Verify your email</h1>
<p style="margin:0 0 24px;font-size:15px;color:#52525b;line-height:1.7;">Hey <strong style="color:#18181b;">%s</strong>, thanks for creating a Birdactyl account. Click the button below to verify your email address.</p>
</td></tr>
<tr><td style="padding:8px 0 32px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%">
<tr><td align="center" style="border-radius:10px;background-color:#18181b;padding:0;">
<a href="%s" target="_blank" style="display:block;padding:14px 0;font-size:14px;font-weight:600;color:#ffffff;text-decoration:none;text-align:center;line-height:1;">Verify email</a>
</td></tr>
</table>
</td></tr>
<tr><td style="border-top:1px solid #e4e4e7;padding:24px 0 0;">
<p style="margin:0 0 16px;font-size:13px;color:#a1a1aa;line-height:1.6;">This link expires in 24 hours. If you didn't create an account, ignore this email.</p>
<p style="margin:0;font-size:12px;color:#a1a1aa;line-height:1.5;word-break:break-all;font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,Courier,monospace;">%s</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, username, verifyURL, verifyURL)
}
