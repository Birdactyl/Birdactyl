import { eventBus } from './eventBus';

let accessToken: string | null = null;
let refreshPromise: Promise<boolean> | null = null;
let currentUser: { id: string; username: string; email: string; is_admin: boolean; force_password_reset: boolean } | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null): void {
  accessToken = token;
}

export function setRefreshToken(token: string): void {
  document.cookie = `refresh_token=${token}; path=/; max-age=${30 * 24 * 60 * 60}; SameSite=Strict`;
}

export function getRefreshToken(): string | null {
  const match = document.cookie.match(/refresh_token=([^;]+)/);
  return match ? match[1] : null;
}

export function clearTokens(): void {
  const wasLoggedIn = accessToken !== null || currentUser !== null;
  accessToken = null;
  currentUser = null;
  document.cookie = 'refresh_token=; path=/; max-age=0';
  if (wasLoggedIn) eventBus.emit('user:logout', {});
}

export function isAuthenticated(): boolean {
  return accessToken !== null || getRefreshToken() !== null;
}

export function getUser() {
  return currentUser;
}

export function setUser(user: { id: string; username: string; email: string; is_admin: boolean; force_password_reset: boolean } | null) {
  const wasNull = currentUser === null;
  currentUser = user;
  if (wasNull && user) {
    eventBus.emit('user:login', { userId: user.id, username: user.username });
  }
}

export function isAdmin(): boolean {
  return currentUser?.is_admin ?? false;
}

export function requiresPasswordReset(): boolean {
  return currentUser?.force_password_reset ?? false;
}

export function clearPasswordResetFlag(): void {
  if (currentUser) {
    currentUser.force_password_reset = false;
  }
}

export async function initAuth(): Promise<boolean> {
  if (accessToken) return true;
  
  if (refreshPromise) return refreshPromise;

  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  refreshPromise = doRefresh(refreshToken);
  const result = await refreshPromise;
  refreshPromise = null;
  return result;
}

async function doRefresh(refreshToken: string): Promise<boolean> {
  try {
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    const data = await response.json();
    
    if (data.success && data.data) {
      setAccessToken(data.data.access_token);
      setRefreshToken(data.data.refresh_token);
      
      const meResponse = await fetch('/api/v1/auth/me', {
        headers: { 'Authorization': `Bearer ${data.data.access_token}` },
      });
      const meData = await meResponse.json();
      if (meData.success && meData.data) {
        setUser(meData.data);
      }
      
      return true;
    }
  } catch {
  }

  clearTokens();
  return false;
}
