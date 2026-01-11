import { api } from './client';
import type { AddonSource } from './packages';

export interface Addon {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  author?: string;
  downloads?: number;
}

export interface AddonVersion {
  id: string;
  name: string;
  download_url: string | null;
  file_name?: string;
  mod_id?: string;
}

export interface InstalledAddon {
  name: string;
  size: number;
  is_dir: boolean;
  mod_time: number;
  mode: string;
  icon?: string;
  project_id?: string;
}

export const getAddonSources = (serverId: string) =>
  api.get<AddonSource[]>(`/servers/${serverId}/addons/sources`);

export const searchAddons = (serverId: string, sourceId: string, query: string, limit = 20, offset = 0) =>
  api.get<Addon[]>(`/servers/${serverId}/addons/search?source=${sourceId}&q=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`);

export const getAddonVersions = (serverId: string, sourceId: string, addonId: string) =>
  api.get<AddonVersion[]>(`/servers/${serverId}/addons/versions?source=${sourceId}&addon=${encodeURIComponent(addonId)}`);

export const listInstalledAddons = (serverId: string, sourceId: string) =>
  api.get<InstalledAddon[]>(`/servers/${serverId}/addons/installed?source=${sourceId}`);

export const installAddon = (serverId: string, sourceId: string, downloadUrl: string | null, fileName: string, modId?: string, fileId?: string) =>
  api.post(`/servers/${serverId}/addons/install`, { source_id: sourceId, download_url: downloadUrl || '', file_name: fileName, mod_id: modId, file_id: fileId });

export const deleteAddon = (serverId: string, sourceId: string, fileName: string) =>
  api.delete(`/servers/${serverId}/addons`, { source_id: sourceId, file_name: fileName });

export interface Modpack {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  author?: string;
  downloads?: number;
}

export interface ModpackVersion {
  id: string;
  name: string;
  download_url: string;
}

export interface ModpackInstallResult {
  name: string;
  files_installed: number;
  files_failed: number;
  failed_files?: string[];
}

export const searchModpacks = (serverId: string, query: string, loader = 'fabric', limit = 20, offset = 0) =>
  api.get<Modpack[]>(`/servers/${serverId}/modpacks/search?q=${encodeURIComponent(query)}&loader=${loader}&limit=${limit}&offset=${offset}`);

export const getModpackVersions = (serverId: string, modpackId: string, loader = 'fabric') =>
  api.get<ModpackVersion[]>(`/servers/${serverId}/modpacks/versions?id=${encodeURIComponent(modpackId)}&loader=${loader}`);

export const installModpack = (serverId: string, downloadUrl: string, sourceId?: string) =>
  api.post<ModpackInstallResult>(`/servers/${serverId}/modpacks/install`, { download_url: downloadUrl, source_id: sourceId || '' });
