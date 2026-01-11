import { ComponentType } from 'react';
import { api } from './api/client';
import { registry } from './registry';
import { isAdmin, getUser } from './auth';

export type GuardEvaluator = (guard: string, user: { id: string; username: string; email: string; is_admin: boolean } | null) => boolean;

export interface PluginUIPage {
  path: string;
  component: string;
  title?: string;
  icon?: string;
  guard?: string | null;
}

export interface PluginUITab {
  id: string;
  component: string;
  target: 'server' | 'user-settings';
  label: string;
  icon?: string;
  order?: number;
  guard?: string | null;
}

export interface PluginUISidebarItem {
  id: string;
  label: string;
  icon?: string;
  href: string;
  section: 'nav' | 'platform' | 'admin';
  order?: number;
  guard?: string | null;
  children?: { label: string; href: string; guard?: string | null }[];
}

export interface PluginUIManifest {
  id: string;
  name: string;
  version: string;
  hasBundle: boolean;
  pages: PluginUIPage[];
  tabs: PluginUITab[];
  sidebarItems: PluginUISidebarItem[];
}

export interface LoadedPlugin {
  manifest: PluginUIManifest;
  module: Record<string, ComponentType<unknown>> | null;
  evaluateGuard?: GuardEvaluator;
}

const loadedPlugins = new Map<string, LoadedPlugin>();
const loadingPromises = new Map<string, Promise<LoadedPlugin | null>>();

function checkGuard(
  guard: string | null | undefined,
  user: { id: string; username: string; email: string; is_admin: boolean } | null,
  customEvaluator?: GuardEvaluator
): boolean {
  if (!guard) return true;
  if (guard === 'admin') return isAdmin();
  if (customEvaluator) {
    return customEvaluator(guard, user);
  }
  return false;
}

export function evaluatePluginGuard(pluginId: string, guard: string | null | undefined): boolean {
  if (!guard) return true;
  if (guard === 'admin') return isAdmin();
  
  const plugin = loadedPlugins.get(pluginId);
  if (!plugin) return false;
  
  const user = getUser();
  if (plugin.evaluateGuard) {
    return plugin.evaluateGuard(guard, user);
  }
  
  return false;
}

export async function fetchPluginManifests(): Promise<PluginUIManifest[]> {
  const result = await api.get<{ plugins: PluginUIManifest[] }>('/plugins/ui/manifests');
  if (!result.success || !result.data) return [];
  return result.data.plugins;
}

export async function loadPluginBundle(pluginId: string): Promise<Record<string, ComponentType<unknown>> | null> {
  try {
    const response = await fetch(`/api/v1/plugins/${pluginId}/ui/bundle.js`);
    if (!response.ok) {
      console.error(`[plugins] Failed to fetch bundle for ${pluginId}: ${response.status}`);
      return null;
    }
    const code = await response.text();
    
    const wrappedCode = `(function(exports) { ${code.replace('var PluginBundle', 'exports.PluginBundle')} return exports; })({})`;
    
    try {
      const result = eval(wrappedCode);
      if (result.PluginBundle) {
        return result.PluginBundle;
      }
    } catch (evalErr) {
      console.error(`[plugins] Failed to eval bundle for ${pluginId}:`, evalErr);
    }

    const script = document.createElement('script');
    script.textContent = code;
    document.head.appendChild(script);
    document.head.removeChild(script);
    
    if ((window as any).PluginBundle) {
      const bundle = (window as any).PluginBundle;
      delete (window as any).PluginBundle;
      return bundle;
    }

    return null;
  } catch (err) {
    console.error(`[plugins] Failed to load bundle for ${pluginId}:`, err);
    return null;
  }
}

export async function loadPlugin(manifest: PluginUIManifest): Promise<LoadedPlugin | null> {
  if (loadedPlugins.has(manifest.id)) {
    return loadedPlugins.get(manifest.id)!;
  }

  if (loadingPromises.has(manifest.id)) {
    return loadingPromises.get(manifest.id)!;
  }

  const loadPromise = (async () => {
    let module: Record<string, ComponentType<unknown>> | null = null;
    let evaluateGuard: GuardEvaluator | undefined;
    
    if (manifest.hasBundle) {
      module = await loadPluginBundle(manifest.id);
      if (module && typeof (module as any).evaluateGuard === 'function') {
        evaluateGuard = (module as any).evaluateGuard;
      }
    }

    const user = getUser();
    for (const item of manifest.sidebarItems) {
      const canAccess = checkGuard(item.guard, user, evaluateGuard);
      if (!canAccess) continue;
      registry.registerSidebarItem({
        id: `plugin-${manifest.id}-${item.id}`,
        label: item.label,
        icon: item.icon || 'puzzle',
        href: item.href,
        section: item.section,
        order: item.order || 100,
        guard: item.guard as 'admin' | null,
        children: item.children?.filter(c => checkGuard(c.guard, user, evaluateGuard)),
      });
    }

    const loaded: LoadedPlugin = { manifest, module, evaluateGuard };
    loadedPlugins.set(manifest.id, loaded);
    loadingPromises.delete(manifest.id);
    return loaded;
  })();

  loadingPromises.set(manifest.id, loadPromise);
  return loadPromise;
}

export async function loadAllPlugins(): Promise<LoadedPlugin[]> {
  const manifests = await fetchPluginManifests();
  const results = await Promise.all(manifests.map(loadPlugin));
  return results.filter((p): p is LoadedPlugin => p !== null);
}

export function getLoadedPlugin(pluginId: string): LoadedPlugin | undefined {
  return loadedPlugins.get(pluginId);
}

export function getAllLoadedPlugins(): LoadedPlugin[] {
  return Array.from(loadedPlugins.values());
}

export function getPluginComponent(pluginId: string, componentName: string): ComponentType<unknown> | null {
  const plugin = loadedPlugins.get(pluginId);
  if (!plugin?.module) return null;
  return plugin.module[componentName] || null;
}

export function getPluginPages(): Array<{ pluginId: string; page: PluginUIPage }> {
  const pages: Array<{ pluginId: string; page: PluginUIPage }> = [];
  for (const plugin of loadedPlugins.values()) {
    for (const page of plugin.manifest.pages) {
      pages.push({ pluginId: plugin.manifest.id, page });
    }
  }
  return pages;
}

export function getPluginTabs(target: PluginUITab['target']): Array<{ pluginId: string; tab: PluginUITab }> {
  const tabs: Array<{ pluginId: string; tab: PluginUITab }> = [];
  for (const plugin of loadedPlugins.values()) {
    for (const tab of plugin.manifest.tabs) {
      if (tab.target === target) {
        tabs.push({ pluginId: plugin.manifest.id, tab });
      }
    }
  }
  return tabs.sort((a, b) => (a.tab.order || 0) - (b.tab.order || 0));
}

export function getPluginSidebarItems(): PluginUISidebarItem[] {
  const items: PluginUISidebarItem[] = [];
  for (const plugin of loadedPlugins.values()) {
    items.push(...plugin.manifest.sidebarItems);
  }
  return items.sort((a, b) => (a.order || 0) - (b.order || 0));
}
