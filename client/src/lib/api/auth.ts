import { api } from './client';
import { getRefreshToken, clearTokens } from '../auth';

export interface Session { id: string; ip: string; user_agent: string; created_at: string; expires_at: string; is_current: boolean; }
export interface User { id: string; username: string; email: string; is_admin: boolean; }
export interface Resources {
  enabled: boolean;
  limits: { ram: number; cpu: number; disk: number; servers: number };
  used: { ram: number; cpu: number; disk: number; servers: number };
}


export const register = (email: string, username: string, password: string) => api.post('/auth/register', { email, username, password });
export const login = (email: string, password: string) => api.post('/auth/login', { email, password });
export const refresh = () => api.post('/auth/refresh', { refresh_token: getRefreshToken() });
export const logout = async () => { const r = await api.post('/auth/logout'); clearTokens(); return r; };
export const getMe = () => api.get<User>('/auth/me');
export const getResources = () => api.get<Resources>('/auth/resources');
export const updateProfile = (username: string, email: string, totpCode?: string) => api.patch<{ id: string; username: string; email: string }>('/auth/profile', { username, email, totp_code: totpCode });
export const sendEmailChangeCode = () => api.post('/auth/profile/email-change-code', {});
export const updatePassword = (currentPassword: string, newPassword: string) => api.patch('/auth/password', { current_password: currentPassword, new_password: newPassword });
export const getSessions = () => api.get<Session[]>('/auth/sessions');
export const revokeSession = (sessionId: string) => api.delete(`/auth/sessions/${sessionId}`);
export const revokeAllSessions = () => api.delete('/auth/sessions');

export interface APIKey { id: string; name: string; key_prefix: string; expires_at: string | null; last_used_at: string | null; created_at: string; }
export interface APIKeyCreated extends APIKey { key: string; }
export const getAPIKeys = () => api.get<APIKey[]>('/auth/api-keys');
export const createAPIKey = (name: string, expiresIn?: number) => api.post<APIKeyCreated>('/auth/api-keys', { name, expires_in: expiresIn });
export const deleteAPIKey = (id: string) => api.delete(`/auth/api-keys/${id}`);

export interface TwoFactorSetupData { secret: string; url: string; }
export const setup2FA = () => api.post<TwoFactorSetupData>('/auth/2fa/setup');
export const enable2FA = (code: string) => api.post<{ backup_codes: string[] }>('/auth/2fa/enable', { code });
export const disable2FA = (password: string) => api.post('/auth/2fa/disable', { password });
export const regenerateBackupCodes = (password: string) => api.post<{ backup_codes: string[] }>('/auth/2fa/backup-codes', { password });
export const verify2FA = (challengeToken: string, code: string) => api.post('/auth/2fa/verify', { challenge_token: challengeToken, code });

export const requestPasswordReset = (email: string) => api.post('/auth/forgot-password', { email });
export const resetPassword = (token: string, password: string) => api.post('/auth/reset-password', { token, password });

export const sendVerificationEmail = (email?: string) => api.post('/auth/email/send-verification', email ? { email } : {});
export const verifyEmail = (token: string) => api.post('/auth/email/verify', { token });
