import { api } from './client';
import type { Server } from './servers';
import type { Package } from './packages';

export interface PaginatedUsers {
  users: { id: string; username: string; email: string; is_admin: boolean; is_banned: boolean; is_root_admin: boolean; force_password_reset: boolean; ram_limit: number | null; cpu_limit: number | null; disk_limit: number | null; server_limit: number | null; created_at: string }[];
  page: number; per_page: number; total: number; total_pages: number; admin_count: number;
}

export interface Node {
  id: string; name: string; fqdn: string; port: number; is_online: boolean; auth_error: boolean; last_heartbeat: string | null; icon?: string;
  system_info: { hostname: string; os: { name: string; version: string; kernel: string; arch: string }; cpu: { cores: number; usage_percent: number }; memory: { total_bytes: number; used_bytes: number; available_bytes: number; usage_percent: number }; disk: { total_bytes: number; used_bytes: number; available_bytes: number; usage_percent: number }; uptime_seconds: number };
  created_at: string;
}

export interface NodeToken { token_id: string; token: string; }

export const adminGetUsers = (page = 1, perPage = 20, search = '', filter = 'all') => {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
  if (search) params.set('search', search);
  if (filter !== 'all') params.set('filter', filter);
  return api.get<PaginatedUsers>(`/admin/users?${params}`);
};
export const adminCreateUser = (email: string, username: string, password: string) => api.post<{ id: string; username: string; email: string }>('/admin/users', { email, username, password });
export const adminBanUsers = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/ban', { user_ids: userIds });
export const adminUnbanUsers = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/unban', { user_ids: userIds });
export const adminSetAdmin = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/set-admin', { user_ids: userIds });
export const adminRevokeAdmin = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/revoke-admin', { user_ids: userIds });
export const adminForcePasswordReset = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/force-reset', { user_ids: userIds });
export const adminDeleteUsers = (userIds: string[]) => api.post<{ affected: number }>('/admin/users/delete', { user_ids: userIds });
export const adminUpdateUser = (userId: string, data: { email?: string; username?: string; password?: string; ram_limit?: number | null; cpu_limit?: number | null; disk_limit?: number | null; server_limit?: number | null }) => api.patch(`/admin/users/${userId}`, data);

export const adminGetNodes = () => api.get<Node[]>('/admin/nodes');
export const adminRefreshNodes = () => api.post<Node[]>('/admin/nodes/refresh');
export const adminCreateNode = (name: string, fqdn: string, port = 8443, icon?: string) => api.post<{ node: Node; token: NodeToken }>('/admin/nodes', { name, fqdn, port, icon });
export const adminGetNode = (id: string) => api.get<Node>(`/admin/nodes/${id}`);
export const adminUpdateNode = (id: string, data: { name?: string; icon?: string }) => api.patch<Node>(`/admin/nodes/${id}`, data);
export const adminDeleteNode = (id: string) => api.delete(`/admin/nodes/${id}`);
export const adminResetNodeToken = (id: string) => api.post<NodeToken>(`/admin/nodes/${id}/reset-token`);
export const adminGetPairingCode = () => api.get<{ code: string }>('/admin/nodes/pairing-code');
export const adminPairNode = (name: string, fqdn: string, port: number, code: string) => api.post<{ node: Node; token: NodeToken }>('/admin/nodes/pair', { name, fqdn, port, code });

export const adminGetServers = () => api.get<Server[]>('/admin/servers');
export const adminCreateServer = (data: { name: string; node_id: string; package_id: string; memory: number; cpu: number; disk: number; user_id?: string }) => api.post<Server>('/admin/servers', data);
export const adminSuspendServers = (serverIds: string[]) => api.post<{ affected: number }>('/admin/servers/suspend', { server_ids: serverIds });
export const adminUnsuspendServers = (serverIds: string[]) => api.post<{ affected: number }>('/admin/servers/unsuspend', { server_ids: serverIds });
export const adminDeleteServers = (serverIds: string[]) => api.post<{ affected: number }>('/admin/servers/delete', { server_ids: serverIds });
export const adminUpdateServerResources = (serverId: string, data: { name?: string; user_id?: string; memory?: number; cpu?: number; disk?: number }) => api.patch<Server>(`/admin/servers/${serverId}/resources`, data);
export const adminTransferServer = (serverId: string, targetNodeId: string) => api.post<{ transfer_id: string }>(`/admin/servers/${serverId}/transfer`, { target_node_id: targetNodeId });
export const adminGetTransferStatus = (transferId: string) => api.get<TransferStatus>(`/admin/transfers/${transferId}`);
export const adminGetAllTransfers = () => api.get<TransferStatus[]>('/admin/transfers');
export const adminViewServer = (serverId: string) => api.post<Server>(`/admin/servers/${serverId}/view`, {});

