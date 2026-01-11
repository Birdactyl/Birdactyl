type EventCallback<T = unknown> = (data: T) => void;

interface EventMap {
  'server:status': { serverId: string; status: string; previousStatus: string };
  'server:stats': { serverId: string; memory: number; memoryLimit: number; cpu: number; disk: number };
  'server:log': { serverId: string; line: string };
  'server:start': { serverId: string };
  'server:stop': { serverId: string };
  'server:restart': { serverId: string };
  'server:kill': { serverId: string };
  'file:created': { serverId: string; path: string };
  'file:deleted': { serverId: string; path: string };
  'file:moved': { serverId: string; from: string; to: string };
  'file:uploaded': { serverId: string; path: string };
  'file:saved': { serverId: string; path: string };
  'navigation': { path: string; previousPath: string };
  'user:login': { userId: string; username: string };
  'user:logout': {};
  [key: `plugin:${string}`]: unknown;
}

type EventName = keyof EventMap | `plugin:${string}`;

class EventBus {
  private listeners: Map<string, Set<EventCallback>> = new Map();

  on<K extends EventName>(event: K, callback: EventCallback<K extends keyof EventMap ? EventMap[K] : unknown>): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback as EventCallback);
    
    return () => this.off(event, callback);
  }

  off<K extends EventName>(event: K, callback: EventCallback<K extends keyof EventMap ? EventMap[K] : unknown>): void {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.delete(callback as EventCallback);
    }
  }

  emit<K extends EventName>(event: K, data: K extends keyof EventMap ? EventMap[K] : unknown): void {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(cb => {
        try {
          cb(data);
        } catch (err) {
          console.error(`[eventBus] Error in listener for ${event}:`, err);
        }
      });
    }
  }

  once<K extends EventName>(event: K, callback: EventCallback<K extends keyof EventMap ? EventMap[K] : unknown>): () => void {
    const wrapper = (data: K extends keyof EventMap ? EventMap[K] : unknown) => {
      this.off(event, wrapper);
      callback(data);
    };
    return this.on(event, wrapper);
  }

  clear(event?: EventName): void {
    if (event) {
      this.listeners.delete(event);
    } else {
      this.listeners.clear();
    }
  }
}

export const eventBus = new EventBus();
export type { EventMap, EventName };
