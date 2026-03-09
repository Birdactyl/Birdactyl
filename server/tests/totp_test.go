package tests

import (
	"testing"

	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func TestTOTP(t *testing.T) {
	requireDB(t)

	password := "SecurePassword123!" // very secure
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		ID:           uuid.New(),
		Username:     "test_totp_user",
		Email:        "test_totp_user@test.com",
		PasswordHash: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	defer database.DB.Where("id = ?", user.ID).Delete(&models.User{})

	var setupResp *services.TwoFactorSetupResponse

	t.Run("Setup TOTP", func(t *testing.T) {
		var err error
		setupResp, err = services.SetupTOTP(user.ID)
		if err != nil {
			t.Fatalf("Failed to setup TOTP: %v", err)
		}
		if setupResp.Secret == "" || setupResp.URL == "" {
			t.Errorf("Expected secret and URL to be generated, got secret: %s, url: %s", setupResp.Secret, setupResp.URL)
		}

		database.DB.First(&user, user.ID)
		if user.TOTPSecret != setupResp.Secret {
			t.Errorf("Expected user record to have TOTP secret")
		}
	})

	var backupCodes []string

	t.Run("Enable TOTP", func(t *testing.T) {
		code, err := totp.GenerateCode(setupResp.Secret, time.Now())
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		backupCodes, err = services.EnableTOTP(user.ID, code)
		if err != nil {
			t.Fatalf("Failed to enable TOTP: %v", err)
		}

		if len(backupCodes) == 0 {
			t.Errorf("Expected backup codes to be generated")
		}

		database.DB.First(&user, user.ID)
		if !user.TOTPEnabled {
			t.Errorf("Expected TOTP to be enabled on user record")
		}
		if user.BackupCodes == "" {
			t.Errorf("Expected backup codes to be saved on user record")
		}
	})

	t.Run("Verify TOTP - Valid Code", func(t *testing.T) {
		code, err := totp.GenerateCode(setupResp.Secret, time.Now())
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		ok := services.VerifyTOTP(user.ID, code)
		if !ok {
			t.Errorf("Expected TOTP verification to succeed")
		}
	})

	t.Run("Verify TOTP - Invalid Code", func(t *testing.T) {
		ok := services.VerifyTOTP(user.ID, "000000")
		if ok {
			t.Errorf("Expected TOTP verification to fail with invalid code")
		}
	})

	t.Run("Verify TOTP - Backup Code", func(t *testing.T) {
		if len(backupCodes) == 0 {
			t.Skip("No backup codes generated")
		}
		
		validBackupCode := backupCodes[0]
		ok := services.VerifyTOTP(user.ID, validBackupCode)
		if !ok {
			t.Errorf("Expected backup code verification to succeed")
		}
		
		ok = services.VerifyTOTP(user.ID, validBackupCode)
		if ok {
			t.Errorf("Expected used backup code verification to fail")
		}
	})
	
	t.Run("Regenerate Backup Codes", func(t *testing.T) {
		newCodes, err := services.RegenerateBackupCodes(user.ID, password)
		if err != nil {
			t.Fatalf("Failed to regenerate backup codes: %v", err)
		}
		
		if len(newCodes) == 0 {
			t.Errorf("Expected new backup codes to be generated")
		}
		
		if len(backupCodes) > 1 {
			oldValidCode := backupCodes[1]
			ok := services.VerifyTOTP(user.ID, oldValidCode)
			if ok {
				t.Errorf("Expected old backup code to fail after regeneration")
			}
		}
	})

	t.Run("Disable TOTP", func(t *testing.T) {
		err := services.DisableTOTP(user.ID, password)
		if err != nil {
			t.Fatalf("Failed to disable TOTP: %v", err)
		}

		database.DB.First(&user, user.ID)
		if user.TOTPEnabled {
			t.Errorf("Expected TOTP to be disabled on user record")
		}
		if user.TOTPSecret != "" {
			t.Errorf("Expected TOTP secret to be cleared")
		}
		if user.BackupCodes != "" {
			t.Errorf("Expected backup codes to be cleared")
		}
	})
	
	t.Run("Verify TOTP - Disabled", func(t *testing.T) {
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		ok := services.VerifyTOTP(user.ID, code)
		if ok {
			t.Errorf("Expected TOTP verification to fail if TOTP is disabled")
		}
	})
}
