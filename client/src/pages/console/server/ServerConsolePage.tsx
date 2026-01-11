import { useState, useEffect } from 'react';
import { useParams, Link, useLocation, Routes, Route } from 'react-router-dom';
import { startServer, stopServer, restartServer, killServer, Server } from '../../../lib/api';
import { formatBytes } from '../../../lib/utils';
import { Icons, StatCard } from '../../../components';
import { PowerButton } from '../../../components/ui/PowerButton';
import { useServerConsole, DEFAULT_STATS, LogLine, ServerStats } from '../../../hooks/useServerConsole';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { getPluginTabs, evaluatePluginGuard } from '../../../lib/pluginLoader';
import { PluginRenderer } from '../../../components/plugins';
import FilesPage from './FilesPage';
import FileEditorPage from './FileEditorPage';
import StartupPage from './StartupPage';
import NetworkPage from './NetworkPage';
import ResourcesPage from './ResourcesPage';
import BackupsPage from './BackupsPage';
import SchedulesPage from './SchedulesPage';
import ServerSettingsPage from './ServerSettingsPage';
import SubusersPage from './SubusersPage';
import AddonsPage from './AddonsPage';
import DatabasesPage from './DatabasesPage';
import ActivityPage from './ActivityPage';

const tabs = [
  { name: 'Console', path: '', icon: 'console' },
  { name: 'Files', path: '/files', icon: 'folder' },
  { name: 'Addons', path: '/addons', icon: 'cube' },
  { name: 'Databases', path: '/databases', icon: 'database' },
  { name: 'Backups', path: '/backups', icon: 'archive' },
  { name: 'Schedules', path: '/schedules', icon: 'clock' },
  { name: 'Startup', path: '/startup', icon: 'sliders' },
  { name: 'Network', path: '/network', icon: 'globe' },
  { name: 'Resources', path: '/resources', icon: 'pieChart' },
  { name: 'Activity', path: '/activity', icon: 'activity' },
  { name: 'Subusers', path: '/subusers', icon: 'users' },
  { name: 'Settings', path: '/settings', icon: 'cogFilled' },
] as const;

interface PluginTabInfo {
  pluginId: string;
  id: string;
  name: string;
  path: string;
  icon: string;
  component: string;
}

