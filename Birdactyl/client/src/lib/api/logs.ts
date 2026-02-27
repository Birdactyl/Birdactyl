import { api } from './client';

export interface ActivityLog {
  id: string;
  user_id: string;
  username: string;
  action: string;
  description: string;
  ip: string;
  user_agent: string;
  is_admin: boolean;
  metadata: string;
  created_at: string;
}

export interface PaginatedLogs {
  logs: ActivityLog[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export const adminGetLogs = (page = 1, perPage = 20, search = '', filter = 'all', from = '', to = '') => {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
  if (search) params.set('search', search);
  if (filter !== 'all') params.set('filter', filter);
  if (from) params.set('from', from);
  if (to) params.set('to', to);
  return api.get<PaginatedLogs>(`/admin/logs?${params}`);
};

export const getServerLogs = (serverId: string, page = 1, perPage = 20, search = '', from = '', to = '') => {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
  if (search) params.set('search', search);
  if (from) params.set('from', from);
  if (to) params.set('to', to);
  return api.get<PaginatedLogs>(`/servers/${serverId}/activity?${params}`);
};
