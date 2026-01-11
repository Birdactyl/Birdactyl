import { useState, useRef, useEffect } from 'react';
import { adminGetIPBans, adminCreateIPBan, adminDeleteIPBan, type IPBan } from '../../../lib/api';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { notify, Button, Input, Modal, Pagination, Icons, Table } from '../../../components';

export default function IPBansPage() {
  const [bans, setBans] = useState<IPBan[]>([]);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [createModal, setCreateModal] = useState({ open: false, loading: false, ip: '', reason: '' });
  const [deleteModal, setDeleteModal] = useState<{ ban: IPBan; loading: boolean } | null>(null);
  const requestId = useRef(0);

  const load = async (p: number, pp: number, s: string, initial = false) => {
    const currentRequest = ++requestId.current;
    setLoading(true);
    const res = await adminGetIPBans(p, pp, s);
    if (currentRequest !== requestId.current) return;
    if (res.success && res.data) {
      setBans(res.data.bans || []);
      setPage(res.data.page);
      setTotalPages(res.data.total_pages);
      setTotal(res.data.total);
    } else {
      notify('Error', res.error || 'Failed to load IP bans', 'error');
    }
    setLoading(false);
    if (initial) { setReady(true); finishLoading(); }
  };

  useEffect(() => { startLoading(); load(1, perPage, '', true); }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
    load(1, perPage, searchInput);
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateModal(m => ({ ...m, loading: true }));
    const res = await adminCreateIPBan(createModal.ip, createModal.reason);
    if (res.success) {
      notify('Success', 'IP banned', 'success');
      setCreateModal({ open: false, loading: false, ip: '', reason: '' });
      load(page, perPage, search);
    } else {
      notify('Error', res.error || 'Failed to ban IP', 'error');
      setCreateModal(m => ({ ...m, loading: false }));
    }
  };

  const handleDelete = async () => {
    if (!deleteModal) return;
    setDeleteModal(m => m && { ...m, loading: true });
    const res = await adminDeleteIPBan(deleteModal.ban.id);
    if (res.success) {
      notify('Success', 'IP unbanned', 'success');
      setDeleteModal(null);
      load(page, perPage, search);
    } else {
      notify('Error', res.error || 'Failed to unban IP', 'error');
      setDeleteModal(m => m && { ...m, loading: false });
    }
  };

  if (!ready) return null;

  const columns = [
    { key: 'ip', header: 'IP Address', render: (ban: IPBan) => <span className="text-sm font-mono text-neutral-100">{ban.ip}</span> },
    { key: 'reason', header: 'Reason', render: (ban: IPBan) => <span className="text-sm text-neutral-400">{ban.reason || '—'}</span> },
    { key: 'created', header: 'Banned', render: (ban: IPBan) => <span className="text-sm text-neutral-400">{new Date(ban.created_at).toLocaleString()}</span> },
    { key: 'actions', header: '', align: 'right' as const, render: (ban: IPBan) => (
      <button onClick={() => setDeleteModal({ ban, loading: false })} className="text-xs text-red-400 hover:text-red-300 transition-colors">Unban</button>
    )},
  ];

  return (
    <>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">IP Bans</h1>
            <p className="text-sm text-neutral-400">Block IP addresses from accessing the panel.</p>
          </div>
          <Button onClick={() => setCreateModal(m => ({ ...m, open: true }))} className="w-full sm:w-auto"><Icons.plus className="w-4 h-4" />Ban IP</Button>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <form onSubmit={handleSearch} className="flex-1 sm:max-w-sm">
            <Input placeholder="Search by IP or reason..." value={searchInput} onChange={e => setSearchInput(e.target.value)} />
          </form>
          <Pagination page={page} totalPages={totalPages} total={total} perPage={perPage} onPageChange={p => load(p, perPage, search)} onPerPageChange={pp => { setPerPage(pp); load(1, pp, search); }} loading={loading} />
        </div>

        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-4 py-2 text-xs text-neutral-400">{total} banned IP{total !== 1 ? 's' : ''}</div>
          <div className="bg-neutral-900/40 rounded-lg p-1">
            <Table columns={columns} data={bans} keyField="id" loading={loading} emptyText="No IP bans" />
          </div>
        </div>
      </div>

      <Modal open={createModal.open} onClose={() => !createModal.loading && setCreateModal(m => ({ ...m, open: false }))} title="Ban IP Address" description="This IP will be blocked from logging in or registering.">
        <form onSubmit={handleCreate} className="space-y-4">
          <Input label="IP Address" placeholder="192.168.1.1" value={createModal.ip} onChange={e => setCreateModal(m => ({ ...m, ip: e.target.value }))} required />
          <Input label="Reason (optional)" placeholder="Abuse, spam, etc." value={createModal.reason} onChange={e => setCreateModal(m => ({ ...m, reason: e.target.value }))} />
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setCreateModal(m => ({ ...m, open: false }))} disabled={createModal.loading}>Cancel</Button>
            <Button type="submit" loading={createModal.loading}>Ban IP</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Unban IP" description={`Remove the ban on ${deleteModal?.ban.ip}?`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button onClick={handleDelete} loading={deleteModal?.loading}>Unban</Button>
        </div>
      </Modal>
    </>
  );
}
