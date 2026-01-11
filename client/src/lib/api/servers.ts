import { api } from './client';
import { eventBus } from '../eventBus';
import type { Package } from './packages';

export interface Server {
  id: string; name: string; description: string; user_id: string; node_id: string; package_id: string;
  status: 'installing' | 'running' | 'stopped' | 'suspended' | 'failed';
  is_suspended: boolean; memory: number; cpu: number; disk: number;
  startup: string; docker_image: string;
  ports: { port: number; primary?: boolean }[];
  variables: Record<string, string>;
  created_at: string; updated_at: string;
  user?: { id: string; username: string; email: string };
  node?: { id: string; name: string; fqdn: string; display_ip?: string };
  package?: Package;
}

export interface ServerStatusResponse {
  status: string;
  stats?: {
    memory: number;
    memory_limit: number;
    cpu: number;
    disk: number;
    network_rx: number;
    network_tx: number;
  };
}

export const getServers = () => api.get<Server[]>('/servers/');
export const getServer = (id: string) => api.get<Server>(`/servers/${id}`);
export const getServerStatus = (id: string) => api.get<ServerStatusResponse>(`/servers/${id}/status`);
export const getServerPermissions = (id: string) => api.get<string[]>(`/servers/${id}/permissions`);
export const createServer = (data: { name: string; description?: string; node_id: string; package_id: string; memory: number; cpu: number; disk: number; ports: { port: number; primary?: boolean }[]; variables: Record<string, string> }) => api.post<Server>('/servers/', data);

export const startServer = async (id: string) => {
  const result = await api.post(`/servers/${id}/start`);
  if (result.success) eventBus.emit('server:start', { serverId: id });
  return result;
};

export const stopServer = async (id: string) => {
  const result = await api.post(`/servers/${id}/stop`);
  if (result.success) eventBus.emit('server:stop', { serverId: id });
  return result;
};

export const restartServer = async (id: string) => {
  const result = await api.post(`/servers/${id}/restart`);
  if (result.success) eventBus.emit('server:restart', { serverId: id });
  return result;
};

export const killServer = async (id: string) => {
  const result = await api.post(`/servers/${id}/kill`);
  if (result.success) eventBus.emit('server:kill', { serverId: id });
  return result;
};

export const reinstallServer = (id: string) => api.post(`/servers/${id}/reinstall`);
export const deleteServer = (id: string) => api.delete(`/servers/${id}`);
export const addAllocation = (serverId: string) => api.post<Server>(`/servers/${serverId}/allocations`);
export const setPrimaryAllocation = (serverId: string, port: number) => api.put<Server>(`/servers/${serverId}/allocations/primary`, { port });
export const deleteAllocation = (serverId: string, port: number) => api.delete<Server>(`/servers/${serverId}/allocations`, { port });
export const updateServerResources = (serverId: string, memory: number, cpu: number, disk: number) => api.patch<Server>(`/servers/${serverId}/resources`, { memory, cpu, disk });
export const updateServerName = (serverId: string, name?: string, description?: string) => api.patch<Server>(`/servers/${serverId}/name`, { name, description });
export const updateServerVariables = (serverId: string, variables: Record<string, string>, startup: string, dockerImage: string) => api.patch<Server>(`/servers/${serverId}/variables`, { variables, startup, docker_image: dockerImage });