export default function ServerConsolePage() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const { server, setServer, ready, logs, stats, setStats, consoleRef, wsRef, addLog, wsError } = useServerConsole(id);
  const { can } = useServerPermissions(id);
  const [ui, setUi] = useState({ command: '', loading: null as string | null, stopping: false });
  const [commandHistory, setCommandHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);

  const basePath = `/console/server/${id}`;
  const currentPath = location.pathname.replace(basePath, '') || '';

  const pluginTabs = getPluginTabs('server')
    .filter(({ pluginId, tab }) => evaluatePluginGuard(pluginId, tab.guard))
    .map(({ pluginId, tab }): PluginTabInfo => ({
      pluginId,
      id: tab.id,
      name: tab.label,
      path: `/plugin/${pluginId}/${tab.id}`,
      icon: tab.icon || 'puzzle',
      component: tab.component,
    }));

  const hasAddonSources = server?.package?.addon_sources && server.package.addon_sources.length > 0;
  const visibleTabs = tabs.filter(tab => {
    if (tab.path === '/subusers') return can('*');
    if (tab.path === '/addons') return hasAddonSources;
    return true;
  });

  const handlePowerAction = async (action: 'start' | 'stop' | 'restart' | 'kill') => {
    if (!server || (ui.loading && action !== 'kill')) return;
    setUi(s => ({ ...s, loading: action }));
    try {
      if (action === 'kill') {
        await killServer(server.id);
        setServer(s => s ? { ...s, status: 'stopped' } : null);
        setUi(s => ({ ...s, stopping: false }));
        setStats(DEFAULT_STATS);
      } else if (action === 'stop') {
        setUi(s => ({ ...s, stopping: true }));
        await stopServer(server.id);
        setServer(s => s ? { ...s, status: 'stopped' } : null);
        setUi(s => ({ ...s, stopping: false }));
        setStats(DEFAULT_STATS);
      } else if (action === 'restart') {
        setUi(s => ({ ...s, stopping: true }));
        setStats(null);
        const res = await restartServer(server.id);
        if (!res.success) addLog(`[ERROR] ${res.error}`, '#ef4444');
        else setServer(s => s ? { ...s, status: 'running' } : null);
        setUi(s => ({ ...s, stopping: false }));
      } else if (action === 'start') {
        setStats(null);
        const res = await startServer(server.id);
        if (!res.success) addLog(`[ERROR] ${res.error}`, '#ef4444');
        else setServer(s => s ? { ...s, status: 'running' } : null);
      }
    } catch { setUi(s => ({ ...s, stopping: false })); }
    setUi(s => ({ ...s, loading: null }));
  };

  const handleCommand = (e: React.FormEvent) => {
    e.preventDefault();
    if (!ui.command.trim() || !wsRef.current) return;
    wsRef.current.send(JSON.stringify({ type: 'command', command: ui.command.trim() }));
    addLog(`> ${ui.command}`);
    setCommandHistory(h => [ui.command.trim(), ...h].slice(0, 50));
    setHistoryIndex(-1);
    setUi(s => ({ ...s, command: '' }));
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (commandHistory.length === 0) return;
      const newIndex = Math.min(historyIndex + 1, commandHistory.length - 1);
      setHistoryIndex(newIndex);
      setUi(s => ({ ...s, command: commandHistory[newIndex] }));
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex <= 0) {
        setHistoryIndex(-1);
        setUi(s => ({ ...s, command: '' }));
      } else {
        const newIndex = historyIndex - 1;
        setHistoryIndex(newIndex);
        setUi(s => ({ ...s, command: commandHistory[newIndex] }));
      }
    }
  };

  if (!ready) return null;
  if (!server) return <div className="text-neutral-400">Server not found</div>;

  return (
    <>
      <nav className="w-[calc(100%+3rem)] sm:w-[calc(100%+4rem)] border-b border-neutral-800 bg-neutral-900 -mt-6 -mx-6 sm:-mx-8 mb-6">
        <div className="flex items-center gap-0.5 px-4 sm:px-6 py-2 overflow-x-auto">
          {visibleTabs.map(tab => {
            const isActive = currentPath === tab.path;
            const IconComponent = Icons[tab.icon];
            return (
              <Link
                key={tab.path}
                to={`${basePath}${tab.path}`}
                className={`group inline-flex items-center gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors cursor-pointer border ${isActive
                    ? 'font-semibold text-neutral-100 bg-neutral-700/90 border-transparent shadow-xs'
                    : 'font-medium text-neutral-400 border-transparent hover:text-neutral-200 hover:border-neutral-800'
                  }`}
              >
                <IconComponent className="h-4 w-4" />
                <span>{tab.name}</span>
              </Link>
            );
          })}
          {pluginTabs.map(tab => {
            const isActive = currentPath === tab.path;
            const IconComponent = Icons[tab.icon as keyof typeof Icons] || Icons.puzzle;
            return (
              <Link
                key={tab.path}
                to={`${basePath}${tab.path}`}
                className={`group inline-flex items-center gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors cursor-pointer border ${isActive
                    ? 'font-semibold text-neutral-100 bg-neutral-700/90 border-transparent shadow-xs'
                    : 'font-medium text-neutral-400 border-transparent hover:text-neutral-200 hover:border-neutral-800'
                  }`}
              >
                <IconComponent className="h-4 w-4" />
                <span>{tab.name}</span>
              </Link>
            );
          })}
        </div>
      </nav>

      <Routes>
        <Route path="files" element={<FilesPage />} />
        <Route path="files/edit" element={<FileEditorPage />} />
        <Route path="addons" element={<AddonsPage />} />
        <Route path="databases" element={<DatabasesPage />} />
        <Route path="backups" element={<BackupsPage />} />
        <Route path="schedules" element={<SchedulesPage />} />
        <Route path="startup" element={<StartupPage />} />
        <Route path="network" element={<NetworkPage />} />
        <Route path="resources" element={<ResourcesPage />} />
        <Route path="activity" element={<ActivityPage />} />
        <Route path="settings" element={<ServerSettingsPage />} />
        <Route path="subusers" element={<SubusersPage />} />
        {pluginTabs.map(tab => (
          <Route
            key={tab.path}
            path={`plugin/${tab.pluginId}/${tab.id}`}
            element={<PluginRenderer pluginId={tab.pluginId} component={tab.component} props={{ serverId: id, server }} />}
          />
        ))}
        <Route path="*" element={
          <ConsoleContent
            server={server}
            logs={logs}
            command={ui.command}
            setCommand={v => setUi(s => ({ ...s, command: v }))}
            handleCommand={handleCommand}
            handleKeyDown={handleKeyDown}
            handlePowerAction={handlePowerAction}
            actionLoading={ui.loading}
            isStopping={ui.stopping}
            stats={stats}
            consoleRef={consoleRef}
            wsError={wsError}
          />
        } />
      </Routes>
    </>
  );
}

