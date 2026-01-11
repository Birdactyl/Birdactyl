import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminGetServers, adminViewServer, adminGetServerCreationStatus, adminSetServerCreationStatus, adminGetTransferStatus, adminGetAllTransfers, adminGetNodes, getAvailablePackages, type Server, type Node, type TransferStatus, type Package, startServer, stopServer, killServer } from '../../../lib/api';
import { useTable } from '../../../hooks/useTable';
import { notify, Button, Input, Checkbox, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, BulkActionBar, Table, Pagination } from '../../../components';
import { EditServerModal, AdminCreateServerModal, ConfirmActionModal, BulkTransferModal, TransferProgressModal } from '../../../components/modals';

type Filter = 'all' | 'running' | 'stopped' | 'suspended';

const filterServer = (s: Server, search: string, filter: Filter): boolean => {
  if (search) {
    const q = search.toLowerCase();
    if (!s.name.toLowerCase().includes(q) && !s.user?.username?.toLowerCase().includes(q) && !s.id.toLowerCase().includes(q)) return false;
  }
  if (filter === 'running') return s.status === 'running' && !s.is_suspended;
  if (filter === 'stopped') return s.status !== 'running' && !s.is_suspended;
  if (filter === 'suspended') return s.is_suspended;
  return true;
};

export default function ServersPage() {
  const navigate = useNavigate();
  const filterFn = useCallback(filterServer, []);
  const table = useTable<Server, Filter>({
    mode: 'client',
    fetchFn: async () => {
      const res = await adminGetServers();
      return { success: res.success, data: res.data, error: res.error };
    },
    filterFn,
    defaultFilter: 'all',
  });

  const [editServer, setEditServer] = useState<Server | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [confirmAction, setConfirmAction] = useState<{ type: 'suspend' | 'unsuspend' | 'delete'; ids: string[] } | null>(null);
  const [bulkTransferIds, setBulkTransferIds] = useState<string[] | null>(null);
  const [activeTransfer, setActiveTransfer] = useState<TransferStatus | null>(null);
  const [showTransferProgress, setShowTransferProgress] = useState(false);
  const [serverCreationEnabled, setServerCreationEnabled] = useState(true);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [packages, setPackages] = useState<Package[]>([]);
  const pollRef = useRef<number | null>(null);

  useEffect(() => { adminGetServerCreationStatus().then(res => { if (res.success && res.data) setServerCreationEnabled(res.data.enabled); }); }, []);
  useEffect(() => { adminGetNodes().then(res => { if (res.success && res.data) setNodes(res.data); }); }, []);
  useEffect(() => { getAvailablePackages().then(res => { if (res.success && res.data) setPackages(res.data); }); }, []);
  useEffect(() => { return () => { if (pollRef.current) clearInterval(pollRef.current); }; }, []);

  useEffect(() => {
    adminGetAllTransfers().then(res => {
      if (res.success && res.data) {
        const active = res.data.find(t => t.stage !== 'complete' && t.stage !== 'failed');
        if (active) startTransferPolling(active);
      }
    });
  }, []);

  const startTransferPolling = (initial: TransferStatus) => {
    setActiveTransfer(initial);
    setShowTransferProgress(true);
    if (pollRef.current) clearInterval(pollRef.current);
    pollRef.current = window.setInterval(async () => {
      const res = await adminGetTransferStatus(initial.id);
      if (res.success && res.data) {
        setActiveTransfer(res.data);
        if (res.data.stage === 'complete' || res.data.stage === 'failed') {
          if (pollRef.current) clearInterval(pollRef.current);
          if (res.data.stage === 'complete') { notify('Success', 'Server transferred successfully', 'success'); table.reload(); }
          else notify('Error', res.data.error || 'Transfer failed', 'error');
        }
      }
    }, 1000);
  };

  const toggleServerCreation = async () => {
    const newVal = !serverCreationEnabled;
    const res = await adminSetServerCreationStatus(newVal);
    if (res.success) {
      setServerCreationEnabled(newVal);
      notify('Success', newVal ? 'Server creation enabled' : 'Server creation disabled', 'success');
    } else {
      notify('Error', res.error || 'Failed to update setting', 'error');
    }
  };

  const handlePower = async (server: Server, action: 'start' | 'stop' | 'kill') => {
    const fn = action === 'start' ? startServer : action === 'stop' ? stopServer : killServer;
    const res = await fn(server.id);
    if (res.success) { notify('Success', `Server ${action} signal sent`, 'success'); table.reload(); }
    else notify('Error', res.error || `Failed to ${action} server`, 'error');
  };

  const viewServer = async (server: Server) => {
    const res = await adminViewServer(server.id);
    if (res.success) navigate(`/console/server/${server.id}`);
    else notify('Error', res.error || 'Failed to view server', 'error');
  };

  if (!table.ready) return null;

  const selectedServers = table.items.filter(s => table.selected.has(s.id));
  const hasSelectedSuspended = selectedServers.some(s => s.is_suspended);
  const hasSelectedUnsuspended = selectedServers.some(s => !s.is_suspended);
  const filterLabels = { all: 'All Servers', running: 'Running', stopped: 'Stopped', suspended: 'Suspended' };

  const columns = [
    { key: 'select', header: <Checkbox checked={table.allSelected} indeterminate={table.someSelected} onChange={table.toggleSelectAll} />, className: 'w-12', render: (server: Server) => <Checkbox checked={table.selected.has(server.id)} onChange={() => table.toggleSelect(server.id)} /> },
    { key: 'server', header: 'Server', render: (server: Server) => (
      <div className="flex items-center gap-3">
        <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${server.status === 'running' && !server.is_suspended ? 'bg-emerald-500/20' : 'bg-neutral-700'}`}>
          <Icons.server className={`w-4 h-4 ${server.status === 'running' && !server.is_suspended ? 'text-emerald-400' : 'text-neutral-400'}`} />
        </div>
        <div>
          <div className="text-sm font-medium text-neutral-100">{server.name}</div>
          <div className="text-xs text-neutral-500">{server.id.slice(0, 8)}</div>
        </div>
      </div>
    )},
    { key: 'owner', header: 'Owner', render: (server: Server) => <span className="text-sm text-neutral-300">{server.user?.username || '—'}</span> },
    { key: 'node', header: 'Node', render: (server: Server) => <span className="text-sm text-neutral-400">{server.node?.name || '—'}</span> },
    { key: 'resources', header: 'Resources', render: (server: Server) => (
      <div className="flex items-center gap-3 text-xs text-neutral-400">
        <span>{server.memory} MB</span>
        <span className="text-neutral-600">•</span>
        <span>{server.cpu}%</span>
        <span className="text-neutral-600">•</span>
        <span>{server.disk} MB</span>
      </div>
    )},
    { key: 'status', header: 'Status', render: (server: Server) => {
      const isTransferring = activeTransfer && activeTransfer.server_id === server.id && activeTransfer.stage !== 'complete' && activeTransfer.stage !== 'failed';
      if (isTransferring) return (
        <div className="flex items-center gap-2 cursor-pointer" onClick={() => setShowTransferProgress(true)}>
          <div className="w-16 h-1.5 bg-neutral-700 rounded-full overflow-hidden">
            <div className="h-full bg-blue-500 transition-all duration-300" style={{ width: `${activeTransfer.progress}%` }} />
          </div>
          <span className="text-xs text-blue-400">{activeTransfer.progress}%</span>
        </div>
      );
      if (server.is_suspended) return <span className="inline-flex items-center rounded-md bg-amber-500/10 px-2 py-1 text-xs font-medium text-amber-400 ring-1 ring-inset ring-amber-500/20">Suspended</span>;
      if (server.status === 'running') return <span className="inline-flex items-center rounded-md bg-emerald-500/10 px-2 py-1 text-xs font-medium text-emerald-400">Online</span>;
      return <span className="inline-flex items-center rounded-md bg-neutral-500/10 px-2 py-1 text-xs font-medium text-neutral-400">Offline</span>;
    }},
    { key: 'actions', header: '', align: 'right' as const, render: (server: Server) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => viewServer(server)}>View</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setEditServer(server)}>Edit</DropdownMenuItem>
          {!server.is_suspended && server.status !== 'running' && <DropdownMenuItem onSelect={() => handlePower(server, 'start')}>Start</DropdownMenuItem>}
          {!server.is_suspended && server.status === 'running' && <DropdownMenuItem onSelect={() => handlePower(server, 'stop')}>Stop</DropdownMenuItem>}
          {!server.is_suspended && server.status === 'running' && <DropdownMenuItem onSelect={() => handlePower(server, 'kill')}>Kill</DropdownMenuItem>}
          {server.is_suspended ? (
            <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'unsuspend', ids: [server.id] })}>Unsuspend</DropdownMenuItem>
          ) : (
            <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'suspend', ids: [server.id] })}>Suspend</DropdownMenuItem>
          )}
          <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'delete', ids: [server.id] })} className="text-red-400">Delete</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )},
  ];

  return (
    <>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">Servers</h1>
            <p className="text-sm text-neutral-400">Manage all servers across the platform.</p>
          </div>
          <Button onClick={() => setShowCreate(true)} className="w-full sm:w-auto"><Icons.plus className="w-4 h-4" />Create Server</Button>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
            <form onSubmit={table.handleSearch} className="flex-1 sm:max-w-sm">
              <Input placeholder="Search by name, owner, or ID..." value={table.searchInput} onChange={e => table.setSearchInput(e.target.value)} />
            </form>
            <div className="flex items-center gap-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    <Icons.filter className="w-4 h-4 text-neutral-400" />
                    {filterLabels[table.filter]}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('all')}>All Servers</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('running')}>Running</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('stopped')}>Stopped</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('suspended')}>Suspended</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    Server Creation: {serverCreationEnabled ? 'On' : 'Off'}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onSelect={() => !serverCreationEnabled && toggleServerCreation()}>On</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => serverCreationEnabled && toggleServerCreation()}>Off</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
          <Pagination page={table.page} totalPages={table.totalPages} total={table.total} perPage={table.perPage} onPageChange={table.setPage} onPerPageChange={table.handlePerPageChange} loading={table.loading} />
        </div>

        <BulkActionBar count={table.selected.size} onClear={table.clearSelection}>
          {hasSelectedUnsuspended && <button onClick={() => setConfirmAction({ type: 'suspend', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-amber-400 hover:bg-amber-500/10 transition-colors">Suspend</button>}
          {hasSelectedSuspended && <button onClick={() => setConfirmAction({ type: 'unsuspend', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-green-400 hover:bg-green-500/10 transition-colors">Unsuspend</button>}
          <button onClick={() => setBulkTransferIds(Array.from(table.selected))} className="text-xs font-medium px-3 py-1.5 rounded-lg text-blue-400 hover:bg-blue-500/10 transition-colors">Transfer</button>
          <button onClick={() => setConfirmAction({ type: 'delete', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors">Delete</button>
        </BulkActionBar>

        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-4 py-2 text-xs text-neutral-400">{table.total} servers</div>
          <div className="bg-neutral-900/40 rounded-lg p-1">
            <Table columns={columns} data={table.items} keyField="id" loading={table.loading} emptyText="No servers found" rowClassName={server => table.selected.has(server.id) ? 'bg-neutral-800/20' : ''} />
          </div>
        </div>
      </div>

      <EditServerModal
        open={!!editServer}
        server={editServer}
        nodes={nodes}
        onClose={() => setEditServer(null)}
        onSaved={table.reload}
        onTransferStarted={startTransferPolling}
      />

      <AdminCreateServerModal
        open={showCreate}
        nodes={nodes}
        packages={packages}
        onClose={() => setShowCreate(false)}
        onCreated={table.reload}
      />

      <ConfirmActionModal
        open={!!confirmAction}
        type={confirmAction?.type || 'suspend'}
        ids={confirmAction?.ids || []}
        onClose={() => setConfirmAction(null)}
        onComplete={() => { table.reload(); table.clearSelection(); }}
      />

      <BulkTransferModal
        open={!!bulkTransferIds}
        serverIds={bulkTransferIds || []}
        nodes={nodes}
        getServerNodeId={id => table.allItems.find(s => s.id === id)?.node_id}
        onClose={() => setBulkTransferIds(null)}
        onComplete={() => { table.clearSelection(); setTimeout(table.reload, 2000); }}
      />

      <TransferProgressModal
        open={showTransferProgress && !!activeTransfer}
        transfer={activeTransfer || { id: '', server_id: '', server_name: '', from_node_id: '', from_node_name: '', to_node_id: '', to_node_name: '', stage: 'pending', progress: 0, started_at: '' }}
        onClose={() => {
          setShowTransferProgress(false);
          if (activeTransfer?.stage === 'complete' || activeTransfer?.stage === 'failed') setActiveTransfer(null);
        }}
      />
    </>
  );
}
