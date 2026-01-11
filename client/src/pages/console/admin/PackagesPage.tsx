import { useEffect, useState, useMemo } from 'react';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { adminGetPackages, adminDeletePackage, Package } from '../../../lib/api';
import { notify, Button, Input, Modal, CreatePackageModal, Icons, Pagination } from '../../../components';

function PackageCard({ pkg, onDelete, onEdit }: { pkg: Package; onDelete: () => void; onEdit: () => void }) {
  const portsCount = pkg.ports?.length || 0;
  const varsCount = pkg.variables?.length || 0;
  return (
    <div className="rounded-xl bg-neutral-800/30 overflow-hidden">
      <div className="p-5">
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-3">
            {pkg.icon ? (
              <img src={pkg.icon} alt="" className="w-10 h-10 rounded-lg object-cover" />
            ) : (
              <div className="w-10 h-10 rounded-lg bg-amber-500/20 flex items-center justify-center">
                <Icons.cube className="w-5 h-5 text-amber-400" />
              </div>
            )}
            <div>
              <h3 className="text-sm font-semibold text-neutral-100">{pkg.name}</h3>
              <p className="text-xs text-neutral-400">{pkg.version || 'No version'} • {pkg.author || 'Unknown author'}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button onClick={onEdit} className="p-1.5 text-neutral-400 hover:text-neutral-100 hover:bg-neutral-700 rounded-lg transition-colors" title="Edit Package">
              <Icons.edit className="w-4 h-4" />
            </button>
            <button onClick={onDelete} className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors" title="Delete Package">
              <Icons.trash className="w-4 h-4" />
            </button>
          </div>
        </div>
        {pkg.description && <p className="text-xs text-neutral-400 mb-3 line-clamp-2">{pkg.description}</p>}
        <div className="flex items-center gap-4 text-xs text-neutral-400">
          <div className="flex items-center gap-1.5">
            <Icons.cubeOutline className="w-3.5 h-3.5" />
            <span className="font-mono">{pkg.docker_image}</span>
          </div>
        </div>
        <div className="flex items-center gap-3 mt-3">
          <span className="inline-flex items-center rounded-md bg-neutral-700/50 px-2 py-1 text-xs text-neutral-300">{portsCount} port{portsCount !== 1 ? 's' : ''}</span>
          <span className="inline-flex items-center rounded-md bg-neutral-700/50 px-2 py-1 text-xs text-neutral-300">{varsCount} variable{varsCount !== 1 ? 's' : ''}</span>
        </div>
      </div>
    </div>
  );
}

export default function PackagesPage() {
  const [packages, setPackages] = useState<Package[]>([]);
  const [ui, setUi] = useState({ ready: false, refreshing: false, search: '', page: 1, perPage: 20 });
  const [showCreate, setShowCreate] = useState(false);
  const [editPackage, setEditPackage] = useState<Package | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ pkg: Package; loading: boolean } | null>(null);

  const filtered = useMemo(() => {
    if (!ui.search) return packages;
    const q = ui.search.toLowerCase();
    return packages.filter(p => p.name.toLowerCase().includes(q) || p.docker_image.toLowerCase().includes(q) || p.author?.toLowerCase().includes(q) || p.id.toLowerCase().includes(q));
  }, [packages, ui.search]);

  const paginated = useMemo(() => {
    const start = (ui.page - 1) * ui.perPage;
    return filtered.slice(start, start + ui.perPage);
  }, [filtered, ui.page, ui.perPage]);

  const totalPages = Math.ceil(filtered.length / ui.perPage) || 1;

  useEffect(() => { startLoading(); loadPackages(true); }, []);

  const loadPackages = async (initial = false) => {
    if (!initial) setUi(s => ({ ...s, refreshing: true }));
    const res = await adminGetPackages();
    if (res.success && res.data) setPackages(res.data);
    setUi(s => ({ ...s, refreshing: false, ready: initial ? true : s.ready }));
    if (initial) finishLoading();
  };

  const handleDelete = async () => {
    if (!confirmDelete) return;
    setConfirmDelete(s => s && { ...s, loading: true });
    const res = await adminDeletePackage(confirmDelete.pkg.id);
    if (res.success) { notify('Success', 'Package deleted', 'success'); setConfirmDelete(null); loadPackages(); }
    else { notify('Error', res.error || 'Failed to delete package', 'error'); setConfirmDelete(s => s && { ...s, loading: false }); }
  };

  if (!ui.ready) return null;

  return (
    <>
      <CreatePackageModal open={showCreate || !!editPackage} onClose={() => { setShowCreate(false); setEditPackage(null); }} onCreated={() => loadPackages()} editPackage={editPackage} />

      <Modal open={!!confirmDelete} onClose={() => !confirmDelete?.loading && setConfirmDelete(null)} title="Delete Package" description={`Are you sure you want to delete "${confirmDelete?.pkg.name}"? This cannot be undone.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setConfirmDelete(null)} disabled={confirmDelete?.loading}>Cancel</Button>
          <Button variant="danger" onClick={handleDelete} loading={confirmDelete?.loading}>Delete</Button>
        </div>
      </Modal>

      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">Packages</h1>
            <p className="text-sm text-neutral-400">Manage server packages and configurations.</p>
          </div>
          <Button onClick={() => setShowCreate(true)} className="w-full sm:w-auto"><Icons.plus className="w-4 h-4" />Add Package</Button>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex items-center gap-3 flex-1">
            <form onSubmit={e => e.preventDefault()} className="flex-1 sm:max-w-sm">
              <Input placeholder="Search by name, image, or author..." value={ui.search} onChange={e => setUi(s => ({ ...s, search: e.target.value, page: 1 }))} />
            </form>
            <button onClick={() => loadPackages()} disabled={ui.refreshing} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 transition-colors disabled:opacity-50 flex-shrink-0" title="Refresh">
              <Icons.refresh className={`w-4 h-4 ${ui.refreshing ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <Pagination page={ui.page} totalPages={totalPages} total={filtered.length} perPage={ui.perPage} onPageChange={p => setUi(s => ({ ...s, page: p }))} onPerPageChange={pp => setUi(s => ({ ...s, perPage: pp, page: 1 }))} loading={ui.refreshing} />
        </div>

        {filtered.length === 0 ? (
          <div className="rounded-xl bg-neutral-800/30 p-12 text-center">
            <div className="w-12 h-12 rounded-full bg-neutral-700/50 flex items-center justify-center mx-auto mb-4">
              <Icons.cube className="w-6 h-6 text-neutral-400" />
            </div>
            <h3 className="text-sm font-medium text-neutral-100 mb-1">{packages.length === 0 ? 'No packages yet' : 'No packages found'}</h3>
            <p className="text-xs text-neutral-400 mb-4">{packages.length === 0 ? 'Create your first package to get started.' : 'Try adjusting your search.'}</p>
            {packages.length === 0 && <Button onClick={() => setShowCreate(true)}><Icons.plus className="w-4 h-4" />Add Package</Button>}
          </div>
        ) : (
          <div className="rounded-xl bg-neutral-800/30">
            <div className="px-4 py-2 text-xs text-neutral-400">{filtered.length} package{filtered.length !== 1 ? 's' : ''}</div>
            <div className="bg-neutral-900/40 rounded-lg p-1">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
                {paginated.map(pkg => <PackageCard key={pkg.id} pkg={pkg} onDelete={() => setConfirmDelete({ pkg, loading: false })} onEdit={() => setEditPackage(pkg)} />)}
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
