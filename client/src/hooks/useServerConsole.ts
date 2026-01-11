import { useEffect, useState, useRef, useCallback } from 'react';
import { getServer, getServerStatus, Server } from '../lib/api';
import { connectServerLogs } from '../lib/api/files';
import { startLoading, finishLoading } from '../lib/pageLoader';
import { eventBus } from '../lib/eventBus';

export interface LogLine { time: string; text: string; color: string; isAxis?: boolean; }
export interface ServerStats { memoryUsage: number; memoryLimit: number; cpuPercent: number; diskUsage: number; netRx: number; netTx: number; }

export const DEFAULT_STATS: ServerStats = { memoryUsage: 0, memoryLimit: 1, cpuPercent: 0, diskUsage: 0, netRx: 0, netTx: 0 };

const MAX_RECONNECT_DELAY = 30000;
const INITIAL_RECONNECT_DELAY = 1000;

function parseLogLine(raw: string): LogLine | null {
  const isBirdactylAxis = raw.includes('\x1b[36m[Birdactyl Axis]') || raw.includes('[Birdactyl Axis]');
  let text = raw
    .replace(/\x1b\[[0-9;?]*[a-zA-Z]/g, '')
    .replace(/␛\[[0-9;?]*[a-zA-Z=]/g, '')
    .replace(/\x1b[=]/g, '')
    .replace(/␛=/g, '')
    .replace(/␛8/g, '')
    .replace(/␛7/g, '')
    .replace(/^\d+T[\d:.]+Z\s*/, '')
    .replace(/>\.+\s*/g, '')
    .replace(/\[K/g, '')
    .replace(/^\s*\d+%\s*#+\s*$/g, '')
    .trim();
  if (!text) return null;

  const timeMatch = text.match(/\[(\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR|DEBUG)\]/);
  const time = timeMatch ? timeMatch[1] : new Date().toLocaleTimeString('en-GB', { hour12: false });

  let color = '#ffffff';
  let isAxis = false;
  if (isBirdactylAxis) {
    isAxis = true;
    text = text.replace(/\[Birdactyl Axis\]\s*/, '');
    color = '#a78bfa';
  } else if (text.includes(' INFO]') || text.includes('/INFO]')) color = '#22ee66';
  else if (text.includes(' WARN]') || text.includes('/WARN]')) color = '#ffd84d';
  else if (text.includes(' ERROR]') || text.includes('/ERROR]')) color = '#ff5555';
  else if (text.includes('Exception') || text.includes('at ')) color = '#ff5555';

  return { time, text, color, isAxis };
}

export function useServerConsole(id: string | undefined) {
  const [server, setServer] = useState<Server | null>(null);
  const [ready, setReady] = useState(false);
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [stats, setStats] = useState<ServerStats | null>(null);
  const [wsError, setWsError] = useState<string | null>(null);
  const [connectionState, setConnectionState] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const consoleRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const prevStatusRef = useRef<string | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectDelayRef = useRef(INITIAL_RECONNECT_DELAY);
  const mountedRef = useRef(true);
  const lastPongRef = useRef<number>(Date.now());

  useEffect(() => {
    startLoading();
    if (!id) return;
    getServer(id).then(res => {
      if (res.success && res.data) {
        setServer(res.data);
        prevStatusRef.current = res.data.status;
      }
      setReady(true);
      finishLoading();
    });
    getServerStatus(id).then(res => {
      if (res.success && res.data?.stats?.disk) {
        setStats(prev => ({ ...(prev || DEFAULT_STATS), diskUsage: res.data!.stats!.disk }));
      }
    });
  }, [id]);

  const connect = useCallback(() => {
    if (!id || !mountedRef.current) return;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setConnectionState('connecting');

    const ws = connectServerLogs(id, (msg) => {
      try {
        const data = JSON.parse(msg);
        if (data.error) {
          if (data.error === 'Permission denied') {
            setWsError("You don't have permission to view the console");
          } else {
            setWsError(data.error);
          }
          return;
        }
        if (data.type === 'log') {
          const parsed = parseLogLine(data.data);
          if (parsed) {
            setLogs(l => [...l.slice(-500), parsed]);
            if (id) eventBus.emit('server:log', { serverId: id, line: data.data });
          }
        } else if (data.type === 'stats' && data.stats) {
          const newStats = {
            memoryUsage: data.stats.memory_usage || 0,
            memoryLimit: data.stats.memory_limit || 1,
            cpuPercent: data.stats.cpu_percent || 0,
            diskUsage: data.stats.disk_usage || 0,
            netRx: data.stats.net_rx || 0,
            netTx: data.stats.net_tx || 0,
          };
          setStats(newStats);
          if (id) eventBus.emit('server:stats', { 
            serverId: id, 
            memory: newStats.memoryUsage, 
            memoryLimit: newStats.memoryLimit, 
            cpu: newStats.cpuPercent, 
            disk: newStats.diskUsage 
          });
        } else if (data.type === 'status') {
          const newStatus = data.status;
          const oldStatus = prevStatusRef.current || '';
          if (prevStatusRef.current === 'installing' && newStatus === 'running') setLogs([]);
          setServer(s => s ? { ...s, status: newStatus } : null);
          prevStatusRef.current = newStatus;
          if (newStatus === 'stopped') setStats(prev => ({ ...DEFAULT_STATS, diskUsage: prev?.diskUsage || 0 }));
          if (id) eventBus.emit('server:status', { serverId: id, status: newStatus, previousStatus: oldStatus });
        }
      } catch {
        const parsed = parseLogLine(msg);
        if (parsed) setLogs(l => [...l.slice(-500), parsed]);
      }
    }, () => {
      if (mountedRef.current) {
        setConnectionState('disconnected');
        scheduleReconnect();
      }
    });

    ws.onopen = () => {
      if (mountedRef.current) {
        setConnectionState('connected');
        setWsError(null);
        reconnectDelayRef.current = INITIAL_RECONNECT_DELAY;
        lastPongRef.current = Date.now();
      }
    };

    ws.onclose = () => {
      if (mountedRef.current) {
        setConnectionState('disconnected');
        scheduleReconnect();
      }
    };

    wsRef.current = ws;
  }, [id]);

  const scheduleReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    const delay = reconnectDelayRef.current;
    reconnectDelayRef.current = Math.min(delay * 1.5, MAX_RECONNECT_DELAY);

    reconnectTimeoutRef.current = window.setTimeout(() => {
      if (mountedRef.current) {
        connect();
      }
    }, delay);
  }, [connect]);

  useEffect(() => {
    mountedRef.current = true;

    return () => {
      mountedRef.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!id || !server) return;
    connect();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [id, server?.id, connect]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && connectionState === 'disconnected') {
        reconnectDelayRef.current = INITIAL_RECONNECT_DELAY;
        connect();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [connectionState, connect]);

  const isNearBottomRef = useRef(true);
  const scrollingRef = useRef(false);
  const scrollTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const el = consoleRef.current;
    if (!el) return;
    
    const handleScroll = () => {
      if (scrollingRef.current) return;
      const threshold = 100;
      isNearBottomRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
    };
    
    el.addEventListener('scroll', handleScroll);
    return () => el.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    const el = consoleRef.current;
    if (el && isNearBottomRef.current) {
      scrollingRef.current = true;
      el.scrollTop = el.scrollHeight;
      if (scrollTimeoutRef.current) {
        clearTimeout(scrollTimeoutRef.current);
      }
      scrollTimeoutRef.current = window.setTimeout(() => {
        scrollingRef.current = false;
      }, 50);
    }
  }, [logs]);

  useEffect(() => {
    return () => {
      if (scrollTimeoutRef.current) {
        clearTimeout(scrollTimeoutRef.current);
      }
    };
  }, []);

  const addLog = (text: string, color = '#888888') => {
    setLogs(l => [...l, { time: new Date().toLocaleTimeString('en-GB', { hour12: false }), text, color }]);
  };

  const reconnect = useCallback(() => {
    reconnectDelayRef.current = INITIAL_RECONNECT_DELAY;
    connect();
  }, [connect]);

  return { server, setServer, ready, logs, setLogs, stats, setStats, consoleRef, wsRef, addLog, wsError, connectionState, reconnect };
}
