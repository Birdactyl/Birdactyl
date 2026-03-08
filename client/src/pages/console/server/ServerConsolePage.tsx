import { useState, useEffect, useMemo } from 'react';
import { useParams, Routes, Route } from 'react-router-dom';
import { startServer, stopServer, restartServer, killServer, Server } from '../../../lib/api';
import { formatBytes } from '../../../lib/utils';
import { Icons, StatusDot } from '../../../components';
import { SubNavigation } from '../../../components/layout/SubNavigation';
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
import SFTPPage from './SFTPPage';
import ServerMountsPage from './ServerMountsPage';
import { getServerMounts } from '../../../lib/api/servers';

const tabs = [
  { name: 'Console', path: '', icon: 'console' },
  { name: 'Files', path: '/files', icon: 'folder' },
  { name: 'Addons', path: '/addons', icon: 'cube' },
  { name: 'Databases', path: '/databases', icon: 'database' },
  { name: 'Backups', path: '/backups', icon: 'archive' },
  { name: 'Schedules', path: '/schedules', icon: 'clock' },
  { name: 'Startup', path: '/startup', icon: 'sliders' },
  { name: 'Network', path: '/network', icon: 'globe' },
  { name: 'SFTP', path: '/sftp', icon: 'key' },
  { name: 'Mounts', path: '/mounts', icon: 'folder' },
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
  const { server, setServer, ready, logs, stats, setStats, consoleRef, wsRef, addLog, wsError } = useServerConsole(id);
  const { can } = useServerPermissions(id);
  const [ui, setUi] = useState({ command: '', loading: null as string | null, stopping: false });
  const [commandHistory, setCommandHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [hasMounts, setHasMounts] = useState(false);

  useEffect(() => {
    if (!id || !can('mount.read')) return;
    getServerMounts(id).then(res => {
      if (res.success && res.data && res.data.length > 0) {
        setHasMounts(true);
      }
    });
  }, [id, can]);

  const basePath = `/console/server/${id}`;

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
  const visibleTabs = useMemo(() => {
    const filtered = tabs.filter(tab => {
      if (tab.path === '/subusers') return can('*');
      if (tab.path === '/addons') return hasAddonSources;
      if (tab.path === '/mounts') return hasMounts && can('mount.read');
      return true;
    });
    const allTabs = [
      ...filtered.map(t => ({ name: t.name, path: t.path, icon: t.icon })),
      ...pluginTabs.map(t => ({ name: t.name, path: t.path, icon: t.icon })),
    ];
    return allTabs;
  }, [can, hasAddonSources, pluginTabs]);

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
      <SubNavigation basePath={basePath} tabs={visibleTabs} />

      <Routes>
        <Route path="files" element={<FilesPage />} />
        <Route path="files/edit" element={<FileEditorPage />} />
        <Route path="addons" element={<AddonsPage />} />
        <Route path="databases" element={<DatabasesPage />} />
        <Route path="backups" element={<BackupsPage />} />
        <Route path="schedules" element={<SchedulesPage />} />
        <Route path="startup" element={<StartupPage />} />
        <Route path="network" element={<NetworkPage />} />
        <Route path="sftp" element={<SFTPPage />} />
        <Route path="mounts" element={<ServerMountsPage />} />
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

  const displayStats = stats || DEFAULT_STATS;
  const isLoading = stats === null && isRunning;
  const primaryPort = (server.ports as any)?.[0]?.port;

  useEffect(() => {
    requestAnimationFrame(() => {
      if (consoleRef.current) {
        consoleRef.current.scrollTop = consoleRef.current.scrollHeight;
      }
    });
  }, [consoleRef]);

  return (
    <div className="space-y-4">
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

      <div className="rounded-xl border border-neutral-800 overflow-hidden">
        <div className="flex flex-wrap items-center gap-4 sm:gap-6 px-4 sm:px-5 py-3 sm:py-4">
          <div className="flex items-center gap-3 min-w-0 shrink-0">
            <StatusDot status={isSuspended ? 'suspended' : server.status} />
            <div className="min-w-0">
              <div className="text-sm font-semibold text-neutral-100 truncate">{server.name}</div>
              <div className="text-xs text-neutral-500">
                {server.node?.display_ip || server.node?.fqdn || 'Unknown'}{primaryPort ? `:${primaryPort}` : ''}
                <span className="mx-1.5 text-neutral-700">|</span>
                {server.node?.name || 'Unknown'}
              </div>
            </div>
          </div>

          <div className="hidden lg:flex items-center divide-x divide-neutral-800 ml-auto mr-4">
            <div className="px-4 first:pl-0">
              <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Memory</div>
              <div className="text-sm font-medium text-neutral-300 tabular-nums">
                {isLoading ? '...' : formatBytes(displayStats.memoryUsage)}
                <span className="text-neutral-600 ml-1">/ {formatBytes(displayStats.memoryLimit || server.memory * 1024 * 1024)}</span>
              </div>
            </div>
            <div className="px-4">
              <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">CPU</div>
              <div className="text-sm font-medium text-neutral-300 tabular-nums">
                {isLoading ? '...' : `${displayStats.cpuPercent.toFixed(1)}%`}
                <span className="text-neutral-600 ml-1">/ {server.cpu}%</span>
              </div>
            </div>
            <div className="px-4">
              <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Disk</div>
              <div className="text-sm font-medium text-neutral-300 tabular-nums">
                {isLoading ? '...' : formatBytes(displayStats.diskUsage)}
                <span className="text-neutral-600 ml-1">/ {(server.disk / 1024).toFixed(1)} GiB</span>
              </div>
            </div>
            <div className="px-4 last:pr-0">
              <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Network</div>
              <div className="text-sm font-medium text-neutral-300 tabular-nums">
                {isLoading ? '...' : `${formatBytes(displayStats.netRx)} / ${formatBytes(displayStats.netTx)}`}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0 ml-auto lg:ml-0">
            <PowerButton variant="start" onClick={() => handlePowerAction('start')} disabled={isRunning || actionLoading !== null} />
            <PowerButton variant="restart" onClick={() => handlePowerAction('restart')} disabled={!isRunning || actionLoading !== null} />
            <PowerButton variant={isStopping ? 'kill' : 'stop'} onClick={() => handlePowerAction(isStopping ? 'kill' : 'stop')} disabled={(!isRunning && !isStopping) || (actionLoading !== null && actionLoading !== 'stop')} />
          </div>
        </div>

        <div className="lg:hidden flex items-center divide-x divide-neutral-800 border-t border-neutral-800 px-4 sm:px-5 py-3 overflow-x-auto">
          <div className="pr-4">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Memory</div>
            <div className="text-xs font-medium text-neutral-300 tabular-nums">{isLoading ? '...' : formatBytes(displayStats.memoryUsage)}</div>
          </div>
          <div className="px-4">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">CPU</div>
            <div className="text-xs font-medium text-neutral-300 tabular-nums">{isLoading ? '...' : `${displayStats.cpuPercent.toFixed(1)}%`}</div>
          </div>
          <div className="px-4">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Disk</div>
            <div className="text-xs font-medium text-neutral-300 tabular-nums">{isLoading ? '...' : formatBytes(displayStats.diskUsage)}</div>
          </div>
          <div className="pl-4">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Net</div>
            <div className="text-xs font-medium text-neutral-300 tabular-nums">{isLoading ? '...' : formatBytes(displayStats.netRx)}</div>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-neutral-800 overflow-hidden">
        <div ref={consoleRef} className="h-[40vh] sm:h-[55vh] bg-[#0a0a0a] overflow-y-auto font-mono text-[12px] sm:text-[13px] leading-[18px] sm:leading-[20px] text-neutral-300 p-3 sm:p-4 scroll-smooth rounded-t-xl">
          {logs.map((log, i) => (
            <div key={i} className={`flex items-start gap-2 -mx-1 px-1 rounded ${log.isAxis ? 'hover:bg-violet-500/[0.03]' : 'hover:bg-white/[0.02]'}`}>
              <span className="text-neutral-600 text-[11px] leading-[20px] select-none shrink-0 tabular-nums">{log.time}</span>
              {log.isAxis && (
                <span className="inline-flex items-center rounded-md bg-violet-500/10 px-1.5 py-0.5 text-[10px] font-medium text-violet-400 ring-1 ring-inset ring-violet-500/20 shrink-0 select-none leading-none">AXIS</span>
              )}
              <span style={{ color: log.color }} className={`break-all ${log.isAxis ? 'text-neutral-400' : ''}`}>{log.text}</span>
            </div>
          ))}
        </div>

        <form onSubmit={handleCommand} className="flex items-center gap-2 sm:gap-3 px-3 sm:px-4 py-2.5 sm:py-3 bg-[#0a0a0a] border-t border-neutral-800/80">
          <span className="text-emerald-500 font-mono text-sm font-bold select-none">$</span>
          <input
            value={command}
            onChange={e => setCommand(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a command..."
            className="flex-1 bg-transparent text-neutral-100 placeholder-neutral-600 outline-none font-mono text-[13px] caret-emerald-500"
            autoComplete="off"
            autoCorrect="off"
            spellCheck={false}
          />
          <button
            type="submit"
            className="text-neutral-600 hover:text-neutral-300 transition-colors"
            tabIndex={-1}
          >
            <Icons.arrowRight className="w-4 h-4" />
          </button>
        </form>
      </div>
    </div>
  );
}
