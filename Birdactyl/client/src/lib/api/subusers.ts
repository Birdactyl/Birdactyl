import { api } from './client';

export interface Subuser {
  id: string;
  server_id: string;
  user_id: string;
  permissions: string[];
  created_at: string;
  updated_at: string;
  user?: { id: string; username: string; email: string };
}

export const getSubusers = (serverId: string) => api.get<Subuser[]>(`/servers/${serverId}/subusers`);
export const addSubuser = (serverId: string, email: string, permissions: string[]) => api.post<Subuser>(`/servers/${serverId}/subusers`, { email, permissions });
export const updateSubuser = (serverId: string, subuserId: string, permissions: string[]) => api.patch<Subuser>(`/servers/${serverId}/subusers/${subuserId}`, { permissions });
export const removeSubuser = (serverId: string, subuserId: string) => api.delete(`/servers/${serverId}/subusers/${subuserId}`);
