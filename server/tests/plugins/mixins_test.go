package plugins_test

import (
	"errors"
	"testing"
	"birdactyl-panel-backend/internal/plugins"
	"github.com/stretchr/testify/assert"
)

func TestPluginSDK_Mixins(t *testing.T) {
	requireDB(t)

	t.Run("Next Action (Modify Input)", func(t *testing.T) {
		input := map[string]interface{}{"value": "original"}

		result, err := plugins.ExecuteMixin("test.mixin.modify", input, func(v map[string]interface{}) (interface{}, error) {
			return v["value"], nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "original_modified", result)
	})

	t.Run("Return Action (Early Halt)", func(t *testing.T) {
		input := map[string]interface{}{"value": "original"}

		executedCoreLogic := false
		result, err := plugins.ExecuteMixin("test.mixin.return", input, func(v map[string]interface{}) (interface{}, error) {
			executedCoreLogic = true
			return v["value"], nil
		})

		assert.NoError(t, err)
		assert.False(t, executedCoreLogic, "Core logic should NOT have executed")
		
		resMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, resMap["halted"])
		assert.Equal(t, "plugin returned early", resMap["reason"])
	})

	t.Run("Error Action", func(t *testing.T) {
		input := map[string]interface{}{"value": "original"}

		executedCoreLogic := false
		result, err := plugins.ExecuteMixin("test.mixin.error", input, func(v map[string]interface{}) (interface{}, error) {
			executedCoreLogic = true
			return v["value"], nil
		})

		assert.Error(t, err)
		assert.False(t, executedCoreLogic, "Core logic should NOT have executed")
		assert.Nil(t, result)
		
		var mixinErr *plugins.MixinError
		assert.True(t, errors.As(err, &mixinErr))
		assert.Equal(t, "plugin manually threw an error", mixinErr.Message)
	})
}
