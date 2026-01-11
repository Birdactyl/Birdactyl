package models

const (
	PermPowerStart   = "power.start"
	PermPowerStop    = "power.stop"
	PermPowerRestart = "power.restart"
	PermPowerKill    = "power.kill"

	PermConsoleRead  = "console.read"
	PermConsoleWrite = "console.write"

	PermFileList       = "file.list"
	PermFileRead       = "file.read"
	PermFileWrite      = "file.write"
	PermFileCreate     = "file.create"
	PermFileDelete     = "file.delete"
	PermFileMove       = "file.move"
	PermFileCopy       = "file.copy"
	PermFileCompress   = "file.compress"
	PermFileDecompress = "file.decompress"
	PermFileUpload     = "file.upload"
	PermFileDownload   = "file.download"

	PermBackupList     = "backup.list"
	PermBackupCreate   = "backup.create"
	PermBackupDelete   = "backup.delete"
	PermBackupDownload = "backup.download"
	PermBackupRestore  = "backup.restore"

	PermDatabaseView   = "database.view"
	PermDatabaseCreate = "database.create"
	PermDatabaseUpdate = "database.update"
	PermDatabaseDelete = "database.delete"

	PermScheduleList   = "schedule.list"
	PermScheduleCreate = "schedule.create"
	PermScheduleUpdate = "schedule.update"
	PermScheduleDelete = "schedule.delete"
	PermScheduleRun    = "schedule.run"

	PermAllocationView       = "allocation.view"
	PermAllocationAdd        = "allocation.add"
	PermAllocationDelete     = "allocation.delete"
	PermAllocationSetPrimary = "allocation.set_primary"

	PermSettingsView      = "settings.view"
	PermSettingsRename    = "settings.rename"
	PermSettingsResources = "settings.resources"

	PermStartupView   = "startup.view"
	PermStartupUpdate = "startup.update"

	PermReinstall = "server.reinstall"

	PermActivityView = "activity.view"

	PermAdmin = "*"
)

var AllPermissions = []string{
	PermPowerStart, PermPowerStop, PermPowerRestart, PermPowerKill,
	PermConsoleRead, PermConsoleWrite,
	PermFileList, PermFileRead, PermFileWrite, PermFileCreate, PermFileDelete,
	PermFileMove, PermFileCopy, PermFileCompress, PermFileDecompress,
	PermFileUpload, PermFileDownload,
	PermBackupList, PermBackupCreate, PermBackupDelete, PermBackupDownload, PermBackupRestore,
	PermDatabaseView, PermDatabaseCreate, PermDatabaseUpdate, PermDatabaseDelete,
	PermScheduleList, PermScheduleCreate, PermScheduleUpdate, PermScheduleDelete, PermScheduleRun,
	PermAllocationView, PermAllocationAdd, PermAllocationDelete, PermAllocationSetPrimary,
	PermSettingsView, PermSettingsRename, PermSettingsResources,
	PermStartupView, PermStartupUpdate,
	PermReinstall,
	PermActivityView,
}

var PermissionGroups = map[string][]string{
	"power":      {PermPowerStart, PermPowerStop, PermPowerRestart, PermPowerKill},
	"console":    {PermConsoleRead, PermConsoleWrite},
	"file":       {PermFileList, PermFileRead, PermFileWrite, PermFileCreate, PermFileDelete, PermFileMove, PermFileCopy, PermFileCompress, PermFileDecompress, PermFileUpload, PermFileDownload},
	"backup":     {PermBackupList, PermBackupCreate, PermBackupDelete, PermBackupDownload, PermBackupRestore},
	"database":   {PermDatabaseView, PermDatabaseCreate, PermDatabaseUpdate, PermDatabaseDelete},
	"schedule":   {PermScheduleList, PermScheduleCreate, PermScheduleUpdate, PermScheduleDelete, PermScheduleRun},
	"allocation": {PermAllocationView, PermAllocationAdd, PermAllocationDelete, PermAllocationSetPrimary},
	"settings":   {PermSettingsView, PermSettingsRename, PermSettingsResources},
	"startup":    {PermStartupView, PermStartupUpdate},
	"server":     {PermReinstall},
	"activity":   {PermActivityView},
}

func HasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == PermAdmin || p == required {
			return true
		}
	}
	return false
}

func HasAnyPermission(permissions []string, required []string) bool {
	for _, r := range required {
		if HasPermission(permissions, r) {
			return true
		}
	}
	return false
}

func OwnerPermissions() []string {
	return []string{PermAdmin}
}
