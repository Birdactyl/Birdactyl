package plugins

import "errors"

var (
	ErrDynamicDisabled = errors.New("dynamic plugin loading is disabled")
	ErrInvalidConfig   = errors.New("invalid plugin config: missing id or address")
	ErrPluginNotFound  = errors.New("plugin not found")
)

type MixinTarget string

const (
	MixinServerCreate    MixinTarget = "server.create"
	MixinServerUpdate    MixinTarget = "server.update"
	MixinServerDelete    MixinTarget = "server.delete"
	MixinServerStart     MixinTarget = "server.start"
	MixinServerStop      MixinTarget = "server.stop"
	MixinServerRestart   MixinTarget = "server.restart"
	MixinServerKill      MixinTarget = "server.kill"
	MixinServerSuspend   MixinTarget = "server.suspend"
	MixinServerUnsuspend MixinTarget = "server.unsuspend"
	MixinServerReinstall MixinTarget = "server.reinstall"
	MixinServerTransfer  MixinTarget = "server.transfer"
	MixinServerList      MixinTarget = "server.list"
	MixinServerGet       MixinTarget = "server.get"

	MixinUserCreate       MixinTarget = "user.create"
	MixinUserUpdate       MixinTarget = "user.update"
	MixinUserDelete       MixinTarget = "user.delete"
	MixinUserAuthenticate MixinTarget = "user.authenticate"
	MixinUserBan          MixinTarget = "user.ban"
	MixinUserUnban        MixinTarget = "user.unban"
	MixinUserList         MixinTarget = "user.list"
	MixinUserGet          MixinTarget = "user.get"

	MixinDatabaseCreate MixinTarget = "database.create"
	MixinDatabaseDelete MixinTarget = "database.delete"
	MixinDatabaseList   MixinTarget = "database.list"

	MixinBackupCreate  MixinTarget = "backup.create"
	MixinBackupDelete  MixinTarget = "backup.delete"
	MixinBackupList    MixinTarget = "backup.list"
	MixinBackupRestore MixinTarget = "backup.restore"

	MixinFileRead       MixinTarget = "file.read"
	MixinFileWrite      MixinTarget = "file.write"
	MixinFileDelete     MixinTarget = "file.delete"
	MixinFileUpload     MixinTarget = "file.upload"
	MixinFileMove       MixinTarget = "file.move"
	MixinFileCopy       MixinTarget = "file.copy"
	MixinFileCompress   MixinTarget = "file.compress"
	MixinFileDecompress MixinTarget = "file.decompress"
	MixinFileList       MixinTarget = "file.list"

	MixinNodeCreate MixinTarget = "node.create"
	MixinNodeDelete MixinTarget = "node.delete"
	MixinNodeList   MixinTarget = "node.list"
	MixinNodeGet    MixinTarget = "node.get"

	MixinPackageCreate MixinTarget = "package.create"
	MixinPackageUpdate MixinTarget = "package.update"
	MixinPackageDelete MixinTarget = "package.delete"
	MixinPackageList   MixinTarget = "package.list"
	MixinPackageGet    MixinTarget = "package.get"

	MixinSubuserAdd    MixinTarget = "subuser.add"
	MixinSubuserUpdate MixinTarget = "subuser.update"
	MixinSubuserRemove MixinTarget = "subuser.remove"
	MixinSubuserList   MixinTarget = "subuser.list"

	MixinIPBanCreate MixinTarget = "ipban.create"
	MixinIPBanDelete MixinTarget = "ipban.delete"
	MixinIPBanList   MixinTarget = "ipban.list"

	MixinAllocationAdd        MixinTarget = "allocation.add"
	MixinAllocationDelete     MixinTarget = "allocation.delete"
	MixinAllocationSetPrimary MixinTarget = "allocation.set_primary"
	MixinAllocationList       MixinTarget = "allocation.list"

	MixinDBHostCreate MixinTarget = "dbhost.create"
	MixinDBHostUpdate MixinTarget = "dbhost.update"
	MixinDBHostDelete MixinTarget = "dbhost.delete"
	MixinDBHostList   MixinTarget = "dbhost.list"

	MixinSettingsUpdate MixinTarget = "settings.update"
	MixinSettingsGet    MixinTarget = "settings.get"

	MixinActivityLogList MixinTarget = "activitylog.list"

	MixinConsoleCommand MixinTarget = "console.command"
)

type EventType string

