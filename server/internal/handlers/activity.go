package handlers

import (
	"encoding/json"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	ActionAuthRegister  = "auth.register"
	ActionAuthLogin     = "auth.login"
	ActionAuthLogout    = "auth.logout"
	ActionAuthLogoutAll = "auth.logout_all"

	ActionProfileUpdate         = "profile.update"
	ActionProfilePasswordChange = "profile.password_change"
	ActionProfileSessionRevoke  = "profile.session_revoke"
	ActionProfileSessionsRevoke = "profile.sessions_revoke_all"

	ActionServerCreate    = "server.create"
	ActionServerDelete    = "server.delete"
	ActionServerStart     = "server.start"
	ActionServerStop      = "server.stop"
	ActionServerKill      = "server.kill"
	ActionServerRestart   = "server.restart"
	ActionServerReinstall = "server.reinstall"
	ActionServerCommand   = "server.command"

	ActionServerNameUpdate      = "server.name.update"
	ActionServerResourcesUpdate = "server.resources.update"
	ActionServerVariablesUpdate = "server.variables.update"
	ActionServerAllocationAdd   = "server.allocation.add"
	ActionServerAllocationPri   = "server.allocation.set_primary"
	ActionServerAllocationDel   = "server.allocation.delete"

	ActionFileCreateFolder  = "server.file.create_folder"
	ActionFileWrite         = "server.file.write"
	ActionFileUpload        = "server.file.upload"
	ActionFileDelete        = "server.file.delete"
	ActionFileMove          = "server.file.move"
	ActionFileCopy          = "server.file.copy"
	ActionFileCompress      = "server.file.compress"
	ActionFileDecompress    = "server.file.decompress"
	ActionFileBulkDelete    = "server.file.bulk_delete"
	ActionFileBulkCopy      = "server.file.bulk_copy"
	ActionFileBulkCompress  = "server.file.bulk_compress"

	ActionBackupCreate  = "server.backup.create"
	ActionBackupDelete  = "server.backup.delete"
	ActionBackupRestore = "server.backup.restore"

	ActionSubuserAdd    = "server.subuser.add"
	ActionSubuserUpdate = "server.subuser.update"
	ActionSubuserRemove = "server.subuser.remove"

	ActionAdminUserCreate     = "admin.user.create"
	ActionAdminUserUpdate     = "admin.user.update"
	ActionAdminUserDelete     = "admin.user.delete"
	ActionAdminUserBan        = "admin.user.ban"
	ActionAdminUserUnban      = "admin.user.unban"
	ActionAdminUserSetAdmin   = "admin.user.set_admin"
	ActionAdminUserRevokeAdm  = "admin.user.revoke_admin"
	ActionAdminUserForceReset = "admin.user.force_reset"

	ActionAdminServerCreate    = "admin.server.create"
	ActionAdminServerView      = "admin.server.view"
	ActionAdminServerSuspend   = "admin.server.suspend"
	ActionAdminServerUnsuspend = "admin.server.unsuspend"
	ActionAdminServerDelete    = "admin.server.delete"
	ActionAdminServerResources = "admin.server.resources"
	ActionAdminServerTransfer  = "admin.server.transfer"

	ActionAdminNodeCreate     = "admin.node.create"
	ActionAdminNodeUpdate     = "admin.node.update"
	ActionAdminNodeDelete     = "admin.node.delete"
	ActionAdminNodeResetToken = "admin.node.reset_token"

	ActionAdminPackageCreate = "admin.package.create"
	ActionAdminPackageUpdate = "admin.package.update"
	ActionAdminPackageDelete = "admin.package.delete"

	ActionAdminIPBanCreate = "admin.ipban.create"
	ActionAdminIPBanDelete = "admin.ipban.delete"

	ActionAdminSettingsRegistration   = "admin.settings.registration"
	ActionAdminSettingsServerCreation = "admin.settings.server_creation"

	ActionDatabaseCreate         = "server.database.create"
	ActionDatabaseDelete         = "server.database.delete"
	ActionDatabaseRotatePassword = "server.database.rotate_password"

	ActionAdminDbHostCreate = "admin.database_host.create"
	ActionAdminDbHostUpdate = "admin.database_host.update"
	ActionAdminDbHostDelete = "admin.database_host.delete"
	ActionAdminDbDelete     = "admin.database.delete"

	ActionAdminDBHostCreate   = "admin.database_host.create"
	ActionAdminDBHostUpdate   = "admin.database_host.update"
	ActionAdminDBHostDelete   = "admin.database_host.delete"
	ActionAdminSettingsUpdate = "admin.settings.update"

	ActionAllocationAdd        = "server.allocation.add"
	ActionAllocationDelete     = "server.allocation.delete"
	ActionAllocationSetPrimary = "server.allocation.set_primary"
)

func LogActivity(userID uuid.UUID, username, action, description, ip, userAgent string, isAdmin bool, metadata map[string]interface{}) {
	var metaStr string
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaStr = string(b)
		}
	}

	log := models.ActivityLog{
		UserID:      userID,
		Username:    username,
		Action:      action,
		Description: description,
		IP:          ip,
		UserAgent:   userAgent,
		IsAdmin:     isAdmin,
		Metadata:    metaStr,
	}
	database.DB.Create(&log)
}

func Log(c *fiber.Ctx, user *models.User, action, description string, metadata map[string]interface{}) {
	LogActivity(user.ID, user.Username, action, description, c.IP(), c.Get("User-Agent"), user.IsAdmin, metadata)
}
