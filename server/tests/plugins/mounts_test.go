package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_MountLifecycle(t *testing.T) {
	requireDB(t)

	database.DB.Exec("DELETE FROM server_mounts")
	database.DB.Exec("DELETE FROM node_mounts")
	database.DB.Exec("DELETE FROM package_mounts")
	database.DB.Exec("DELETE FROM mounts")

	api := pluginAPI

	desc := "A test plugin mount"
	src := "/tmp/source"
	tgt := "/tmp/target"
	
	mount, err := api.CreateMount("Test SDK Mount", desc, src, tgt, false, true, true)
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.Equal(t, "Test SDK Mount", mount.Name)
	assert.Equal(t, desc, mount.Description)
	assert.Equal(t, tgt, mount.Target)
	assert.Equal(t, true, mount.UserMountable)
	assert.NotEmpty(t, mount.ID)

	mount2, err := api.CreateMount("Test Backend Mount", "Other mount", "/tmp/source2", "/tmp/target2", true, false, false)
	assert.NoError(t, err)

	fetchedMount, err := api.GetMount(mount.ID)
	assert.NoError(t, err)
	assert.Equal(t, mount.ID, fetchedMount.ID)

	mounts := api.ListMounts()
	assert.GreaterOrEqual(t, len(mounts), 2)
	
	newDesc := "Updated sdk mount description"
	navigable := false
	updatedMount, err := api.UpdateMount(mount.ID, nil, &newDesc, nil, nil, nil, nil, &navigable)
	assert.NoError(t, err)
	assert.NotNil(t, updatedMount)
	assert.Equal(t, newDesc, updatedMount.Description)
	assert.Equal(t, false, updatedMount.Navigable)

	err = api.DeleteMount(mount.ID)
	assert.NoError(t, err)
	err = api.DeleteMount(mount2.ID)
	assert.NoError(t, err)

	var count int64
	database.DB.Model(&models.Mount{}).Where("id = ?", mount.ID).Count(&count)
	assert.Equal(t, int64(0), count)
	
	database.DB.Model(&models.Mount{}).Where("id = ?", mount2.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
