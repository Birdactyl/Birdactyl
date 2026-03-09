package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_PackageLifecycle(t *testing.T) {
	requireDB(t)

	database.DB.Exec("DELETE FROM packages")

	api := pluginAPI

	pkg, err := api.CreatePackage(
		"Test Package", "A dummy package", "ubuntu:latest",
		"./start.sh", "^C", "{}", 2048, 100, 5000, true,
	)
	assert.NoError(t, err)
	assert.NotNil(t, pkg)
	assert.Equal(t, "Test Package", pkg.Name)
	assert.Equal(t, "ubuntu:latest", pkg.DockerImage)

	pkg2, err := api.CreatePackage(
		"Second Package", "Another dummy package", "alpine:latest",
		"sh start.sh", "^C", "{}", 1024, 50, 2000, true,
	)
	assert.NoError(t, err)

	fetchedPkg, err := api.GetPackage(pkg.ID)
	assert.NoError(t, err)
	assert.Equal(t, pkg.ID, fetchedPkg.ID)
	assert.Equal(t, pkg.Name, fetchedPkg.Name)

	packages := api.ListPackages()
	assert.GreaterOrEqual(t, len(packages), 2)

	newName := "Updated Package"
	updatedPkg, err := api.UpdatePackage(pkg.ID, &newName, nil, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Package", updatedPkg.Name)

	err = api.DeletePackage(pkg.ID)
	assert.NoError(t, err)
	
	err = api.DeletePackage(pkg2.ID)
	assert.NoError(t, err)

	var count int64
	database.DB.Model(&models.Package{}).Where("id = ?", pkg.ID).Count(&count)
	assert.Equal(t, int64(0), count)
	
	database.DB.Model(&models.Package{}).Where("id = ?", pkg2.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
