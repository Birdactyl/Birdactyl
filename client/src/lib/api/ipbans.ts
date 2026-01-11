import { api } from './client';

export interface IPBan {
  id: number;
  ip: string;
  reason: string;
  banned_by: string;
  created_at: string;
}

export interface PaginatedIPBans {
  bans: IPBan[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export const adminGetIPBans = (page = 1, perPage = 20, search = '') => {
  const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
  if (search) params.set('search', search);
  return api.get<PaginatedIPBans>(`/admin/ip-bans?${params}`);
};

export const adminCreateIPBan = (ip: string, reason: string) => api.post<IPBan>('/admin/ip-bans', { ip, reason });

export const adminDeleteIPBan = (id: number) => api.delete(`/admin/ip-bans/${id}`);
