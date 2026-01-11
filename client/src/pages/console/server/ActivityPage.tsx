import { useState, useRef, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, getServerLogs, type Server, type ActivityLog } from '../../../lib/api';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { notify, Input, Pagination, Icons, DatePicker, Table, PermissionDenied } from '../../../components';

const actionLabels: Record<string, string> = {
  'server.create': 'Create Server',
  'server.delete': 'Delete Server',
  'server.start': 'Start Server',
  'server.stop': 'Stop Server',
  'server.kill': 'Kill Server',
  'server.restart': 'Restart Server',
  'server.reinstall': 'Reinstall Server',
  'server.command': 'Send Command',
  'server.name.update': 'Rename Server',
  'server.resources.update': 'Update Resources',
  'server.variables.update': 'Update Variables',
  'server.allocation.add': 'Add Allocation',
  'server.allocation.set_primary': 'Set Primary Allocation',
  'server.allocation.delete': 'Delete Allocation',
  'server.file.create_folder': 'Create Folder',
  'server.file.write': 'Write File',
  'server.file.upload': 'Upload File',
  'server.file.delete': 'Delete File',
  'server.file.move': 'Move File',
  'server.file.copy': 'Copy File',
  'server.file.compress': 'Compress File',
  'server.file.decompress': 'Decompress File',
  'server.file.bulk_delete': 'Bulk Delete Files',
  'server.file.bulk_copy': 'Bulk Copy Files',
  'server.file.bulk_compress': 'Bulk Compress Files',
  'server.backup.create': 'Create Backup',
  'server.backup.delete': 'Delete Backup',
  'server.subuser.add': 'Add Subuser',
  'server.subuser.update': 'Update Subuser',
  'server.subuser.remove': 'Remove Subuser',
  'server.database.create': 'Create Database',
  'server.database.delete': 'Delete Database',
  'server.database.rotate_password': 'Rotate DB Password',
  'server.schedule.create': 'Create Schedule',
  'server.schedule.update': 'Update Schedule',
  'server.schedule.delete': 'Delete Schedule',
  'server.schedule.run': 'Run Schedule',
  'server.addon.install': 'Install Addon',
  'server.addon.delete': 'Delete Addon',
  'server.modpack.install': 'Install Modpack',
  'admin.server.create': 'Create Server (Admin)',
  'admin.server.view': 'View Server (Admin)',
  'admin.server.suspend': 'Suspend Server',
  'admin.server.unsuspend': 'Unsuspend Server',
  'admin.server.delete': 'Delete Server (Admin)',
  'admin.server.resources': 'Update Resources (Admin)',
  'admin.server.transfer': 'Transfer Server',
};

const getActionIcon = (action: string) => {
  if (action.includes('start')) return <Icons.play className="w-4 h-4 text-emerald-400" />;
  if (action.includes('stop') || action.includes('kill')) return <Icons.stop className="w-4 h-4 text-red-400" />;
  if (action.includes('restart')) return <Icons.refresh className="w-4 h-4 text-amber-400" />;
  if (action.includes('reinstall')) return <Icons.refresh className="w-4 h-4 text-orange-400" />;
  if (action.includes('file') || action.includes('folder')) return <Icons.folder className="w-4 h-4 text-amber-500" />;
  if (action.includes('backup')) return <Icons.archive className="w-4 h-4 text-blue-400" />;
  if (action.includes('subuser')) return <Icons.users className="w-4 h-4 text-violet-400" />;
  if (action.includes('database')) return <Icons.database className="w-4 h-4 text-violet-400" />;
  if (action.includes('schedule')) return <Icons.clock className="w-4 h-4 text-sky-400" />;
  if (action.includes('allocation') || action.includes('network')) return <Icons.globe className="w-4 h-4 text-sky-400" />;
  if (action.includes('addon') || action.includes('modpack')) return <Icons.cube className="w-4 h-4 text-amber-400" />;
  if (action.includes('command')) return <Icons.console className="w-4 h-4 text-neutral-300" />;
  if (action.includes('variables') || action.includes('startup')) return <Icons.sliders className="w-4 h-4 text-purple-400" />;
  if (action.includes('name') || action.includes('rename')) return <Icons.edit className="w-4 h-4 text-blue-400" />;
  if (action.includes('resources')) return <Icons.pieChart className="w-4 h-4 text-emerald-400" />;
  if (action.includes('create')) return <Icons.plus className="w-4 h-4 text-emerald-400" />;
  if (action.includes('delete')) return <Icons.trash className="w-4 h-4 text-red-400" />;
  if (action.includes('suspend')) return <Icons.noAccess className="w-4 h-4 text-amber-400" />;
  if (action.includes('unsuspend')) return <Icons.check className="w-4 h-4 text-emerald-400" />;
  if (action.includes('transfer')) return <Icons.move className="w-4 h-4 text-blue-400" />;
  if (action.includes('view')) return <Icons.eye className="w-4 h-4 text-neutral-400" />;
  return <Icons.activity className="w-4 h-4 text-neutral-400" />;
};

