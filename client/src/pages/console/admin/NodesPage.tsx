import { useEffect, useState, useMemo } from 'react';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { adminGetNodes, adminCreateNode, adminDeleteNode, adminResetNodeToken, adminRefreshNodes, adminGetPairingCode, adminPairNode, adminUpdateNode } from '../../../lib/api';
import { notify, Button, Input, Modal, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, Table, Pagination } from '../../../components';

interface Node {
  id: string; name: string; fqdn: string; port: number; is_online: boolean; auth_error: boolean; last_heartbeat: string | null; icon?: string;
  system_info: { hostname: string; os: { name: string; version: string; kernel: string; arch: string }; cpu: { cores: number; usage_percent: number }; memory: { total_bytes: number; used_bytes: number; available_bytes: number; usage_percent: number }; disk: { total_bytes: number; used_bytes: number; available_bytes: number; usage_percent: number }; uptime_seconds: number };
  created_at: string;
}

interface NodeToken { token_id: string; token: string; }
type Filter = 'all' | 'online' | 'offline';

const formatTimeAgo = (date: string) => {
  const s = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
  return s < 60 ? 'just now' : s < 3600 ? `${Math.floor(s / 60)}m ago` : s < 86400 ? `${Math.floor(s / 3600)}h ago` : `${Math.floor(s / 86400)}d ago`;
};

const formatBytes = (bytes: number) => {
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
};

function UsageBar({ value, color }: { value: number; color: string }) {
  return (
    <div className="flex items-center gap-2">
      <div className="w-16 bg-neutral-700 rounded-full h-1.5">
        <div className={`${color} h-1.5 rounded-full transition-all`} style={{ width: `${Math.min(value, 100)}%` }} />
      </div>
      <span className="text-xs text-neutral-400 w-8">{value.toFixed(0)}%</span>
    </div>
  );
}

