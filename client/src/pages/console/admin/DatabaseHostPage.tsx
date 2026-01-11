import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { adminGetHostDatabases, adminDeleteHostDatabase, HostDatabase } from '../../../lib/api/admin';
import { notify, Button, Modal, Icons, Table, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../../../components';

export default function DatabaseHostPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [databases, setDatabases] = useState<HostDatabase[]>([]);
  const [ready, setReady] = useState(false);
  const [deleteModal, setDeleteModal] = useState<HostDatabase | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!id) return;
    startLoading();
    adminGetHostDatabases(id).then(res => {
      if (res.success && res.data) setDatabases(res.data);
      setReady(true);
      finishLoading();
    });
  }, [id]);

  const handleDelete = async () => {
    if (!id || !deleteModal) return;
    setLoading(true);
    const res = await adminDeleteHostDatabase(id, deleteModal.id);
    setLoading(false);
    if (res.success) {
      setDatabases(prev => prev.filter(d => d.id !== deleteModal.id));
      setDeleteModal(null);
      notify('Success', 'Database deleted', 'success');
    } else {
      notify('Error', res.error || 'Failed to delete database', 'error');
    }
  };

  if (!ready) return null;

  const columns = [
    { key: 'name', header: 'Database', render: (db: HostDatabase) => (
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-violet-500/20">
          <Icons.database className="w-4 h-4 text-violet-400" />
        </div>
        <div>
          <div className="text-sm font-medium text-neutral-100">{db.database_name}</div>
          <div className="text-xs text-neutral-500">{db.username}</div>
        </div>
      </div>
    )},
    { key: 'server', header: 'Server', render: (db: HostDatabase) => (
      <span className="text-sm text-neutral-300">{db.server_name || db.server_id}</span>
    )},
    { key: 'created', header: 'Created', render: (db: HostDatabase) => (
      <span className="text-sm text-neutral-400">{new Date(db.created_at).toLocaleDateString()}</span>
    )},
    { key: 'actions', header: '', align: 'right' as const, render: (db: HostDatabase) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setDeleteModal(db)} className="text-red-400">Delete</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )},
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <button onClick={() => navigate('/console/admin/database-hosts')} className="font-medium text-neutral-200 hover:text-neutral-100">Database Hosts</button>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Databases</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Databases</h1>
          <p className="text-sm text-neutral-400">{databases.length} database{databases.length !== 1 ? 's' : ''} on this host.</p>
        </div>
        <Button variant="secondary" onClick={() => navigate('/console/admin/database-hosts')}><Icons.chevronLeft className="w-4 h-4" />Back</Button>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table columns={columns} data={databases} keyField="id" emptyText="No databases on this host" />
        </div>
      </div>

      <Modal open={!!deleteModal} onClose={() => setDeleteModal(null)} title="Delete Database" description="Are you sure? This will permanently delete the database and its data.">
        <div className="space-y-4">
          {deleteModal && (
            <div className="text-sm text-neutral-300">
              <p>Database: <code className="text-violet-400">{deleteModal.database_name}</code></p>
              <p>Server: <code className="text-neutral-400">{deleteModal.server_name}</code></p>
            </div>
          )}
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={loading}>Cancel</Button>
            <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