export default function ActivityPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const requestId = useRef(0);
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => res.success && res.data && setServer(res.data));
  }, [id]);

  const load = async (p: number, pp: number, s: string, from: string, to: string, initial = false) => {
    if (!id) return;
    const currentRequest = ++requestId.current;
    setLoading(true);
    const res = await getServerLogs(id, p, pp, s, from, to);
    if (currentRequest !== requestId.current) return;
    if (res.success && res.data) {
      setLogs(res.data.logs || []);
      setPage(res.data.page);
      setTotalPages(res.data.total_pages);
      setTotal(res.data.total);
    } else {
      notify('Error', res.error || 'Failed to load activity', 'error');
    }
    setLoading(false);
    if (initial) setReady(true);
  };

  useEffect(() => {
    if (id && !permsLoading) load(1, perPage, '', '', '', true);
  }, [id, permsLoading]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
    load(1, perPage, searchInput, fromDate, toDate);
  };

  const clearFilters = () => {
    setSearchInput('');
    setSearch('');
    setFromDate('');
    setToDate('');
    load(1, perPage, '', '', '');
  };

  if (permsLoading || !ready) return null;
  if (!can('activity.view')) return <PermissionDenied message="You don't have permission to view activity logs" />;

  const hasFilters = search || fromDate || toDate;

  const columns = [
    {
      key: 'expand', header: '', className: 'w-8',
      render: (log: ActivityLog) => (
        <button onClick={() => { const next = new Set(expanded); next.has(log.id) ? next.delete(log.id) : next.add(log.id); setExpanded(next); }} className="text-neutral-400 hover:text-neutral-200 transition">
          <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(log.id) ? 'rotate-90' : ''}`} />
        </button>
      )
    },
    {
      key: 'action', header: 'Action',
      render: (log: ActivityLog) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-neutral-800 flex items-center justify-center">
            {getActionIcon(log.action)}
          </div>
          <div>
            <span className="text-sm font-medium text-neutral-100">{actionLabels[log.action] || log.action}</span>
            {log.description && <div className="text-xs text-neutral-500 line-clamp-1 max-w-xs">{log.description}</div>}
          </div>
        </div>
      )
    },
    {
      key: 'user', header: 'User',
      render: (log: ActivityLog) => (
        <div className="flex items-center gap-2">
          <span className="inline-flex items-center justify-center h-6 w-6 rounded-md bg-neutral-700 text-xs font-medium text-neutral-200">{log.username?.[0]?.toUpperCase() || '?'}</span>
          <span className="text-sm text-neutral-300">{log.username}</span>
        </div>
      )
    },
    { key: 'ip', header: 'IP', render: (log: ActivityLog) => <span className="text-sm text-neutral-400 font-mono">{log.ip}</span> },
    { key: 'time', header: 'Time', render: (log: ActivityLog) => <span className="text-sm text-neutral-400">{new Date(log.created_at).toLocaleString()}</span> },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Activity</span>
      </div>

      <div>
        <h1 className="text-xl font-semibold text-neutral-100">Activity Log</h1>
        <p className="text-sm text-neutral-400">View all actions performed on this server.</p>
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <form onSubmit={handleSearch} className="flex-1 sm:max-w-sm">
            <Input placeholder="Search by user or action..." value={searchInput} onChange={e => setSearchInput(e.target.value)} />
          </form>
          <Pagination page={page} totalPages={totalPages} total={total} perPage={perPage} onPageChange={p => load(p, perPage, search, fromDate, toDate)} onPerPageChange={pp => { setPerPage(pp); load(1, pp, search, fromDate, toDate); }} loading={loading} />
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <DatePicker label="From:" value={fromDate} onChange={v => { setFromDate(v); load(1, perPage, search, v, toDate); }} />
          <DatePicker label="To:" value={toDate} onChange={v => { setToDate(v); load(1, perPage, search, fromDate, v); }} />
          {hasFilters && <button onClick={clearFilters} className="text-xs text-neutral-400 hover:text-neutral-200 transition-colors">Clear filters</button>}
        </div>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{total} log entries</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table
            columns={columns}
            data={logs}
            keyField="id"
            loading={loading}
            emptyText="No activity recorded yet"
            expandable={{
              isExpanded: log => expanded.has(log.id),
              render: log => {
                const meta = log.metadata ? JSON.parse(log.metadata) : null;
                return (
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                    <div><div className="text-neutral-500 mb-1">User Agent</div><div className="text-neutral-300 truncate max-w-xs" title={log.user_agent}>{log.user_agent || '-'}</div></div>
                    <div><div className="text-neutral-500 mb-1">User ID</div><div className="text-neutral-300 font-mono">{log.user_id.slice(0, 8)}</div></div>
                    <div><div className="text-neutral-500 mb-1">Log ID</div><div className="text-neutral-300 font-mono">{log.id.slice(0, 8)}</div></div>
                    <div><div className="text-neutral-500 mb-1">Raw Action</div><div className="text-neutral-300 font-mono">{log.action}</div></div>
                    {meta && Object.keys(meta).length > 0 && (
                      <div className="col-span-full"><div className="text-neutral-500 mb-1">Details</div><div className="text-neutral-300 font-mono bg-neutral-800/50 rounded p-2 overflow-x-auto">{Object.entries(meta).map(([k, v]) => <div key={k}><span className="text-neutral-500">{k}:</span> {typeof v === 'object' ? JSON.stringify(v) : String(v)}</div>)}</div></div>
                    )}
                  </div>
                );
              }
            }}
          />
        </div>
      </div>
    </div>
  );
}
