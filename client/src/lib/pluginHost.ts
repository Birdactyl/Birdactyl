import * as React from 'react';
import * as ReactDOM from 'react-dom';
import * as ReactRouterDOM from 'react-router-dom';

import * as UI from '../components/ui';
import { Icons } from '../components/Icons';
import { notify } from '../components/feedback/Notification';
import { api } from './api/client';
import { getUser, isAdmin } from './auth';
import { eventBus, EventMap, EventName } from './eventBus';

export interface PluginAPI {
  pluginId: string;
  get: <T = unknown>(path: string) => Promise<T>;
  post: <T = unknown>(path: string, body?: unknown) => Promise<T>;
  put: <T = unknown>(path: string, body?: unknown) => Promise<T>;
  delete: <T = unknown>(path: string) => Promise<T>;
  notify: (title: string, message: string, type?: 'success' | 'error' | 'info') => void;
  getUser: () => { id: string; username: string; email: string; is_admin: boolean } | null;
  isAdmin: () => boolean;
  navigate: (path: string) => void;
}

function createPluginAPI(pluginId: string, navigate: (path: string) => void): PluginAPI {
  const basePath = `/plugins/${pluginId}`;
  
  return {
    pluginId,
    get: async <T = unknown>(path: string) => {
      const result = await api.get<T>(`${basePath}${path}`);
      if (!result.success) throw new Error(result.error || 'Request failed');
      return result.data as T;
    },
    post: async <T = unknown>(path: string, body?: unknown) => {
      const result = await api.post<T>(`${basePath}${path}`, body);
      if (!result.success) throw new Error(result.error || 'Request failed');
      return result.data as T;
    },
    put: async <T = unknown>(path: string, body?: unknown) => {
      const result = await api.put<T>(`${basePath}${path}`, body);
      if (!result.success) throw new Error(result.error || 'Request failed');
      return result.data as T;
    },
    delete: async <T = unknown>(path: string) => {
      const result = await api.delete<T>(`${basePath}${path}`);
      if (!result.success) throw new Error(result.error || 'Request failed');
      return result.data as T;
    },
    notify,
    getUser,
    isAdmin,
    navigate,
  };
}

const PluginAPIContext = React.createContext<PluginAPI | null>(null);

export function usePluginAPI(): PluginAPI {
  const ctx = React.useContext(PluginAPIContext);
  if (!ctx) throw new Error('usePluginAPI must be used within a PluginProvider');
  return ctx;
}

export const PluginAPIProvider = PluginAPIContext.Provider;
export { createPluginAPI };

export interface BirdactylHost {
  React: typeof React;
  ReactDOM: typeof ReactDOM;
  ReactRouterDOM: typeof ReactRouterDOM;
  UI: typeof UI & { Icons: typeof Icons };
  hooks: {
    usePluginAPI: typeof usePluginAPI;
    useState: typeof React.useState;
    useEffect: typeof React.useEffect;
    useCallback: typeof React.useCallback;
    useMemo: typeof React.useMemo;
    useRef: typeof React.useRef;
    useContext: typeof React.useContext;
  };
  events: {
    on: <K extends EventName>(event: K, callback: (data: EventMap[K]) => void) => () => void;
    off: <K extends EventName>(event: K, callback: (data: EventMap[K]) => void) => void;
    once: <K extends EventName>(event: K, callback: (data: EventMap[K]) => void) => () => void;
    emit: <K extends EventName>(event: K, data: EventMap[K]) => void;
  };
  createPluginAPI: typeof createPluginAPI;
  PluginAPIProvider: typeof PluginAPIProvider;
}

export function initPluginHost(): void {
  const host: BirdactylHost = {
    React,
    ReactDOM,
    ReactRouterDOM,
    UI: {
      ...UI,
      Icons,
    },
    hooks: {
      usePluginAPI,
      useState: React.useState,
      useEffect: React.useEffect,
      useCallback: React.useCallback,
      useMemo: React.useMemo,
      useRef: React.useRef,
      useContext: React.useContext,
    },
    events: {
      on: (event: any, callback: any) => eventBus.on(event, callback),
      off: (event: any, callback: any) => eventBus.off(event, callback),
      once: (event: any, callback: any) => eventBus.once(event, callback),
      emit: (event: any, data: any) => eventBus.emit(event, data),
    },
    createPluginAPI,
    PluginAPIProvider,
  };

  (window as any).React = React;
  (window as any).ReactDOM = ReactDOM;
  (window as any).BIRDACTYL = host;
  
  (window as any).BIRDACTYL_SDK = {
    usePluginAPI,
    useState: React.useState,
    useEffect: React.useEffect,
    useCallback: React.useCallback,
    useMemo: React.useMemo,
    useRef: React.useRef,
    useContext: React.useContext,
    useEvent: function(event: string, callback: (data: any) => void) {
      const callbackRef = React.useRef(callback);
      const eventRef = React.useRef(event);
      callbackRef.current = callback;
      eventRef.current = event;
      React.useEffect(() => {
        const handler = (data: any) => callbackRef.current(data);
        const unsub = eventBus.on(eventRef.current as EventName, handler);
        return unsub;
      }, []);
    },
    events: {
      on: (event: string, callback: (data: any) => void) => eventBus.on(event as EventName, callback),
      off: (event: string, callback: (data: any) => void) => eventBus.off(event as EventName, callback),
      once: (event: string, callback: (data: any) => void) => eventBus.once(event as EventName, callback),
      emit: (event: string, data: any) => eventBus.emit(event as EventName, data),
    },
    ...UI,
    Icons,
  };
}

declare global {
  interface Window {
    BIRDACTYL: BirdactylHost;
  }
}
