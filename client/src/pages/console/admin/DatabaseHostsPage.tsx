import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { adminGetDatabaseHosts, adminCreateDatabaseHost, adminDeleteDatabaseHost, AdminDatabaseHost } from '../../../lib/api/admin';
import { notify, Button, Input, Modal, Icons, Table, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../../../components';

export default function DatabaseHostsPage() {
  const navigate = useNavigate();
  const [hosts, setHosts] = useState<AdminDatabaseHost[]>([]);
  const [ready, setReady] = useState(false);
  const [createModal, setCreateModal] = useState(false);
  const [deleteModal, setDeleteModal] = useState<AdminDatabaseHost | null>(null);
  const [form, setForm] = useState({ name: '', host: '', port: '3306', username: '', password: '', max_databases: '0' });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    startLoading();
    adminGetDatabaseHosts().then(res => {
      if (res.success && res.data) setHosts(res.data);
      setReady(true);
      finishLoading();
    });
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    const res = await adminCreateDatabaseHost({
      name: form.name,
      host: form.host,
      port: parseInt(form.port) || 3306,
      username: form.username,
      password: form.password,
      max_databases: parseInt(form.max_databases) || 0,
    });
    setLoading(false);
    if (res.success && res.data) {
      setHosts(prev => [...prev, res.data!]);
      setCreateModal(false);
      setForm({ name: '', host: '', port: '3306', username: '', password: '', max_databases: '0' });
      notify('Success', 'Database host created', 'success');
    } else {
      notify('Error', res.error || 'Failed to create database host', 'error');
    }
  };

  const handleDelete = async () => {
    if (!deleteModal) return;
    setLoading(true);
    const res = await adminDeleteDatabaseHost(deleteModal.id);
    setLoading(false);
    if (res.success) {
      setHosts(prev => prev.filter(h => h.id !== deleteModal.id));
      setDeleteModal(null);
      notify('Success', 'Database host deleted', 'success');
    } else {
      notify('Error', res.error || 'Failed to delete database host', 'error');
    }
  };

  if (!ready) return null;

  const columns = [
    { key: 'name', header: 'Name', render: (h: AdminDatabaseHost) => (
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-violet-500/20">
          <Icons.database className="w-4 h-4 text-violet-400" />
        </div>
        <div>
          <div className="text-sm font-medium text-neutral-100">{h.name}</div>
          <div className="text-xs text-neutral-500">{h.host}:{h.port}</div>
        </div>
      </div>
    )},
    { key: 'username', header: 'Username', render: (h: AdminDatabaseHost) => <code className="text-sm text-neutral-300 font-mono">{h.username}</code> },
    { key: 'databases', header: 'Databases', render: (h: AdminDatabaseHost) => (
      <span className="text-sm text-neutral-300">{h.databases}{h.max_databases > 0 ? ` / ${h.max_databases}` : ''}</span>
    )},
    { key: 'actions', header: '', align: 'right' as const, render: (h: AdminDatabaseHost) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => navigate(`/console/admin/database-hosts/${h.id}`)}>View Databases</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => setDeleteModal(h)} disabled={h.databases > 0} className="text-red-400">Delete</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    )},
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">Admin</span>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Database Hosts</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Database Hosts</h1>
          <p className="text-sm text-neutral-400">Manage MySQL database hosts for server databases.</p>
        </div>
        <Button onClick={() => setCreateModal(true)}><Icons.plus className="w-4 h-4" />Add Host</Button>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{hosts.length} host{hosts.length !== 1 ? 's' : ''}</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table columns={columns} data={hosts} keyField="id" emptyText="No database hosts configured" />
        </div>
      </div>

      <Modal open={createModal} onClose={() => setCreateModal(false)} title="Add Database Host" description="Add a MySQL server to host databases.">
        <form onSubmit={handleCreate} className="space-y-4">
          <Input label="Name" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Main Database Server" required />
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <Input label="Host" value={form.host} onChange={e => setForm(f => ({ ...f, host: e.target.value }))} placeholder="db.example.com" required />
            </div>
            <Input label="Port" type="number" value={form.port} onChange={e => setForm(f => ({ ...f, port: e.target.value }))} placeholder="3306" required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Input label="Username" value={form.username} onChange={e => setForm(f => ({ ...f, username: e.target.value }))} placeholder="root" required />
            <Input label="Password" type="password" value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} required />
          </div>
          <Input label="Max Databases (0 = unlimited)" type="number" value={form.max_databases} onChange={e => setForm(f => ({ ...f, max_databases: e.target.value }))} />
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setCreateModal(false)} disabled={loading}>Cancel</Button>
            <Button type="submit" loading={loading}>Create</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => setDeleteModal(null)} title="Delete Database Host" description="Are you sure you want to delete this database host?">
        <div className="space-y-4">
          {deleteModal && <p className="text-sm text-neutral-300">Host: <code className="text-violet-400">{deleteModal.name}</code></p>}
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={loading}>Cancel</Button>
            <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
