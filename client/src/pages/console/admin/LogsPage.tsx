import { useState, useRef, useEffect } from 'react';
import { adminGetLogs, type ActivityLog } from '../../../lib/api';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { notify, Button, Input, Pagination, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DatePicker, Table } from '../../../components';

type Filter = 'all' | 'admin' | 'user';

const actionLabels: Record<string, string> = {
  'auth.register': 'Register',
  'auth.login': 'Login',
  'auth.logout': 'Logout',
  'auth.logout_all': 'Logout All',
  'profile.update': 'Update Profile',
  'profile.password_change': 'Change Password',
  'profile.session_revoke': 'Revoke Session',
  'profile.sessions_revoke_all': 'Revoke All Sessions',
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
  'admin.user.create': 'Create User',
  'admin.user.update': 'Update User',
  'admin.user.delete': 'Delete User',
  'admin.user.ban': 'Ban User',
  'admin.user.unban': 'Unban User',
  'admin.user.set_admin': 'Grant Admin',
  'admin.user.revoke_admin': 'Revoke Admin',
  'admin.user.force_reset': 'Force Password Reset',
  'admin.server.create': 'Create Server (Admin)',
  'admin.server.view': 'View Server (Admin)',
  'admin.server.suspend': 'Suspend Server',
  'admin.server.unsuspend': 'Unsuspend Server',
  'admin.server.delete': 'Delete Server (Admin)',
  'admin.server.resources': 'Update Resources (Admin)',
  'admin.server.transfer': 'Transfer Server',
  'admin.node.create': 'Create Node',
  'admin.node.delete': 'Delete Node',
  'admin.node.reset_token': 'Reset Node Token',
  'admin.package.create': 'Create Package',
  'admin.package.update': 'Update Package',
  'admin.package.delete': 'Delete Package',
  'admin.ipban.create': 'Ban IP',
  'admin.ipban.delete': 'Unban IP',
  'admin.settings.registration': 'Toggle Registration',
  'admin.settings.server_creation': 'Toggle Server Creation',
  'admin.database_host.create': 'Create Database Host',
  'admin.database_host.update': 'Update Database Host',
  'admin.database_host.delete': 'Delete Database Host',
  'admin.database.delete': 'Delete Database (Admin)',
};

const getActionIcon = (action: string) => {
  if (action.includes('login') || action.includes('register')) return <Icons.key className="w-4 h-4 text-emerald-400" />;
  if (action.includes('logout')) return <Icons.logout className="w-4 h-4 text-neutral-400" />;
  if (action.includes('password') || action.includes('session')) return <Icons.shield className="w-4 h-4 text-amber-400" />;
  if (action.includes('profile')) return <Icons.edit className="w-4 h-4 text-blue-400" />;
  if (action.includes('start')) return <Icons.play className="w-4 h-4 text-emerald-400" />;
  if (action.includes('stop') || action.includes('kill')) return <Icons.stopFilled className="w-4 h-4 text-red-400" />;
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
  if (action.includes('node')) return <Icons.server className="w-4 h-4 text-sky-400" />;
  if (action.includes('package')) return <Icons.cube className="w-4 h-4 text-amber-400" />;
  if (action.includes('ipban')) return <Icons.noAccess className="w-4 h-4 text-red-400" />;
  if (action.includes('user') && action.includes('ban')) return <Icons.noAccess className="w-4 h-4 text-red-400" />;
  if (action.includes('user') && action.includes('unban')) return <Icons.check className="w-4 h-4 text-emerald-400" />;
  if (action.includes('user')) return <Icons.users className="w-4 h-4 text-blue-400" />;
  if (action.includes('settings')) return <Icons.cog className="w-4 h-4 text-neutral-400" />;
  if (action.includes('suspend')) return <Icons.noAccess className="w-4 h-4 text-amber-400" />;
  if (action.includes('unsuspend')) return <Icons.check className="w-4 h-4 text-emerald-400" />;
  if (action.includes('transfer')) return <Icons.move className="w-4 h-4 text-blue-400" />;
  if (action.includes('view')) return <Icons.eye className="w-4 h-4 text-neutral-400" />;
  if (action.includes('create')) return <Icons.plus className="w-4 h-4 text-emerald-400" />;
  if (action.includes('delete')) return <Icons.trash className="w-4 h-4 text-red-400" />;
  if (action.includes('server')) return <Icons.server className="w-4 h-4 text-sky-400" />;
  return <Icons.activity className="w-4 h-4 text-neutral-400" />;
};

