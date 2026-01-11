import { api, API_BASE, ParsedResponse } from './client';
import { getAccessToken } from '../auth';
import { eventBus } from '../eventBus';

export interface FileEntry { name: string; size: number; is_dir: boolean; mod_time: number; mode: string; }
export interface SearchResult { name: string; path: string; size: number; is_dir: boolean; mod_time: number; }

export const listFiles = (serverId: string, path = '/') => api.get<FileEntry[]>(`/servers/${serverId}/files?path=${encodeURIComponent(path)}`);
export const readFile = (serverId: string, path: string) => api.get<string>(`/servers/${serverId}/files/read?path=${encodeURIComponent(path)}`);
export const searchFiles = (serverId: string, query: string) => api.get<SearchResult[]>(`/servers/${serverId}/files/search?q=${encodeURIComponent(query)}`);

export const deleteFile = async (serverId: string, path: string) => {
  const result = await api.delete(`/servers/${serverId}/files?path=${encodeURIComponent(path)}`);
  if (result.success) eventBus.emit('file:deleted', { serverId, path });
  return result;
};

export const bulkDeleteFiles = async (serverId: string, paths: string[]) => {
  const result = await api.post<{ deleted: number }>(`/servers/${serverId}/files/bulk-delete`, { paths });
  if (result.success) paths.forEach(path => eventBus.emit('file:deleted', { serverId, path }));
  return result;
};

export const bulkCopyFiles = (serverId: string, paths: string[], dest: string) => api.post<{ copied: number }>(`/servers/${serverId}/files/bulk-copy`, { paths, dest });
export const bulkCompressFiles = (serverId: string, paths: string[], dest: string, format: string) => api.post(`/servers/${serverId}/files/bulk-compress`, { paths, dest, format });

export const moveFile = async (serverId: string, from: string, to: string) => {
  const result = await api.post(`/servers/${serverId}/files/move`, { from, to });
  if (result.success) eventBus.emit('file:moved', { serverId, from, to });
  return result;
};

export const copyFile = (serverId: string, from: string, to: string) => api.post(`/servers/${serverId}/files/copy`, { from, to });
export const compressFile = (serverId: string, path: string, dest: string, format: string) => api.post(`/servers/${serverId}/files/compress`, { path, dest, format });
export const decompressFile = (serverId: string, path: string, dest: string) => api.post(`/servers/${serverId}/files/decompress`, { path, dest });

export const createFolder = async (serverId: string, path: string) => {
  const result = await api.post(`/servers/${serverId}/files/folder`, { path });
  if (result.success) eventBus.emit('file:created', { serverId, path });
  return result;
};

export const writeFile = async (serverId: string, path: string, content: string) => {
  const result = await api.post(`/servers/${serverId}/files/write`, { path, content });
  if (result.success) eventBus.emit('file:saved', { serverId, path });
  return result;
};

export function getDownloadUrl(serverId: string, path: string): string {
  return `${API_BASE}/servers/${serverId}/files/download?path=${encodeURIComponent(path)}&token=${getAccessToken()}`;
}

export async function uploadFile(serverId: string, path: string, file: File, onProgress?: (loaded: number, total: number) => void, signal?: AbortSignal): Promise<ParsedResponse<void>> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('path', path);

  return new Promise((resolve) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', `${API_BASE}/servers/${serverId}/files/upload`);
    xhr.setRequestHeader('Authorization', `Bearer ${getAccessToken()}`);
    if (signal) signal.addEventListener('abort', () => xhr.abort());
    xhr.upload.onprogress = (e) => { if (e.lengthComputable && onProgress) onProgress(e.loaded, e.total); };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        eventBus.emit('file:uploaded', { serverId, path: path + '/' + file.name });
        resolve({ success: true });
      } else {
        resolve({ success: false, error: 'Upload failed' });
      }
    };
    xhr.onerror = () => resolve({ success: false, error: 'Upload failed' });
    xhr.onabort = () => resolve({ success: false, error: 'Cancelled' });
    xhr.send(formData);
  });
}

export function connectServerLogs(serverId: string, onMessage: (msg: string) => void, onError?: (err: Event) => void): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${protocol}//${window.location.host}${API_BASE}/servers/${serverId}/logs?token=${getAccessToken()}`);
  ws.onmessage = (event) => onMessage(event.data);
  ws.onerror = (err) => onError?.(err);
  return ws;
}
