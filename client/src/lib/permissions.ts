export const Permissions = {
  POWER_START: 'power.start',
  POWER_STOP: 'power.stop',
  POWER_RESTART: 'power.restart',
  POWER_KILL: 'power.kill',

  CONSOLE_READ: 'console.read',
  CONSOLE_WRITE: 'console.write',

  FILE_LIST: 'file.list',
  FILE_READ: 'file.read',
  FILE_WRITE: 'file.write',
  FILE_CREATE: 'file.create',
  FILE_DELETE: 'file.delete',
  FILE_MOVE: 'file.move',
  FILE_COPY: 'file.copy',
  FILE_COMPRESS: 'file.compress',
  FILE_DECOMPRESS: 'file.decompress',
  FILE_UPLOAD: 'file.upload',
  FILE_DOWNLOAD: 'file.download',

  BACKUP_LIST: 'backup.list',
  BACKUP_CREATE: 'backup.create',
  BACKUP_DELETE: 'backup.delete',
  BACKUP_DOWNLOAD: 'backup.download',

  SCHEDULE_LIST: 'schedule.list',
  SCHEDULE_CREATE: 'schedule.create',
  SCHEDULE_UPDATE: 'schedule.update',
  SCHEDULE_DELETE: 'schedule.delete',
  SCHEDULE_RUN: 'schedule.run',

  ALLOCATION_VIEW: 'allocation.view',
  ALLOCATION_ADD: 'allocation.add',
  ALLOCATION_DELETE: 'allocation.delete',
  ALLOCATION_SET_PRIMARY: 'allocation.set_primary',

  SETTINGS_VIEW: 'settings.view',
  SETTINGS_RENAME: 'settings.rename',
  SETTINGS_RESOURCES: 'settings.resources',

  REINSTALL: 'server.reinstall',

  ACTIVITY_VIEW: 'activity.view',

  ADMIN: '*',
} as const;

export type Permission = typeof Permissions[keyof typeof Permissions];

export const PermissionGroups = {
  power: [Permissions.POWER_START, Permissions.POWER_STOP, Permissions.POWER_RESTART, Permissions.POWER_KILL],
  console: [Permissions.CONSOLE_READ, Permissions.CONSOLE_WRITE],
  file: [Permissions.FILE_LIST, Permissions.FILE_READ, Permissions.FILE_WRITE, Permissions.FILE_CREATE, Permissions.FILE_DELETE, Permissions.FILE_MOVE, Permissions.FILE_COPY, Permissions.FILE_COMPRESS, Permissions.FILE_DECOMPRESS, Permissions.FILE_UPLOAD, Permissions.FILE_DOWNLOAD],
  backup: [Permissions.BACKUP_LIST, Permissions.BACKUP_CREATE, Permissions.BACKUP_DELETE, Permissions.BACKUP_DOWNLOAD],
  schedule: [Permissions.SCHEDULE_LIST, Permissions.SCHEDULE_CREATE, Permissions.SCHEDULE_UPDATE, Permissions.SCHEDULE_DELETE, Permissions.SCHEDULE_RUN],
  allocation: [Permissions.ALLOCATION_VIEW, Permissions.ALLOCATION_ADD, Permissions.ALLOCATION_DELETE, Permissions.ALLOCATION_SET_PRIMARY],
  settings: [Permissions.SETTINGS_VIEW, Permissions.SETTINGS_RENAME, Permissions.SETTINGS_RESOURCES],
  server: [Permissions.REINSTALL],
  activity: [Permissions.ACTIVITY_VIEW],
};

export const PermissionLabels: Record<string, string> = {
  [Permissions.POWER_START]: 'Start Server',
  [Permissions.POWER_STOP]: 'Stop Server',
  [Permissions.POWER_RESTART]: 'Restart Server',
  [Permissions.POWER_KILL]: 'Kill Server',
  [Permissions.CONSOLE_READ]: 'View Console',
  [Permissions.CONSOLE_WRITE]: 'Send Commands',
  [Permissions.FILE_LIST]: 'List Files',
  [Permissions.FILE_READ]: 'Read Files',
  [Permissions.FILE_WRITE]: 'Write Files',
  [Permissions.FILE_CREATE]: 'Create Files/Folders',
  [Permissions.FILE_DELETE]: 'Delete Files',
  [Permissions.FILE_MOVE]: 'Move/Rename Files',
  [Permissions.FILE_COPY]: 'Copy Files',
  [Permissions.FILE_COMPRESS]: 'Compress Files',
  [Permissions.FILE_DECOMPRESS]: 'Decompress Files',
  [Permissions.FILE_UPLOAD]: 'Upload Files',
  [Permissions.FILE_DOWNLOAD]: 'Download Files',
  [Permissions.BACKUP_LIST]: 'List Backups',
  [Permissions.BACKUP_CREATE]: 'Create Backups',
  [Permissions.BACKUP_DELETE]: 'Delete Backups',
  [Permissions.BACKUP_DOWNLOAD]: 'Download Backups',
  [Permissions.SCHEDULE_LIST]: 'View Schedules',
  [Permissions.SCHEDULE_CREATE]: 'Create Schedules',
  [Permissions.SCHEDULE_UPDATE]: 'Edit Schedules',
  [Permissions.SCHEDULE_DELETE]: 'Delete Schedules',
  [Permissions.SCHEDULE_RUN]: 'Run Schedules',
  [Permissions.ALLOCATION_VIEW]: 'View Allocations',
  [Permissions.ALLOCATION_ADD]: 'Add Allocations',
  [Permissions.ALLOCATION_DELETE]: 'Delete Allocations',
  [Permissions.ALLOCATION_SET_PRIMARY]: 'Set Primary Allocation',
  [Permissions.SETTINGS_VIEW]: 'View Settings',
  [Permissions.SETTINGS_RENAME]: 'Rename Server',
  [Permissions.SETTINGS_RESOURCES]: 'Update Resources',
  [Permissions.REINSTALL]: 'Reinstall Server',
  [Permissions.ACTIVITY_VIEW]: 'View Activity Log',
};

export function hasPermission(permissions: string[], required: string): boolean {
  return permissions.includes(Permissions.ADMIN) || permissions.includes(required);
}

export function hasAnyPermission(permissions: string[], required: string[]): boolean {
  return required.some(r => hasPermission(permissions, r));
}