export default function LogsPage() {
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState('');
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [searchInput, setSearchInput] = useState('');
  const [filter, setFilter] = useState<Filter>('all');
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const requestId = useRef(0);

  const load = async (p: number, pp: number, s: string, f: Filter, from: string, to: string, initial = false) => {
    const currentRequest = ++requestId.current;
    setLoading(true);
    const res = await adminGetLogs(p, pp, s, f, from, to);
    if (currentRequest !== requestId.current) return;
    if (res.success && res.data) {
      setLogs(res.data.logs || []);
      setPage(res.data.page);
      setTotalPages(res.data.total_pages);
      setTotal(res.data.total);
    } else {
      notify('Error', res.error || 'Failed to load logs', 'error');
    }
    setLoading(false);
    if (initial) { setReady(true); finishLoading(); }
  };

  useEffect(() => {
    startLoading();
    load(1, perPage, '', 'all', '', '', true);
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
    load(1, perPage, searchInput, filter, fromDate, toDate);
  };

  const handleFilterChange = (f: Filter) => {
    setFilter(f);
    load(1, perPage, search, f, fromDate, toDate);
  };

  const clearFilters = () => {
    setSearchInput('');
    setSearch('');
    setFilter('all');
    setFromDate('');
    setToDate('');
    load(1, perPage, '', 'all', '', '');
  };

  const handleExport = async () => {
    setExporting(true);
    const allLogs: ActivityLog[] = [];
    let currentPage = 1;
    let totalPages = 1;
    
    while (currentPage <= totalPages) {
      const res = await adminGetLogs(currentPage, 100, search, filter, fromDate, toDate);
      if (!res.success || !res.data?.logs) {
        notify('Error', 'Failed to export logs', 'error');
        setExporting(false);
        return;
      }
      allLogs.push(...res.data.logs);
      totalPages = res.data.total_pages;
      currentPage++;
    }

    const csvRows = [['Time', 'User', 'Admin', 'Action', 'Description', 'IP'].join(',')];
    allLogs.forEach(log => {
      csvRows.push([
        new Date(log.created_at).toISOString(),
        `"${log.username}"`,
        log.is_admin ? 'Yes' : 'No',
        `"${actionLabels[log.action] || log.action}"`,
        `"${(log.description || '').replace(/"/g, '""')}"`,
        log.ip,
      ].join(','));
    });
    const blob = new Blob([csvRows.join('\n')], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `activity-logs-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
    setExporting(false);
    notify('Exported', `${allLogs.length} log entries exported`, 'success');
  };

  if (!ready) return null;

  const filterLabels = { all: 'All Logs', admin: 'Admin Actions', user: 'User Actions' };
  const hasFilters = search || filter !== 'all' || fromDate || toDate;

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
          <div className="flex items-center gap-1.5">
            <span className="text-sm text-neutral-300">{log.username}</span>
            {log.is_admin && <span className="inline-flex items-center rounded-md bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-medium text-amber-400 ring-1 ring-inset ring-amber-500/20">Admin</span>}
          </div>
        </div>
      )
    },
    { key: 'ip', header: 'IP', render: (log: ActivityLog) => <span className="text-sm text-neutral-400 font-mono">{log.ip}</span> },
    { key: 'time', header: 'Time', render: (log: ActivityLog) => <span className="text-sm text-neutral-400">{new Date(log.created_at).toLocaleString()}</span> },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Activity Logs</h1>
          <p className="text-sm text-neutral-400">View all user and admin activity across the platform.</p>
        </div>
        <Button onClick={handleExport} loading={exporting} variant="secondary" className="w-full sm:w-auto"><Icons.download className="w-4 h-4" />Export CSV</Button>
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
            <form onSubmit={handleSearch} className="flex-1 sm:max-w-sm">
              <Input placeholder="Search by user or action..." value={searchInput} onChange={e => setSearchInput(e.target.value)} />
            </form>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                  <Icons.filter className="w-4 h-4 text-neutral-400" />
                  {filterLabels[filter]}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onSelect={() => handleFilterChange('all')}>All Logs</DropdownMenuItem>
                <DropdownMenuItem onSelect={() => handleFilterChange('admin')}>Admin Actions</DropdownMenuItem>
                <DropdownMenuItem onSelect={() => handleFilterChange('user')}>User Actions</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <Pagination page={page} totalPages={totalPages} total={total} perPage={perPage} onPageChange={p => load(p, perPage, search, filter, fromDate, toDate)} onPerPageChange={pp => { setPerPage(pp); load(1, pp, search, filter, fromDate, toDate); }} loading={loading} />
        </div>

        <div className="flex items-center gap-3">
          <DatePicker label="From:" value={fromDate} onChange={v => { setFromDate(v); load(1, perPage, search, filter, v, toDate); }} />
          <DatePicker label="To:" value={toDate} onChange={v => { setToDate(v); load(1, perPage, search, filter, fromDate, v); }} />
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
            emptyText="No logs found"
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
