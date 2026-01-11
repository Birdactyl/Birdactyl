import { api } from './client';

export interface ServerDatabase {
  id: string;
  database_name: string;
  username: string;
  password: string;
  host: string;
  port: number;
  created_at: string;
}

export interface DatabaseHost {
  id: string;
  name: string;
  host: string;
  port: number;
  max_databases: number;
  used: number;
}

export const getServerDatabases = (serverId: string) => api.get<ServerDatabase[]>(`/servers/${serverId}/databases`);
export const getDatabaseHosts = (serverId: string) => api.get<DatabaseHost[]>(`/servers/${serverId}/databases/hosts`);
export const createServerDatabase = (serverId: string, hostId: string, name?: string) => api.post<ServerDatabase>(`/servers/${serverId}/databases`, { host_id: hostId, name });
export const deleteServerDatabase = (serverId: string, dbId: string) => api.delete(`/servers/${serverId}/databases/${dbId}`);
export const rotateDatabasePassword = (serverId: string, dbId: string) => api.post<ServerDatabase>(`/servers/${serverId}/databases/${dbId}/rotate`);
