import { getAccessToken, setAccessToken, setRefreshToken, getRefreshToken, clearTokens } from '../auth';
import { notify } from '../../components/feedback/Notification';

export const API_BASE = '/api/v1';

interface Notification { title: string; message: string; type?: 'error' | 'success' | 'info'; }
interface ApiError { code?: number; message?: string; retry_after?: number; }
interface ApiResponse<T = unknown> { success: boolean; data?: T & { tokens?: TokenPair }; error?: string | ApiError; tokens?: TokenPair; notifications?: Notification[]; }
interface TokenPair { access_token: string; refresh_token: string; }
export interface ParsedResponse<T = unknown> { success: boolean; data?: T; error?: string; rateLimited?: boolean; retryAfter?: number; hasNotifications?: boolean; }

function parseResponse<T>(data: ApiResponse<T>): ParsedResponse<T> {
  let hasNotifications = false;
  if (data.notifications && data.notifications.length > 0) {
    hasNotifications = true;
    for (const n of data.notifications) {
      notify(n.title, n.message, n.type || 'info');
    }
  }
  const result: ParsedResponse<T> = { success: data.success, data: data.data, hasNotifications };
  if (data.error) {
    result.error = typeof data.error === 'string' ? data.error : data.error.message || 'Something went wrong';
    if (typeof data.error !== 'string' && data.error.code === 429) {
      result.rateLimited = true;
      result.retryAfter = data.error.retry_after;
    }
  }
  return result;
}

let isRefreshing = false;
let refreshPromise: Promise<ParsedResponse<unknown>> | null = null;

async function doRefresh(): Promise<ParsedResponse<unknown>> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) { clearTokens(); return { success: false, error: 'No refresh token' }; }
  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refresh_token: refreshToken }) });
    const data: ApiResponse<unknown> = await response.json();
    const tokenData = data.data as { access_token?: string; refresh_token?: string } | undefined;
    if (data.success && tokenData?.access_token && tokenData?.refresh_token) {
      setAccessToken(tokenData.access_token);
      setRefreshToken(tokenData.refresh_token);
      return { success: true };
    }
  } catch {}
  clearTokens();
  return { success: false, error: 'Refresh failed' };
}

export async function request<T>(endpoint: string, options: RequestInit = {}, retry = true): Promise<ParsedResponse<T>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...options.headers as Record<string, string> };
  const token = getAccessToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });
    if (response.status === 401 && retry && getRefreshToken()) {
      if (!isRefreshing) { isRefreshing = true; refreshPromise = doRefresh(); }
      const refreshResult = await refreshPromise!;
      isRefreshing = false;
      refreshPromise = null;
      if (refreshResult.success) return request<T>(endpoint, options, false);
      return { success: false, error: 'Session expired' };
    }
    const data: ApiResponse<T> = await response.json();
    const tokens = data.tokens || (data.data as { tokens?: TokenPair })?.tokens;
    if (tokens) { setAccessToken(tokens.access_token); setRefreshToken(tokens.refresh_token); }
    return parseResponse(data);
  } catch {
    return { success: false, error: 'Could not connect to server' };
  }
}

export const api = {
  get: <T>(endpoint: string) => request<T>(endpoint),
  post: <T>(endpoint: string, body?: unknown) => request<T>(endpoint, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(endpoint: string, body?: unknown) => request<T>(endpoint, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined }),
  put: <T>(endpoint: string, body?: unknown) => request<T>(endpoint, { method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
  delete: <T>(endpoint: string, body?: unknown) => request<T>(endpoint, { method: 'DELETE', body: body ? JSON.stringify(body) : undefined }),
};
