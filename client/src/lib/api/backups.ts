import { api, API_BASE } from './client';

export interface Backup {
  id: string;
  name: string;
  size: number;
  created_at: number;
  completed: boolean;
}

export const listBackups = (serverId: string) => api.get<Backup[]>(`/servers/${serverId}/backups`);
export const createBackup = (serverId: string, name?: string) => api.post<Backup>(`/servers/${serverId}/backups`, { name });
export const deleteBackup = (serverId: string, backupId: string) => api.delete(`/servers/${serverId}/backups/${backupId}`);
export const restoreBackup = (serverId: string, backupId: string) => api.post(`/servers/${serverId}/backups/${backupId}/restore`);
export const getBackupDownloadUrl = (serverId: string, backupId: string) => `${API_BASE}/servers/${serverId}/backups/${backupId}/download`;