const (
	EventServerCreating   EventType = "server.creating"
	EventServerCreated    EventType = "server.created"
	EventServerDeleting   EventType = "server.deleting"
	EventServerDeleted    EventType = "server.deleted"
	EventServerStarting   EventType = "server.starting"
	EventServerStarted    EventType = "server.started"
	EventServerStopping   EventType = "server.stopping"
	EventServerStopped    EventType = "server.stopped"
	EventServerRestarting EventType = "server.restarting"
	EventServerKilling    EventType = "server.killing"
	EventServerReinstall  EventType = "server.reinstalling"
	EventServerSuspended  EventType = "server.suspended"
	EventServerUnsuspended EventType = "server.unsuspended"
	EventServerUpdated    EventType = "server.updated"
	EventServerTransferred EventType = "server.transferred"

	EventUserRegistering EventType = "user.registering"
	EventUserRegistered  EventType = "user.registered"
	EventUserLoggingIn   EventType = "user.logging_in"
	EventUserLoggedIn    EventType = "user.logged_in"
	EventUserLogout      EventType = "user.logout"
	EventUserBanned      EventType = "user.banned"
	EventUserUnbanned    EventType = "user.unbanned"
	EventUserUpdated     EventType = "user.updated"
	EventUserDeleted     EventType = "user.deleted"

	EventDatabaseCreating EventType = "database.creating"
	EventDatabaseCreated  EventType = "database.created"
	EventDatabaseDeleting EventType = "database.deleting"
	EventDatabaseDeleted  EventType = "database.deleted"

	EventBackupCreating  EventType = "backup.creating"
	EventBackupCreated   EventType = "backup.created"
	EventBackupDeleting  EventType = "backup.deleting"
	EventBackupDeleted   EventType = "backup.deleted"
	EventBackupRestoring EventType = "backup.restoring"
	EventBackupRestored  EventType = "backup.restored"

	EventFileUploading EventType = "file.uploading"
	EventFileUploaded  EventType = "file.uploaded"
	EventFileDeleting  EventType = "file.deleting"
	EventFileDeleted   EventType = "file.deleted"
	EventFileWriting   EventType = "file.writing"
	EventFileWritten   EventType = "file.written"

	EventSubuserAdding   EventType = "subuser.adding"
	EventSubuserAdded    EventType = "subuser.added"
	EventSubuserRemoving EventType = "subuser.removing"
	EventSubuserRemoved  EventType = "subuser.removed"

	EventNodeCreated  EventType = "node.created"
	EventNodeDeleted  EventType = "node.deleted"
	EventNodeOnline   EventType = "node.online"
	EventNodeOffline  EventType = "node.offline"

	EventPackageCreated EventType = "package.created"
	EventPackageUpdated EventType = "package.updated"
	EventPackageDeleted EventType = "package.deleted"

	EventIPBanCreated EventType = "ipban.created"
	EventIPBanDeleted EventType = "ipban.deleted"

	EventSettingsUpdated EventType = "settings.updated"

	EventSystemStartup  EventType = "system.startup"
	EventSystemShutdown EventType = "system.shutdown"

	EventPluginLoaded   EventType = "plugin.loaded"
	EventPluginUnloaded EventType = "plugin.unloaded"
)

var SyncEvents = map[EventType]bool{
	EventServerCreating:   true,
	EventServerDeleting:   true,
	EventServerStarting:   true,
	EventServerStopping:   true,
	EventServerRestarting: true,
	EventServerKilling:    true,
	EventServerReinstall:  true,
	EventUserRegistering:  true,
	EventUserLoggingIn:    true,
	EventDatabaseCreating: true,
	EventDatabaseDeleting: true,
	EventBackupCreating:   true,
	EventBackupDeleting:   true,
	EventBackupRestoring:  true,
	EventFileUploading:    true,
	EventFileDeleting:     true,
	EventFileWriting:      true,
	EventSubuserAdding:    true,
	EventSubuserRemoving:  true,
}

type Permission string

const (
	PermServerRead    Permission = "server.read"
	PermServerWrite   Permission = "server.write"
	PermServerCommand Permission = "server.command"
	PermServerManage  Permission = "server.manage"
	PermUserRead      Permission = "user.read"
	PermUserWrite     Permission = "user.write"
	PermFileRead      Permission = "file.read"
	PermFileWrite     Permission = "file.write"
	PermDatabaseRead  Permission = "database.read"
	PermDatabaseWrite Permission = "database.write"
	PermBackupRead    Permission = "backup.read"
	PermBackupWrite   Permission = "backup.write"
	PermLog           Permission = "log"
	PermAdmin         Permission = "admin"
)

type PluginConfig struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Address     string       `yaml:"address"`
	Token       string       `yaml:"token"`
	Binary      string       `yaml:"binary"`
	Events      []EventType  `yaml:"events"`
	Permissions []Permission `yaml:"permissions"`
}
