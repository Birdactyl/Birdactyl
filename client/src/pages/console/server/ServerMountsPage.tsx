import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServerMounts, mountServerMount, unmountServerMount, type ServerMountResponse } from '../../../lib/api';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { notify, Button, Icons, Table, PermissionDenied, ContextMenu } from '../../../components';
import { useServerPermissions } from '../../../hooks/useServerPermissions';

export default function ServerMountsPage() {
  const { id } = useParams<{ id: string }>();
  const [mounts, setMounts] = useState<ServerMountResponse[]>([]);
  const [ready, setReady] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const { can, loading: permsLoading } = useServerPermissions(id);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const load = async (initial = false) => {
    if (!id) return;
    if (initial) startLoading();
    
    const res = await getServerMounts(id);
    if (res.success && res.data) {
      setMounts(res.data);
    } else if (initial) {
      notify('Error', res.error || 'Failed to load mounts', 'error');
    }
    
    if (initial) {
      setReady(true);
      finishLoading();
    }
  };

  useEffect(() => {
    load(true);
  }, [id]);

  const handleToggle = async (mount: ServerMountResponse) => {
    if (!id || actionLoading) return;
    setActionLoading(mount.id);
    
    let res;
    if (mount.is_mounted) {
      res = await unmountServerMount(id, mount.id);
    } else {
      res = await mountServerMount(id, mount.id);
    }
    
    if (res.success) {
      const baseMsg = res.data?.message || `Mount ${mount.is_mounted ? 'detached' : 'attached'} successfully`;
      notify('Success', baseMsg + '. Please start or restart your server for changes to take effect.', 'success');
      setMounts(prev => prev.map(m => m.id === mount.id ? { ...m, is_mounted: !mount.is_mounted } : m));
    } else {
      notify('Error', res.error || `Failed to ${mount.is_mounted ? 'detach' : 'attach'} mount`, 'error');
    }
    
    setActionLoading(null);
  };

  const toggleExpand = (mountId: string) => setExpanded(prev => { const next = new Set(prev); next.has(mountId) ? next.delete(mountId) : next.add(mountId); return next; });

  if (permsLoading || !ready) return null;
  if (!can('mount.read')) return <PermissionDenied message="You don't have permission to view mounts" />;

  const getMountActions = (mount: ServerMountResponse) => [
      ...(can('mount.update') ? [{ label: mount.is_mounted ? 'Detach from Server' : 'Attach to Server', onClick: () => handleToggle(mount), variant: mount.is_mounted ? 'danger' as const : 'default' as const }] : []),
  ];

  const columns = [
    {
      key: 'expand', header: '', className: 'w-8', render: (mount: ServerMountResponse) => (
        <button onClick={() => toggleExpand(mount.id)} className="text-neutral-400 hover:text-neutral-200 transition">
          <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(mount.id) ? 'rotate-90' : ''}`} />
        </button>
      )
    },
    {
      key: 'name', header: 'Mount', render: (mount: ServerMountResponse) => (
        <div className="flex items-center gap-3">
          <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${mount.is_mounted ? 'bg-emerald-500/20' : 'bg-neutral-500/20'}`}>
            <Icons.folder className={`w-4 h-4 ${mount.is_mounted ? 'text-emerald-400' : 'text-neutral-400'}`} />
          </div>
          <div>
            <div className="text-sm font-medium text-neutral-100">{mount.name}</div>
            <div className="text-xs text-neutral-500">{mount.source}</div>
          </div>
        </div>
      )
    },
    {
      key: 'target', header: 'Target Path', render: (mount: ServerMountResponse) => (
        <div className="flex items-center gap-2">
          <code className="text-sm text-neutral-300 font-mono">{mount.target}</code>
        </div>
      )
    },
    {
      key: 'status', header: 'Status', render: (mount: ServerMountResponse) => (
        <div className="flex flex-col">
          <span className={`text-sm ${mount.is_mounted ? 'text-emerald-400 font-medium' : 'text-neutral-500'}`}>
            {mount.is_mounted ? 'Attached' : 'Not Attached'}
          </span>
          <span className="text-xs text-neutral-500 mt-0.5">
            {mount.read_only ? 'Read Only' : 'Read & Write'}
          </span>
        </div>
      )
    },
    {
      key: 'actions', header: '', align: 'right' as const, render: (mount: ServerMountResponse) => (
        <ContextMenu
          align="end"
          trigger={<Button variant="ghost" loading={actionLoading === mount.id} disabled={actionLoading !== null}><Icons.ellipsis className="w-5 h-5" /></Button>}
          items={getMountActions(mount)}
        />
      )
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">Server</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">Mounts</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Mounts</h1>
          <p className="text-sm text-neutral-400">Manage storage mounts available for your server.</p>
        </div>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{mounts.length} mount{mounts.length !== 1 ? 's' : ''}</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table
            columns={columns}
            data={mounts}
            keyField="id"
            emptyText="No mounts available for this server configuration"
            expandable={{
              isExpanded: m => expanded.has(m.id),
              render: m => (
                <div className="grid grid-cols-1 gap-4 text-xs">
                  <div>
                    <div className="text-neutral-500 mb-1">Description</div>
                    <div className="text-neutral-300">
                      {m.description || 'No description provided.'}
                    </div>
                  </div>
                </div>
              )
            }}
            contextMenu={getMountActions}
          />
        </div>
      </div>
    </div>
  );
}
