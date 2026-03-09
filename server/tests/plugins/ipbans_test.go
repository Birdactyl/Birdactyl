package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_IPBanLifecycle(t *testing.T) {
	requireDB(t)

	database.DB.Exec("DELETE FROM ip_bans")

	api := pluginAPI

	ban, err := api.CreateIPBan("192.168.1.100", "Spamming API")
	assert.NoError(t, err)
	assert.NotNil(t, ban)
	assert.Equal(t, "192.168.1.100", ban.IP)
	assert.Equal(t, "Spamming API", ban.Reason)

	ban2, err := api.CreateIPBan("10.0.0.50", "Malicious activity")
	assert.NoError(t, err)

	bans := api.ListIPBans()
	assert.GreaterOrEqual(t, len(bans), 2)

	found := false
	for _, b := range bans {
		if b.IP == "192.168.1.100" {
			found = true
			break
		}
	}
	assert.True(t, found, "First ban should exist in the list")

	err = api.DeleteIPBan(ban.ID)
	assert.NoError(t, err)
	
	err = api.DeleteIPBan(ban2.ID)
	assert.NoError(t, err)

	var count int64
	database.DB.Model(&models.IPBan{}).Where("id = ?", ban.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