function ConsoleContent({ server, logs, command, setCommand, handleCommand, handleKeyDown, handlePowerAction, actionLoading, isStopping, stats, consoleRef, wsError }: {
  server: Server; logs: LogLine[]; command: string; setCommand: (v: string) => void;
  handleCommand: (e: React.FormEvent) => void; handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void; handlePowerAction: (action: 'start' | 'stop' | 'restart' | 'kill') => void;
  actionLoading: string | null; isStopping: boolean; stats: ServerStats | null; consoleRef: React.RefObject<HTMLDivElement>; wsError: string | null;
}) {
  const isRunning = server.status === 'running';
  const isSuspended = server.is_suspended;
  const statusColor = isSuspended ? 'bg-amber-500' : isRunning ? 'bg-emerald-500' : server.status === 'stopped' ? 'bg-neutral-500' : 'bg-yellow-500';
  const displayStats = stats || DEFAULT_STATS;
  const isLoading = stats === null && isRunning;

  useEffect(() => {
    requestAnimationFrame(() => {
      if (consoleRef.current) {
        consoleRef.current.scrollTop = consoleRef.current.scrollHeight;
      }
    });
  }, [consoleRef]);

  return (
    <div className="space-y-6">
      {wsError && (
        <div className="rounded-lg bg-red-500/10 border border-red-500/20 px-4 py-3 flex items-center gap-3">
          <Icons.errorCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
          <div>
            <p className="text-sm font-medium text-red-400">Access Denied</p>
            <p className="text-xs text-red-400/70">{wsError}</p>
          </div>
        </div>
      )}

      {isSuspended && (
        <div className="rounded-lg bg-amber-500/10 border border-amber-500/20 px-4 py-3 flex items-center gap-3">
          <Icons.errorCircle className="w-5 h-5 text-amber-400 flex-shrink-0" />
          <div>
            <p className="text-sm font-medium text-amber-400">Server Suspended</p>
            <p className="text-xs text-amber-400/70">This server has been suspended by an administrator. Please contact support for assistance.</p>
          </div>
        </div>
      )}

      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <span className="relative inline-flex h-3 w-3 items-center justify-center">
              <span className={`absolute h-3 w-3 rounded-full ${statusColor} opacity-35`}></span>
              <span className={`relative h-1.5 w-1.5 rounded-full ${statusColor}`}></span>
            </span>
            <h1 className="text-xl font-semibold text-neutral-900 dark:text-neutral-100">{server.name}</h1>
          </div>
          <div className="mt-1 flex items-center gap-2">
            <Icons.globe className="w-4 h-4 text-neutral-500 dark:text-neutral-400" />
            <span className="text-xs text-neutral-900 dark:text-neutral-100">{server.node?.display_ip || server.node?.fqdn || 'Unknown'}:{(server.ports as any)?.[0]?.port || '?'}</span>
            <span className="text-xs text-neutral-400 dark:text-neutral-500">•</span>
            <span className="text-xs text-neutral-600 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 cursor-pointer transition-colors">{server.node?.name || 'Unknown'}</span>
          </div>
        </div>

        <div className="flex w-full sm:w-auto items-center gap-2 flex-wrap">
          <PowerButton variant="start" onClick={() => handlePowerAction('start')} disabled={isRunning || actionLoading !== null} />
          <PowerButton variant="restart" onClick={() => handlePowerAction('restart')} disabled={!isRunning || actionLoading !== null} />
          <PowerButton variant={isStopping ? 'kill' : 'stop'} onClick={() => handlePowerAction(isStopping ? 'kill' : 'stop')} disabled={(!isRunning && !isStopping) || (actionLoading !== null && actionLoading !== 'stop')} />
        </div>
      </div>

      <div className="space-y-4">
        <div className="rounded-xl border border-neutral-800 overflow-hidden">
          <div ref={consoleRef} className="h-[50vh] bg-neutral-900 overflow-y-auto font-mono text-sm text-neutral-200 p-3">
            {logs.map((log, i) => (
              <div key={i} className="flex items-start gap-1.5 mb-0.5 whitespace-pre-wrap">
                <span className="bg-neutral-800 text-neutral-200 rounded-full px-1.5 text-[11px] font-semibold leading-[18px]">{log.time}</span>
                {log.isAxis && (
                  <span className="bg-white text-neutral-900 rounded-full px-1.5 text-[11px] font-semibold leading-[18px]">Birdactyl</span>
                )}
                <span style={{ color: log.color }}>{log.text}</span>
              </div>
            ))}
          </div>
          <form onSubmit={handleCommand} className="flex items-center gap-2 px-3 py-2.5 bg-neutral-800/50 border-t border-neutral-800">
            <span className="font-mono text-neutral-500 select-none">$</span>
            <input
              value={command}
              onChange={e => setCommand(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="enter command"
              className="flex-1 bg-transparent text-neutral-100 placeholder-neutral-500 outline-none font-mono text-sm"
              autoComplete="off"
              autoCorrect="off"
              spellCheck={false}
            />
          </form>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <StatCard label="Memory" value={isLoading ? '...' : formatBytes(displayStats.memoryUsage)} max={`of ${formatBytes(displayStats.memoryLimit || server.memory * 1024 * 1024)}`} percent={(displayStats.memoryUsage / (displayStats.memoryLimit || server.memory * 1024 * 1024)) * 100} icon="pieChart" />
          <StatCard label="CPU" value={isLoading ? '...' : `${displayStats.cpuPercent.toFixed(2)}%`} max={`of ${server.cpu}.00%`} percent={(displayStats.cpuPercent / server.cpu) * 100} icon="cpu" />
          <StatCard label="Disk" value={isLoading ? '...' : formatBytes(displayStats.diskUsage)} max={`of ${(server.disk / 1024).toFixed(2)} GiB`} percent={(displayStats.diskUsage / (server.disk * 1024 * 1024)) * 100} icon="disk" />
          <StatCard label="Network" value={isLoading ? '...' : formatBytes(displayStats.netRx)} max={`↓ RX · ↑ ${formatBytes(displayStats.netTx)} TX`} percent={50} icon="networkSignal" />
        </div>
      </div>
    </div>
  );
}