export default function NodesPage() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [ui, setUi] = useState({ ready: false, refreshing: false, search: '', filter: 'all' as Filter, page: 1, perPage: 20 });
  const [createModal, setCreateModal] = useState({ open: false, loading: false, name: '', fqdn: '', port: '8443', icon: '' });
  const [pairModal, setPairModal] = useState({ open: false, loading: false, name: '', fqdn: '', port: '8443', code: '', stage: 'form' as 'form' | 'waiting', error: '', icon: '' });
  const [tokenModal, setTokenModal] = useState<{ node: Node; token: NodeToken } | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ node: Node; loading: boolean } | null>(null);
  const [editModal, setEditModal] = useState<{ open: boolean; loading: boolean; node: Node | null; name: string; icon: string }>({ open: false, loading: false, node: null, name: '', icon: '' });
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const filtered = useMemo(() => {
    return nodes.filter(n => {
      if (ui.search && !n.name.toLowerCase().includes(ui.search.toLowerCase()) && !n.fqdn.toLowerCase().includes(ui.search.toLowerCase()) && !n.id.toLowerCase().includes(ui.search.toLowerCase())) return false;
      if (ui.filter === 'online') return n.is_online;
      if (ui.filter === 'offline') return !n.is_online;
      return true;
    });
  }, [nodes, ui.search, ui.filter]);

  const paginated = useMemo(() => {
    const start = (ui.page - 1) * ui.perPage;
    return filtered.slice(start, start + ui.perPage);
  }, [filtered, ui.page, ui.perPage]);

  const totalPages = Math.ceil(filtered.length / ui.perPage) || 1;

  useEffect(() => { startLoading(); loadNodes(true); }, []);

  const loadNodes = async (initial = false) => {
    if (!initial) setUi(s => ({ ...s, refreshing: true }));
    const res = initial ? await adminGetNodes() : await adminRefreshNodes();
    if (res.success && res.data) setNodes(res.data);
    setUi(s => ({ ...s, refreshing: false, ready: initial ? true : s.ready }));
    if (initial) finishLoading();
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateModal(s => ({ ...s, loading: true }));
    const res = await adminCreateNode(createModal.name, createModal.fqdn, parseInt(createModal.port) || 8443, createModal.icon || undefined);
    if (res.success && res.data) {
      setTokenModal({ node: res.data.node, token: res.data.token });
      setCreateModal({ open: false, loading: false, name: '', fqdn: '', port: '8443', icon: '' });
      loadNodes();
    } else {
      notify('Failed', res.error || 'Could not create node', 'error');
      setCreateModal(s => ({ ...s, loading: false }));
    }
  };

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editModal.node) return;
    setEditModal(s => ({ ...s, loading: true }));
    const res = await adminUpdateNode(editModal.node.id, { name: editModal.name, icon: editModal.icon });
    if (res.success) {
      notify('Updated', 'Node updated successfully', 'success');
      setEditModal({ open: false, loading: false, node: null, name: '', icon: '' });
      loadNodes();
    } else {
      notify('Failed', res.error || 'Could not update node', 'error');
      setEditModal(s => ({ ...s, loading: false }));
    }
  };

  const handleStartPairing = async (e: React.FormEvent) => {
    e.preventDefault();
    const name = pairModal.name;
    const fqdn = pairModal.fqdn;
    const port = parseInt(pairModal.port) || 8443;
    
    setPairModal(s => ({ ...s, loading: true, error: '' }));
    const codeRes = await adminGetPairingCode();
    if (!codeRes.success || !codeRes.data) {
      notify('Failed', 'Could not generate pairing code', 'error');
      setPairModal(s => ({ ...s, loading: false }));
      return;
    }
    const code = codeRes.data.code;
    setPairModal(s => ({ ...s, code, stage: 'waiting', loading: false }));

    const res = await adminPairNode(name, fqdn, port, code);
    if (res.success && res.data) {
      notify('Paired', `Node ${res.data.node.name} connected successfully`, 'success');
      setPairModal({ open: false, loading: false, name: '', fqdn: '', port: '8443', code: '', stage: 'form', error: '', icon: '' });
      loadNodes();
    } else {
      setPairModal(s => ({ ...s, error: res.error || 'Could not pair with node' }));
    }
  };

  const handleDelete = async () => {
    if (!deleteModal) return;
    setDeleteModal(s => s && { ...s, loading: true });
    const res = await adminDeleteNode(deleteModal.node.id);
    res.success ? notify('Node deleted', `${deleteModal.node.name} has been removed`, 'success') : notify('Failed', res.error || 'Could not delete node', 'error');
    setDeleteModal(null);
    if (res.success) loadNodes();
  };

  const handleResetToken = async (node: Node) => {
    const res = await adminResetNodeToken(node.id);
    if (res.success && res.data) { setTokenModal({ node, token: res.data }); notify('Token reset', 'New token generated', 'success'); }
    else notify('Failed', res.error || 'Could not reset token', 'error');
  };

  const copyToken = () => { if (tokenModal) { navigator.clipboard.writeText(`${tokenModal.token.token_id}.${tokenModal.token.token}`); notify('Copied', 'Token copied to clipboard', 'success'); } };

  const filterLabels = { all: 'All Nodes', online: 'Online', offline: 'Offline' };

  if (!ui.ready) return null;

  const columns = [
    { key: 'expand', header: '', className: 'w-8', render: (node: Node) => (
      <button onClick={() => { const next = new Set(expanded); next.has(node.id) ? next.delete(node.id) : next.add(node.id); setExpanded(next); }} className="text-neutral-400 hover:text-neutral-200 transition">
        <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(node.id) ? 'rotate-90' : ''}`} />
      </button>
    )},
    { key: 'node', header: 'Node', render: (node: Node) => (
      <div className="flex items-center gap-3">
        {node.icon ? (
          <img src={node.icon} alt="" className="w-8 h-8 rounded-lg object-cover" />
        ) : (
          <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${node.is_online ? 'bg-emerald-500/20' : 'bg-neutral-700'}`}>
            <Icons.server className={`w-4 h-4 ${node.is_online ? 'text-emerald-400' : 'text-neutral-400'}`} />
          </div>
        )}
        <div>
          <div className="text-sm font-medium text-neutral-100">{node.name}</div>
          <div className="text-xs text-neutral-500">{node.fqdn}:{node.port}</div>
        </div>
      </div>
    )},
    { key: 'status', header: 'Status', render: (node: Node) => node.auth_error ? (
      <span className="inline-flex items-center rounded-md bg-red-500/10 px-2 py-1 text-xs font-medium text-red-400 ring-1 ring-inset ring-red-500/20">Auth Error</span>
    ) : node.is_online ? (
      <span className="inline-flex items-center rounded-md bg-emerald-500/10 px-2 py-1 text-xs font-medium text-emerald-400 ring-1 ring-inset ring-emerald-500/20">Online</span>
    ) : (
      <span className="inline-flex items-center rounded-md bg-neutral-500/10 px-2 py-1 text-xs font-medium text-neutral-400 ring-1 ring-inset ring-neutral-500/20">Offline</span>
    )},
    { key: 'cpu', header: 'CPU', render: (node: Node) => {
      const hasInfo = node.is_online && node.system_info?.hostname && !node.auth_error;
      return hasInfo ? <UsageBar value={node.system_info.cpu.usage_percent} color={node.system_info.cpu.usage_percent > 80 ? 'bg-red-400' : 'bg-sky-400'} /> : <span className="text-xs text-neutral-600">—</span>;
    }},
    { key: 'memory', header: 'Memory', render: (node: Node) => {
      const hasInfo = node.is_online && node.system_info?.hostname && !node.auth_error;
      return hasInfo ? <UsageBar value={node.system_info.memory.usage_percent} color={node.system_info.memory.usage_percent > 80 ? 'bg-red-400' : 'bg-violet-400'} /> : <span className="text-xs text-neutral-600">—</span>;
    }},
    { key: 'disk', header: 'Disk', render: (node: Node) => {
      const hasInfo = node.is_online && node.system_info?.hostname && !node.auth_error;
      return hasInfo ? <UsageBar value={node.system_info.disk.usage_percent} color={node.system_info.disk.usage_percent > 80 ? 'bg-red-400' : 'bg-amber-400'} /> : <span className="text-xs text-neutral-600">—</span>;
    }},
    { key: 'lastseen', header: 'Last Seen', render: (node: Node) => <span className="text-sm text-neutral-400">{node.last_heartbeat ? formatTimeAgo(node.last_heartbeat) : 'Never'}</span> },
    { key: 'actions', header: 'Actions', align: 'right' as const, render: (node: Node) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setEditModal({ open: true, loading: false, node, name: node.name, icon: node.icon || '' })}>Edit</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => handleResetToken(node)}>Reset Token</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setDeleteModal({ node, loading: false })} className="text-red-400">Delete</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )},
  ];

  return (
    <>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div><h1 className="text-xl font-semibold text-neutral-100">Nodes</h1><p className="text-sm text-neutral-400">Manage server nodes running Birdactyl Axis.</p></div>
          <div className="flex items-center gap-2">
            <Button variant="secondary" onClick={() => setPairModal(s => ({ ...s, open: true }))} className="flex-1 sm:flex-none">Pair Node</Button>
            <Button onClick={() => setCreateModal(s => ({ ...s, open: true }))} className="flex-1 sm:flex-none"><Icons.plus className="w-4 h-4" />Add Node</Button>
          </div>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
            <form onSubmit={e => e.preventDefault()} className="flex-1 sm:max-w-sm">
              <Input placeholder="Search by name, FQDN, or ID..." value={ui.search} onChange={e => setUi(s => ({ ...s, search: e.target.value, page: 1 }))} />
            </form>
            <div className="flex items-center gap-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    <Icons.filter className="w-4 h-4 text-neutral-400" />
                    {filterLabels[ui.filter]}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onSelect={() => setUi(s => ({ ...s, filter: 'all', page: 1 }))}>All Nodes</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => setUi(s => ({ ...s, filter: 'online', page: 1 }))}>Online</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => setUi(s => ({ ...s, filter: 'offline', page: 1 }))}>Offline</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <button onClick={() => loadNodes()} disabled={ui.refreshing} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 transition-colors disabled:opacity-50" title="Refresh">
                <Icons.refresh className={`w-4 h-4 ${ui.refreshing ? 'animate-spin' : ''}`} />
              </button>
            </div>
          </div>
          <Pagination page={ui.page} totalPages={totalPages} total={filtered.length} perPage={ui.perPage} onPageChange={p => setUi(s => ({ ...s, page: p }))} onPerPageChange={pp => setUi(s => ({ ...s, perPage: pp, page: 1 }))} loading={ui.refreshing} />
        </div>

        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-4 py-2 text-xs text-neutral-400">{filtered.length} node{filtered.length !== 1 ? 's' : ''}</div>
          <div className="bg-neutral-900/40 rounded-lg p-1">
            <Table
              columns={columns}
              data={paginated}
              keyField="id"
              loading={ui.refreshing}
              emptyText={nodes.length === 0 ? 'No nodes yet' : 'No nodes found'}
              expandable={{
                isExpanded: node => expanded.has(node.id),
                render: node => {
                  const hasInfo = node.is_online && node.system_info?.hostname && !node.auth_error;
                  if (node.auth_error) return <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20"><p className="text-xs text-red-400">Token mismatch. Reset the token and update Axis configuration.</p></div>;
                  if (!hasInfo) return <p className="text-xs text-neutral-500">No system information available. Node may be offline or not yet connected.</p>;
                  return (
                    <div className="grid grid-cols-4 gap-6 text-xs">
                      <div><div className="text-neutral-500 mb-1">System</div><div className="text-neutral-200">{node.system_info.os.name} {node.system_info.os.version}</div><div className="text-neutral-400">{node.system_info.os.arch}</div></div>
                      <div><div className="text-neutral-500 mb-1">CPU</div><div className="text-neutral-200">{node.system_info.cpu.cores} cores</div><div className="text-neutral-400">{node.system_info.cpu.usage_percent.toFixed(1)}% usage</div></div>
                      <div><div className="text-neutral-500 mb-1">Memory</div><div className="text-neutral-200">{formatBytes(node.system_info.memory.used_bytes)} / {formatBytes(node.system_info.memory.total_bytes)}</div><div className="text-neutral-400">{formatBytes(node.system_info.memory.available_bytes)} available</div></div>
                      <div><div className="text-neutral-500 mb-1">Disk</div><div className="text-neutral-200">{formatBytes(node.system_info.disk.used_bytes)} / {formatBytes(node.system_info.disk.total_bytes)}</div><div className="text-neutral-400">{formatBytes(node.system_info.disk.available_bytes)} available</div></div>
                    </div>
                  );
                }
              }}
            />
          </div>
        </div>
      </div>

      <Modal open={createModal.open} onClose={() => setCreateModal(s => ({ ...s, open: false }))} title="Add Node" description="Create a new node to connect a server.">
        <form onSubmit={handleCreate} className="space-y-4">
          <Input label="Name" placeholder="us-east-1" value={createModal.name} onChange={e => setCreateModal(s => ({ ...s, name: e.target.value }))} required />
          <Input label="FQDN" placeholder="node.example.com" value={createModal.fqdn} onChange={e => setCreateModal(s => ({ ...s, fqdn: e.target.value }))} required />
          <Input label="Port" type="number" placeholder="8443" value={createModal.port} onChange={e => setCreateModal(s => ({ ...s, port: e.target.value }))} />
          <Input label="Icon URL (optional)" placeholder="https://example.com/flag.png" value={createModal.icon} onChange={e => setCreateModal(s => ({ ...s, icon: e.target.value }))} />
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setCreateModal(s => ({ ...s, open: false }))}>Cancel</Button>
            <Button type="submit" loading={createModal.loading}>Create</Button>
          </div>
        </form>
      </Modal>

      <Modal open={editModal.open} onClose={() => !editModal.loading && setEditModal(s => ({ ...s, open: false }))} title="Edit Node" description="Update node name and icon.">
        <form onSubmit={handleEdit} className="space-y-4">
          <Input label="Name" placeholder="us-east-1" value={editModal.name} onChange={e => setEditModal(s => ({ ...s, name: e.target.value }))} required />
          <Input label="Icon URL (optional)" placeholder="https://example.com/flag.png" value={editModal.icon} onChange={e => setEditModal(s => ({ ...s, icon: e.target.value }))} />
          {editModal.icon && (
            <div className="flex items-center gap-2">
              <span className="text-xs text-neutral-400">Preview:</span>
              <img src={editModal.icon} alt="" className="w-8 h-8 rounded-lg object-cover" onError={e => (e.target as HTMLImageElement).style.display = 'none'} />
            </div>
          )}
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setEditModal(s => ({ ...s, open: false }))} disabled={editModal.loading}>Cancel</Button>
            <Button type="submit" loading={editModal.loading}>Save</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!tokenModal} onClose={() => setTokenModal(null)} title="Node Token" description="Save this token - it won't be shown again.">
        <div className="space-y-4">
          <div className="rounded-lg bg-neutral-800 p-4">
            <div className="text-xs text-neutral-400 mb-2">Configuration for Birdactyl Axis</div>
            <pre className="text-xs text-neutral-200 font-mono overflow-x-auto whitespace-pre-wrap break-all">{`panel:\n  url: ${window.location.origin}\n  token: ${tokenModal?.token.token_id}.${tokenModal?.token.token}`}</pre>
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={copyToken}><Icons.clipboard className="w-4 h-4" />Copy Token</Button>
            <Button onClick={() => setTokenModal(null)}>Done</Button>
          </div>
        </div>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Delete Node" description={`Are you sure you want to delete "${deleteModal?.node.name}"? This action cannot be undone.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button variant="danger" onClick={handleDelete} loading={deleteModal?.loading}>Delete</Button>
        </div>
      </Modal>

      <Modal open={pairModal.open} onClose={() => !pairModal.loading && setPairModal(s => ({ ...s, open: false, stage: 'form', code: '', error: '' }))} title="Pair Node" description={pairModal.stage === 'form' ? 'Connect to a node running in pairing mode.' : 'Waiting for confirmation on the node...'}>
        {pairModal.stage === 'form' ? (
          <form onSubmit={handleStartPairing} className="space-y-4">
            <div className="rounded-lg bg-sky-500/10 border border-sky-500/20 p-3">
              <p className="text-xs text-sky-400">Run <code className="bg-neutral-800 px-1.5 py-0.5 rounded font-mono">axis pair</code> on the node first, then fill in the details below.</p>
            </div>
            <Input label="Name" placeholder="us-east-1" value={pairModal.name} onChange={e => setPairModal(s => ({ ...s, name: e.target.value }))} required />
            <Input label="FQDN" placeholder="node.example.com" value={pairModal.fqdn} onChange={e => setPairModal(s => ({ ...s, fqdn: e.target.value }))} required />
            <Input label="Port" inputMode="numeric" placeholder="8443" value={pairModal.port} onChange={e => setPairModal(s => ({ ...s, port: e.target.value }))} />
            <div className="flex justify-end gap-3 pt-4">
              <Button variant="ghost" onClick={() => setPairModal(s => ({ ...s, open: false }))}>Cancel</Button>
              <Button type="submit" loading={pairModal.loading}>Start Pairing</Button>
            </div>
          </form>
        ) : (
          <div className="space-y-4">
            <div className="text-center py-4">
              <div className="text-xs text-neutral-400 mb-2">Verify this code matches on the node terminal:</div>
              <div className="text-4xl font-mono font-bold text-neutral-100 tracking-widest">{pairModal.code}</div>
            </div>
            <div className="rounded-lg bg-neutral-800 p-3">
              <p className="text-xs text-neutral-400">On the node terminal, type <code className="bg-neutral-700 px-1.5 py-0.5 rounded font-mono">y</code> and press Enter to confirm the pairing.</p>
            </div>
            {pairModal.error ? (
              <div className="space-y-3">
                <div className="rounded-lg bg-red-500/10 border border-red-500/20 p-3">
                  <p className="text-xs text-red-400">{pairModal.error}</p>
                </div>
                <div className="flex justify-end gap-3">
                  <Button variant="ghost" onClick={() => setPairModal(s => ({ ...s, open: false, stage: 'form', code: '', error: '' }))}>Cancel</Button>
                  <Button onClick={() => setPairModal(s => ({ ...s, stage: 'form', code: '', error: '' }))}>Try Again</Button>
                </div>
              </div>
            ) : (
              <>
                <div className="flex items-center justify-center gap-2 text-neutral-400">
                  <span className="inline-block rounded-full border-2 border-current border-t-transparent animate-spin h-4 w-4"></span>
                  <span className="text-sm">Waiting for confirmation...</span>
                </div>
                <div className="flex justify-end pt-2">
                  <Button variant="ghost" onClick={() => setPairModal(s => ({ ...s, open: false, stage: 'form', code: '', error: '' }))}>Cancel</Button>
                </div>
              </>
            )}
          </div>
        )}
      </Modal>
    </>
  );
}
