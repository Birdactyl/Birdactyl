package plugins_test

import (
	"testing"
	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_NodeLifecycle(t *testing.T) {
	requireDB(t)

	database.DB.Exec("DELETE FROM servers")
	database.DB.Exec("DELETE FROM nodes")

	api := pluginAPI

	node, token, err := api.CreateNode("Test Node", "node1.example.com", 8080)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.NotEmpty(t, token)
	assert.Equal(t, "Test Node", node.Name)
	assert.Equal(t, "node1.example.com", node.FQDN)
	assert.Equal(t, int32(8080), node.Port)

	node2, _, err := api.CreateNode("Second Node", "node2.example.com", 8081)
	assert.NoError(t, err)

	fetchedNode, err := api.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, fetchedNode.ID)
	assert.Equal(t, node.FQDN, fetchedNode.FQDN)

	nodes := api.ListNodes()
	assert.GreaterOrEqual(t, len(nodes), 2)

	newToken, err := api.ResetNodeToken(node.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, token, newToken)

	err = api.DeleteNode(node.ID)
	assert.NoError(t, err)
	
	err = api.DeleteNode(node2.ID)
	assert.NoError(t, err)

	var count int64
	database.DB.Model(&models.Node{}).Where("id = ?", node.ID).Count(&count)
	assert.Equal(t, int64(0), count)
	
	database.DB.Model(&models.Node{}).Where("id = ?", node2.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
