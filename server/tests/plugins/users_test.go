package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_UserLifecycle(t *testing.T) {
	requireDB(t)

	database.DB.Exec("DELETE FROM subusers")
	database.DB.Exec("DELETE FROM activity_logs")
	database.DB.Exec("DELETE FROM ip_registrations")
	database.DB.Exec("DELETE FROM users")

	api := pluginAPI

	createdUser, err := api.CreateUser("sdk-user@example.com", "sdk_tester", "mypassword123")
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, "sdk_tester", createdUser.Username)

	createdUser2, err := api.CreateUser("sdk-user2@example.com", "sdk_tester2", "mypassword123")
	assert.NoError(t, err)

	user1, err := api.GetUser(createdUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.Email, user1.Email)

	userByEmail, err := api.GetUserByEmail("sdk-user2@example.com")
	assert.NoError(t, err)
	assert.Equal(t, createdUser2.ID, userByEmail.ID)

	userByUsername, err := api.GetUserByUsername("sdk_tester")
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, userByUsername.ID)

	users := api.ListUsers()
	assert.GreaterOrEqual(t, len(users), 2)

	usernameUpdate := "updated_tester"
	emailUpdate := "sdk-updated@example.com"
	updatedUser, err := api.UpdateUser(createdUser.ID, &usernameUpdate, &emailUpdate)
	assert.NoError(t, err)
	assert.Equal(t, "updated_tester", updatedUser.Username)
	assert.Equal(t, "sdk-updated@example.com", updatedUser.Email)

	err = api.BanUser(createdUser.ID)
	assert.NoError(t, err)
	bannedUser, _ := api.GetUser(createdUser.ID)
	assert.True(t, bannedUser.IsBanned)

	err = api.UnbanUser(createdUser.ID)
	assert.NoError(t, err)
	unbannedUser, _ := api.GetUser(createdUser.ID)
	assert.False(t, unbannedUser.IsBanned)

	err = api.SetAdmin(createdUser.ID)
	assert.NoError(t, err)
	adminUser, _ := api.GetUser(createdUser.ID)
	assert.True(t, adminUser.IsAdmin)

	err = api.RevokeAdmin(createdUser.ID)
	assert.NoError(t, err)
	revokedUser, _ := api.GetUser(createdUser.ID)
	assert.False(t, revokedUser.IsAdmin)

	err = api.ForcePasswordReset(createdUser.ID)
	assert.NoError(t, err)
	forcedUser, _ := api.GetUser(createdUser.ID)
	assert.True(t, forcedUser.ForcePasswordReset)

	err = api.RequestPasswordReset(updatedUser.Email)
	if err != nil && err.Error() != "rpc error: code = Internal desc = SMTP is disabled" {
		assert.NoError(t, err)
	}

	err = api.SendVerificationEmail(createdUser.ID)
	if err != nil && err.Error() != "rpc error: code = Internal desc = SMTP is disabled" {
		assert.NoError(t, err)
	}

	status2fa, err := api.GetUser2FAStatus(createdUser.ID)
	assert.NoError(t, err)
	assert.False(t, status2fa)

	err = api.AdminDisable2FA(createdUser.ID)
	assert.NoError(t, err)

	ramLimit := int32(2048)
	cpuLimit := int32(100)
	diskLimit := int32(5000)
	serverLimit := int32(2)
	err = api.SetUserResources(createdUser.ID, &ramLimit, &cpuLimit, &diskLimit, &serverLimit)
	assert.NoError(t, err)

	resourcesUser, _ := api.GetUser(createdUser.ID)
	assert.Equal(t, ramLimit, resourcesUser.RamLimit)
	assert.Equal(t, cpuLimit, resourcesUser.CpuLimit)
	assert.Equal(t, diskLimit, resourcesUser.DiskLimit)
	assert.Equal(t, serverLimit, resourcesUser.ServerLimit)

	err = api.DeleteUser(createdUser.ID)
	assert.NoError(t, err)
	err = api.DeleteUser(createdUser2.ID)
	assert.NoError(t, err)
	
	var count int64
	database.DB.Model(&models.User{}).Where("id = ?", createdUser.ID).Count(&count)
	assert.Equal(t, int64(0), count)
	
	database.DB.Model(&models.User{}).Where("id = ?", createdUser2.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
