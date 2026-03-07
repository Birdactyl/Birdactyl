package benchmarks

import (
	"encoding/json"
	"testing"

	"birdactyl-panel-backend/internal/models"
)

func BenchmarkHasPermission_Hit(b *testing.B) {
	perms := models.AllPermissions
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasPermission(perms, models.PermFileRead)
	}
}

func BenchmarkHasPermission_Miss(b *testing.B) {
	perms := []string{models.PermPowerStart, models.PermPowerStop, models.PermConsoleRead}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasPermission(perms, models.PermFileDelete)
	}
}

func BenchmarkHasPermission_Wildcard(b *testing.B) {
	perms := []string{models.PermAdmin}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasPermission(perms, models.PermFileDelete)
	}
}

func BenchmarkHasPermission_LastElement(b *testing.B) {
	perms := models.AllPermissions
	last := perms[len(perms)-1]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasPermission(perms, last)
	}
}

func BenchmarkHasAnyPermission_SingleMatch(b *testing.B) {
	perms := models.AllPermissions
	required := []string{models.PermConsoleRead}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasAnyPermission(perms, required)
	}
}

func BenchmarkHasAnyPermission_MultipleRequired(b *testing.B) {
	perms := []string{models.PermPowerStart, models.PermConsoleRead, models.PermFileList}
	required := []string{
		models.PermBackupCreate,
		models.PermDatabaseView,
		models.PermScheduleList,
		models.PermFileList,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasAnyPermission(perms, required)
	}
}

func BenchmarkHasAnyPermission_NoMatch(b *testing.B) {
	perms := []string{models.PermPowerStart, models.PermPowerStop}
	required := []string{
		models.PermFileRead,
		models.PermFileWrite,
		models.PermBackupCreate,
		models.PermDatabaseView,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.HasAnyPermission(perms, required)
	}
}

func BenchmarkSubuserGetPermissions(b *testing.B) {
	perms := []string{models.PermPowerStart, models.PermConsoleRead, models.PermFileList, models.PermFileRead}
	permsJSON, _ := json.Marshal(perms)
	subuser := &models.Subuser{Permissions: permsJSON}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subuser.GetPermissions()
	}
}

func BenchmarkSubuserGetPermissions_Full(b *testing.B) {
	permsJSON, _ := json.Marshal(models.AllPermissions)
	subuser := &models.Subuser{Permissions: permsJSON}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subuser.GetPermissions()
	}
}

func BenchmarkSubuserHasPermission(b *testing.B) {
	perms := []string{models.PermPowerStart, models.PermConsoleRead, models.PermFileList, models.PermFileRead}
	permsJSON, _ := json.Marshal(perms)
	subuser := &models.Subuser{Permissions: permsJSON}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subuser.HasPermission(models.PermFileRead)
	}
}

func BenchmarkOwnerPermissions(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.OwnerPermissions()
	}
}
