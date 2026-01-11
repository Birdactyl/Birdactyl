import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Server, getServer, addAllocation, setPrimaryAllocation, deleteAllocation } from '../../../lib/api';
import { useAsyncCallback } from '../../../hooks/useAsync';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import Button from '../../../components/ui/Button';
import { Icons } from '../../../components/Icons';
import { PermissionDenied, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, Table } from '../../../components';

interface Allocation { port: number; primary?: boolean; }

const parsePorts = (data: any) => typeof data.ports === 'string' ? JSON.parse(data.ports) : data.ports || [];

export default function NetworkPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [allocations, setAllocations] = useState<Allocation[]>([]);
  const { can, loading: permsLoading } = useServerPermissions(id);

  const fetchServer = async () => {
    if (!id) return;
    const res = await getServer(id);
    if (res.success && res.data) {
      setServer(res.data);
      setAllocations(parsePorts(res.data));
    }
  };

  useEffect(() => { fetchServer(); }, [id]);

  const [handleAddAllocation, addLoading] = useAsyncCallback(async () => {
    if (!id) return;
    const res = await addAllocation(id);
    if (res.success && res.data) {
      setAllocations(Array.isArray(res.data) ? res.data : parsePorts(res.data));
    }
  }, [id]);

  const [handleSetPrimary, primaryLoading] = useAsyncCallback(async (port: number) => {
    if (!id) return;
    const res = await setPrimaryAllocation(id, port);
    if (res.success && res.data) {
      setAllocations(Array.isArray(res.data) ? res.data : parsePorts(res.data));
    }
  }, [id]);

  const [handleDelete, deleteLoading] = useAsyncCallback(async (port: number) => {
    if (!id) return;
    const res = await deleteAllocation(id, port);
    if (res.success && res.data) {
      setAllocations(Array.isArray(res.data) ? res.data : parsePorts(res.data));
    }
  }, [id]);

  const loading = addLoading || primaryLoading || deleteLoading;

  if (permsLoading) return null;
  if (!can('allocation.view')) return <PermissionDenied message="You don't have permission to view network settings" />;

  if (!server) return <div className="text-neutral-400">Loading...</div>;

  const host = server.node?.display_ip || server.node?.fqdn || 'unknown';

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server.name}</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">Network</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Network</h1>
          <p className="text-sm text-neutral-400">Manage IP and port allocations for this server.</p>
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto">
          {can('allocation.add') && <Button className="w-full sm:w-auto" onClick={handleAddAllocation} disabled={loading}>New allocation</Button>}
        </div>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{allocations.length} allocation{allocations.length !== 1 ? 's' : ''}</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table
            columns={[
              { key: 'address', header: 'Address', render: (alloc: Allocation) => (
                <div className="text-sm font-medium text-neutral-100 flex items-center gap-2">
                  <span>{host}:{alloc.port}</span>
                  {alloc.primary && (
                    <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[11px] font-semibold text-emerald-300">
                      PRIMARY
                    </span>
                  )}
                </div>
              )},
              { key: 'host', header: 'Host', render: () => <span className="text-sm text-neutral-300">{host}</span> },
              { key: 'port', header: 'Port', render: (alloc: Allocation) => <span className="text-sm text-neutral-300 tabular-nums">{alloc.port}</span> },
              { key: 'actions', header: '', align: 'right' as const, render: (alloc: Allocation) => (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" disabled={loading}><Icons.ellipsis className="w-5 h-5" /></Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {can('allocation.set_primary') && !alloc.primary && (
                      <DropdownMenuItem onSelect={() => handleSetPrimary(alloc.port)}>Set primary</DropdownMenuItem>
                    )}
                    {can('allocation.delete') && !alloc.primary && (
                      <DropdownMenuItem onSelect={() => handleDelete(alloc.port)} className="text-red-400">Delete</DropdownMenuItem>
                    )}
                    {alloc.primary && (
                      <DropdownMenuItem disabled>Primary allocation</DropdownMenuItem>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
              )},
            ]}
            data={allocations}
            keyField="port"
            emptyText="No allocations found"
          />
        </div>
      </div>
    </div>
  );
}
