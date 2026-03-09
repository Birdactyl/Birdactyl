package tests

import (
	"encoding/json"
	"testing"

	"birdactyl-panel-backend/internal/models"
)

func TestHasPermission(t *testing.T) {
	t.Run("Exact Match", func(t *testing.T) {
		perms := []string{models.PermFileRead, models.PermFileWrite}
		if !models.HasPermission(perms, models.PermFileRead) {
			t.Errorf("Expected to have %s permission", models.PermFileRead)
		}
	})

	t.Run("Missing Permission", func(t *testing.T) {
		perms := []string{models.PermFileRead}
		if models.HasPermission(perms, models.PermFileWrite) {
			t.Errorf("Expected not to have %s permission", models.PermFileWrite)
		}
	})

	t.Run("Admin Wildcard", func(t *testing.T) {
		perms := []string{models.PermAdmin}
		if !models.HasPermission(perms, models.PermPowerStart) {
			t.Errorf("Expected wildcard admin to have %s permission", models.PermPowerStart)
		}
	})
}

func TestHasAnyPermission(t *testing.T) {
	t.Run("Single Match", func(t *testing.T) {
		perms := []string{models.PermConsoleRead, models.PermFileList}
		required := []string{models.PermFileWrite, models.PermConsoleRead}
		if !models.HasAnyPermission(perms, required) {
			t.Errorf("Expected to match one of the required permissions")
		}
	})

	t.Run("No Match", func(t *testing.T) {
		perms := []string{models.PermConsoleRead}
		required := []string{models.PermFileWrite, models.PermFileList}
		if models.HasAnyPermission(perms, required) {
			t.Errorf("Expected not to match any of the required permissions")
		}
	})
}

func TestSubuserPermissions(t *testing.T) {
	perms := []string{models.PermPowerStart, models.PermConsoleRead}
	permsJSON, _ := json.Marshal(perms)
	
	subuser := &models.Subuser{
		Permissions: permsJSON,
	}

	t.Run("GetPermissions", func(t *testing.T) {
		parsed := subuser.GetPermissions()
		if len(parsed) != 2 {
			t.Errorf("Expected 2 permissions, got %d", len(parsed))
		}
		if parsed[0] != models.PermPowerStart || parsed[1] != models.PermConsoleRead {
			t.Errorf("Permissions parsed incorrectly: %v", parsed)
		}
	})

	t.Run("HasPermission Check", func(t *testing.T) {
		if !subuser.HasPermission(models.PermPowerStart) {
			t.Errorf("Subuser should have %s permission", models.PermPowerStart)
		}
		if subuser.HasPermission(models.PermFileList) {
			t.Errorf("Subuser should not have %s permission", models.PermFileList)
		}
	})
}