export interface TransferStatus {
  id: string;
  server_id: string;
  server_name: string;
  from_node_id: string;
  from_node_name: string;
  to_node_id: string;
  to_node_name: string;
  stage: 'pending' | 'stopping' | 'archiving' | 'downloading' | 'uploading' | 'importing' | 'cleanup' | 'complete' | 'failed';
  progress: number;
  error?: string;
  started_at: string;
  completed_at?: string;
}

export const adminGetPackages = () => api.get<Package[]>('/admin/packages');
export const adminCreatePackage = (pkg: Partial<Package>) => api.post<Package>('/admin/packages', pkg);
export const adminGetPackage = (id: string) => api.get<Package>(`/admin/packages/${id}`);
export const adminUpdatePackage = (id: string, pkg: Partial<Package>) => api.patch<Package>(`/admin/packages/${id}`, pkg);
export const adminDeletePackage = (id: string) => api.delete(`/admin/packages/${id}`);

export const adminGetRegistrationStatus = () => api.get<{ enabled: boolean }>('/admin/settings/registration');
export const adminSetRegistrationStatus = (enabled: boolean) => api.patch<{ enabled: boolean }>('/admin/settings/registration', { enabled });

export const adminGetServerCreationStatus = () => api.get<{ enabled: boolean }>('/admin/settings/server-creation');
export const adminSetServerCreationStatus = (enabled: boolean) => api.patch<{ enabled: boolean }>('/admin/settings/server-creation', { enabled });


export interface AdminDatabaseHost {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  max_databases: number;
  databases: number;
  created_at: string;
}

export interface HostDatabase {
  id: string;
  database_name: string;
  username: string;
  server_id: string;
  server_name: string;
  created_at: string;
}

export const adminGetDatabaseHosts = () => api.get<AdminDatabaseHost[]>('/admin/database-hosts');
export const adminCreateDatabaseHost = (data: { name: string; host: string; port: number; username: string; password: string; max_databases?: number }) => api.post<AdminDatabaseHost>('/admin/database-hosts', data);
export const adminUpdateDatabaseHost = (id: string, data: { name?: string; host?: string; port?: number; username?: string; password?: string; max_databases?: number }) => api.patch(`/admin/database-hosts/${id}`, data);
export const adminDeleteDatabaseHost = (id: string) => api.delete(`/admin/database-hosts/${id}`);
export const adminGetHostDatabases = (hostId: string) => api.get<HostDatabase[]>(`/admin/database-hosts/${hostId}/databases`);
export const adminDeleteHostDatabase = (hostId: string, dbId: string) => api.delete(`/admin/database-hosts/${hostId}/databases/${dbId}`);

export interface InstalledPlugin { id: string; name: string; address: string; online: boolean; }
export const adminGetPlugins = () => api.get<{ plugins: InstalledPlugin[] }>('/admin/plugins');
export const adminDeletePlugin = (filename: string) => api.delete(`/admin/plugins/file/${encodeURIComponent(filename)}`);

export { type APIKey, type APIKeyCreated } from './auth';
export const adminGetUserAPIKeys = (userId: string) => api.get<import('./auth').APIKey[]>(`/admin/users/${userId}/api-keys`);
export const adminCreateUserAPIKey = (userId: string, name: string, expiresIn?: number) => api.post<import('./auth').APIKeyCreated>(`/admin/users/${userId}/api-keys`, { name, expires_in: expiresIn });
export const adminDeleteUserAPIKey = (userId: string, keyId: string) => api.delete(`/admin/users/${userId}/api-keys/${keyId}`);

