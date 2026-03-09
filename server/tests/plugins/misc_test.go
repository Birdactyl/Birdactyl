package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_Settings(t *testing.T) {
	requireDB(t)

	api := pluginAPI

	settings := api.GetSettings()
	assert.NotNil(t, settings)

	originalReg := settings.RegistrationEnabled
	
	err := api.SetRegistrationEnabled(!originalReg)
	assert.NoError(t, err)

	newSettings := api.GetSettings()
	assert.Equal(t, !originalReg, newSettings.RegistrationEnabled)

	api.SetRegistrationEnabled(originalReg)
}

func TestPluginSDK_KV(t *testing.T) {
	requireDB(t)
	api := pluginAPI

	api.SetKV("my-plugin-key", "some-amazing-value")

	val, found := api.GetKV("my-plugin-key")
	assert.True(t, found)
	assert.Equal(t, "some-amazing-value", val)

	api.DeleteKV("my-plugin-key")

	val2, found2 := api.GetKV("my-plugin-key")
	assert.False(t, found2)
	assert.Empty(t, val2)
}

func TestPluginSDK_ActivityLogs(t *testing.T) {
	requireDB(t)
	api := pluginAPI
	
	database.DB.Exec("DELETE FROM activity_logs")

	activity := &models.ActivityLog{
		Description: "A test activity",
		Action: "test.action",
		Username: "tester",
	}
	database.DB.Create(activity)

	logs := api.GetActivityLogs(5)
	assert.GreaterOrEqual(t, len(logs), 1)

	found := false
	for _, l := range logs {
		if l.Action == "test.action" {
			found = true
			assert.Equal(t, "A test activity", l.Description)
			break
		}
	}
	assert.True(t, found)
}
