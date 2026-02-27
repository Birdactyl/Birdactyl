import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { startLoading, finishLoading } from '../../lib/pageLoader';
import { getServers, Server, startServer, stopServer, getServerStatus } from '../../lib/api';
import { getUser } from '../../lib/auth';
import { notify, Button, ContextMenu, ContextMenuZone, Icons, StatusDot } from '../../components';
import { CreateServerModal, DeleteServerModal } from '../../components/modals';


function formatTimeAgo(date: string) {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  const weeks = Math.floor(days / 7);
  return `${weeks}w ago`;
}

function ServerRow({ server, onAction }: { server: Server; onAction: () => void }) {
  const [actionLoading, setActionLoading] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [stats, setStats] = useState({ memoryGiB: '0.00', cpuPercent: '0.00' });

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>;

    const fetchStats = async () => {
      if (server.status !== 'running') {
        setStats({ memoryGiB: '0.00', cpuPercent: '0.00' });
        return;
      }
      try {
        const res = await getServerStatus(server.id);
        if (res.success && res.data?.stats) {
          const memGiB = (res.data.stats.memory / (1024 * 1024 * 1024)).toFixed(2);
          const cpuPct = res.data.stats.cpu.toFixed(2);
          setStats({ memoryGiB: memGiB, cpuPercent: cpuPct });
        }
      } catch (err) {
        // ignorance is bliss
      }
    };

    fetchStats();
    if (server.status === 'running') {
      interval = setInterval(fetchStats, 5000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [server.id, server.status]);
  const currentUser = getUser();
  const isOwner = server.user_id === currentUser?.id;
  const ownerName = isOwner ? currentUser?.username : server.user?.username;

  const handleStart = async () => {
    setActionLoading(true);
    const res = await startServer(server.id);
    if (res.success) { notify('Server Starting', server.name, 'success'); onAction(); }
    else notify('Error', res.error || 'Failed to start', 'error');
    setActionLoading(false);
  };

  const handleStop = async () => {
    setActionLoading(true);
    const res = await stopServer(server.id);
    if (res.success) { notify('Server Stopped', server.name, 'success'); onAction(); }
    else notify('Error', res.error || 'Failed to stop', 'error');
    setActionLoading(false);
  };

  const memoryGB = (server.memory / 1024).toFixed(2);
  const cpuCores = Math.ceil(server.cpu / 100);
  const isUnavailable = server.status === 'failed';

  const getServerActions = () => [
    ...((server.status === 'stopped' || server.status === 'failed') ? [{ label: 'Start', onClick: handleStart }] : []),
    ...(server.status === 'running' ? [{ label: 'Stop', onClick: handleStop }] : []),
    ...(isOwner ? [{ label: 'Delete', onClick: () => setShowDeleteModal(true), variant: 'danger' as const }] : []),
  ];

  return (
    <>
      <DeleteServerModal
        open={showDeleteModal}
        serverId={server.id}
        serverName={server.name}
        onClose={() => setShowDeleteModal(false)}
        onDeleted={onAction}
      />
      <ContextMenuZone as="tr" items={getServerActions()} className="relative hover:bg-neutral-800/50 transition-colors group">
        <td className="px-3 sm:px-6 py-3 sm:py-4 whitespace-nowrap">
          <Link to={`/console/server/${server.id}`} className="absolute inset-0 z-0" />
          <div className="text-sm font-medium text-neutral-100 flex items-center gap-2 relative z-10 pointer-events-none">
            <span className="truncate">{server.name}</span>
            <StatusDot status={server.status} />
            {server.description && <span className="text-neutral-500 text-xs truncate max-w-[200px]">{"\u2014"} {server.description}</span>}
            {!isOwner && <span className="inline-flex items-center rounded-md bg-sky-500/10 px-1.5 py-0.5 text-[10px] font-medium text-sky-400 ring-1 ring-inset ring-sky-500/20">Shared</span>}
          </div>
          <div className="mt-0.5 hidden sm:inline-flex items-center gap-1 text-xs text-neutral-400">
            <span className="tabular-nums">{stats.memoryGiB}</span>
            <span>of</span>
            <span className="tabular-nums">{memoryGB}</span>
            <span>GiB memory usage</span>
            <span className="w-1 h-1 rounded-full bg-neutral-700 mx-1" />
            <span className="tabular-nums">{stats.cpuPercent}%</span>
            <span>CPU usage</span>
            <span className="text-neutral-500">(of {cpuCores} cores)</span>
          </div>
        </td>
        <td className="pl-3 pr-6 py-4 whitespace-nowrap hidden md:table-cell">
          <div className="flex items-center gap-2">
            <div className="w-[22px] h-[22px] rounded-full ring-1 ring-neutral-800 overflow-hidden flex-shrink-0">
              <div className="w-full h-full bg-neutral-700 flex items-center justify-center text-xs text-neutral-400">
                {ownerName?.charAt(0).toUpperCase() || 'U'}
              </div>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <span className="text-neutral-100 font-medium">{ownerName || 'Unknown'}</span>
              <span className="w-1 h-1 rounded-full bg-neutral-700" />
              <span className="text-neutral-400 tabular-nums">{formatTimeAgo(server.updated_at)}</span>
            </div>
          </div>
        </td>
        <td className="px-3 sm:px-6 py-3 sm:py-4 whitespace-nowrap text-right text-sm font-medium relative z-10">
          <ContextMenu
            align="end"
            trigger={
              <button
                disabled={actionLoading}
                className="p-1 text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800/80 rounded-lg transition-colors disabled:opacity-50"
              >
                {actionLoading ? (
                  <span className="inline-block rounded-full border-2 border-current border-t-transparent animate-spin h-4 w-4" />
                ) : (
                  <Icons.ellipsis className="w-5 h-5" />
                )}
              </button>
            }
            items={getServerActions()}
          />
        </td>
        {isUnavailable && (
          <td className="absolute inset-0 pointer-events-none" aria-hidden="true">
            <div className="absolute inset-0 backdrop-blur-sm bg-neutral-900/60 flex items-center justify-center">
              <div className="text-center">
                <div className="text-sm font-semibold text-neutral-200">
                  Server <span className="font-bold">{server.name}</span> is unavailable
                </div>
                <div className="mt-1 text-xs text-neutral-400">
                  We couldn't connect to the node running your server.
                </div>
              </div>
            </div>
          </td>
        )}
      </ContextMenuZone>
    </>
  );
}

export default function ServersPage() {
  const [ready, setReady] = useState(false);
  const [servers, setServers] = useState<Server[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  const loadServers = async () => {
    setLoading(true);
    const res = await getServers();
    if (res.success && res.data) setServers(res.data);
    setLoading(false);
  };

  useEffect(() => {
    startLoading();
    loadServers().then(() => { setReady(true); finishLoading(); });
  }, []);

  if (!ready) return null;

  return (
    <>
      <CreateServerModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={loadServers} />

      <div className="space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">Servers</h1>
            <p className="text-sm text-neutral-400">View and manage your servers.</p>
          </div>
          <Button onClick={() => setShowCreate(true)} className="w-full sm:w-auto">
            <Icons.plus className="w-4 h-4" />
            Create Server
          </Button>
        </div>

        <div className="bg-neutral-900/40 rounded-lg p-1 overflow-x-auto">
          <table className="min-w-full">
            <colgroup>
              <col style={{ width: '55%' }} />
              <col style={{ width: '35%' }} />
              <col style={{ width: '10%' }} />
            </colgroup>
            <thead>
              <tr>
                <th className="px-3 sm:px-6 py-2 text-left text-xs font-medium text-neutral-400 uppercase tracking-wider">Name</th>
                <th className="pl-3 pr-6 py-2 text-left text-xs font-medium text-neutral-400 uppercase tracking-wider hidden md:table-cell">Updated</th>
                <th className="relative px-6 py-2"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
          </table>
          <div className="mt-1 rounded-lg border border-neutral-800 overflow-hidden">
            <table className="min-w-full">
              <colgroup>
                <col style={{ width: '55%' }} />
                <col style={{ width: '35%' }} />
                <col style={{ width: '10%' }} />
              </colgroup>
              <tbody className={`bg-neutral-900/50 divide-y divide-neutral-700 transition-opacity ${loading ? 'opacity-50' : ''}`}>
                {servers.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-6 py-12 text-center text-neutral-400">
                      {loading ? 'Loading...' : 'No servers yet. Create one to get started.'}
                    </td>
                  </tr>
                ) : (
                  servers.map(server => <ServerRow key={server.id} server={server} onAction={loadServers} />)
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </>
  );
}
