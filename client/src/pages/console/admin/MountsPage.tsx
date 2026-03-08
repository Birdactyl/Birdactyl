import { useState, useRef, useEffect } from 'react';
import { adminGetMounts, adminCreateMount, adminUpdateMount, adminDeleteMount, adminGetNodes, adminGetPackages, type Mount, type Node, type Package } from '../../../lib/api';
import { startLoading, finishLoading } from '../../../lib/pageLoader';
import { notify, Button, Input, Modal, Pagination, Icons, Table, SlidePanel, Checkbox, ContextMenu } from '../../../components';

export default function MountsPage() {
  const [mounts, setMounts] = useState<Mount[]>([]);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [ready, setReady] = useState(false);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');

  const [availableNodes, setAvailableNodes] = useState<Node[]>([]);
  const [availablePackages, setAvailablePackages] = useState<Package[]>([]);

  const [slideOpen, setSlideOpen] = useState(false);
  const [editingMount, setEditingMount] = useState<Mount | null>(null);
  const [slideLoading, setSlideLoading] = useState(false);

  const [formName, setFormName] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formSource, setFormSource] = useState('');
  const [formTarget, setFormTarget] = useState('');
  const [formReadOnly, setFormReadOnly] = useState(false);
  const [formUserMountable, setFormUserMountable] = useState(false);
  const [formNavigable, setFormNavigable] = useState(false);
  const [formNodes, setFormNodes] = useState<string[]>([]);
  const [formPackages, setFormPackages] = useState<string[]>([]);

  const [deleteModal, setDeleteModal] = useState<{ mount: Mount; loading: boolean } | null>(null);
  const requestId = useRef(0);

  const load = async (p: number, pp: number, s: string, initial = false) => {
    const currentRequest = ++requestId.current;
    setLoading(true);
    const res = await adminGetMounts(p, pp, s);
    if (currentRequest !== requestId.current) return;
    if (res.success && res.data) {
      setMounts(res.data.mounts || []);
      setPage(res.data.page);
      setTotalPages(res.data.total_pages);
      setTotal(res.data.total);
    } else {
      notify('Error', res.error || 'Failed to load mounts', 'error');
    }
    
    if (initial) {
      const [nodesRes, packagesRes] = await Promise.all([
        adminGetNodes(),
        adminGetPackages()
      ]);
      if (nodesRes.success && nodesRes.data) setAvailableNodes(nodesRes.data);
      if (packagesRes.success && packagesRes.data) setAvailablePackages(packagesRes.data || []);
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

  const openSlide = (mount?: Mount) => {
    if (mount) {
      setEditingMount(mount);
      setFormName(mount.name);
      setFormDescription(mount.description || '');
      setFormSource(mount.source);
      setFormTarget(mount.target);
      setFormReadOnly(mount.read_only);
      setFormUserMountable(mount.user_mountable);
      setFormNavigable(mount.navigable);
      setFormNodes(mount.nodes?.map(n => n.id) || []);
      setFormPackages(mount.packages?.map(p => p.id) || []);
    } else {
      setEditingMount(null);
      setFormName('');
      setFormDescription('');
      setFormSource('');
      setFormTarget('');
      setFormReadOnly(false);
      setFormUserMountable(false);
      setFormNavigable(false);
      setFormNodes([]);
      setFormPackages([]);
    }
    setSlideOpen(true);
  };

  const handleSave = async () => {
    if (!formName || !formSource || !formTarget) {
      notify('Error', 'Name, Source, and Target are required.', 'error');
      return;
    }

    setSlideLoading(true);
    let res;
    if (editingMount) {
      res = await adminUpdateMount(editingMount.id, {
        name: formName,
        description: formDescription,
        source: formSource,
        target: formTarget,
        read_only: formReadOnly,
        user_mountable: formUserMountable,
        navigable: formNavigable,
        nodes: formNodes,
        packages: formPackages
      });
    } else {
      res = await adminCreateMount(formName, formDescription, formSource, formTarget, formReadOnly, formUserMountable, formNavigable, formNodes, formPackages);
    }

    setSlideLoading(false);
    if (res.success) {
      notify('Success', editingMount ? 'Mount updated' : 'Mount created', 'success');
      setSlideOpen(false);
      load(page, perPage, search);
    } else {
      notify('Error', res.error || 'Failed to save mount', 'error');
    }
  };

  const handleDelete = async () => {
    if (!deleteModal) return;
    setDeleteModal(m => m && { ...m, loading: true });
    const res = await adminDeleteMount(deleteModal.mount.id);
    if (res.success) {
      notify('Success', 'Mount deleted', 'success');
      setDeleteModal(null);
      load(page, perPage, search);
    } else {
      notify('Error', res.error || 'Failed to delete mount', 'error');
      setDeleteModal(m => m && { ...m, loading: false });
    }
  };

  if (!ready) return null;

  const getMountActions = (m: Mount) => [
    { label: 'Edit', onClick: () => openSlide(m) },
    { label: 'Delete', onClick: () => setDeleteModal({ mount: m, loading: false }), variant: 'danger' as const },
  ];

  const columns = [
    { 
      key: 'name', header: 'Mount', render: (m: Mount) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-indigo-500/20">
            <Icons.folder className="w-4 h-4 text-indigo-400" />
          </div>
          <div>
            <div className="text-sm font-medium text-neutral-100">{m.name}</div>
            <div className="text-xs text-neutral-500">{m.id.substring(0, 8)}</div>
          </div>
        </div>
      ) 
    },
    { key: 'description', header: 'Description', render: (m: Mount) => <span className="text-sm text-neutral-400">{m.description || '\u2014'}</span> },
    { key: 'source', header: 'Source Path', render: (m: Mount) => <span className="text-sm text-neutral-400">{m.source}</span> },
    { key: 'target', header: 'Target Path', render: (m: Mount) => <span className="text-sm text-neutral-400">{m.target}</span> },
    { key: 'readOnly', header: 'Access', render: (m: Mount) => <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${m.read_only ? 'bg-amber-500/10 text-amber-400 ring-amber-500/20' : 'bg-emerald-500/10 text-emerald-400 ring-emerald-500/20'}`}>{m.read_only ? 'Read Only' : 'Read/Write'}</span> },
    { key: 'userMountable', header: 'User Mountable', render: (m: Mount) => <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${m.user_mountable ? 'bg-emerald-500/10 text-emerald-400 ring-emerald-500/20' : 'bg-neutral-500/10 text-neutral-400 ring-neutral-500/20'}`}>{m.user_mountable ? 'Yes' : 'No'}</span> },
    { key: 'navigable', header: 'Navigable', render: (m: Mount) => <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${m.navigable ? 'bg-emerald-500/10 text-emerald-400 ring-emerald-500/20' : 'bg-neutral-500/10 text-neutral-400 ring-neutral-500/20'}`}>{m.navigable ? 'Yes' : 'No'}</span> },
    {
      key: 'actions', header: '', align: 'right' as const, render: (m: Mount) => (
        <ContextMenu
          align="end"
          trigger={<Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>}
          items={getMountActions(m)}
        />
      )
    },
  ];

  return (
    <>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">Mounts</h1>
            <p className="text-sm text-neutral-400">Configure global directories that can be mounted to servers.</p>
          </div>
          <Button onClick={() => openSlide()} className="w-full sm:w-auto"><Icons.plus className="w-4 h-4 mr-2" />New Mount</Button>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <form onSubmit={handleSearch} className="flex-1 sm:max-w-sm">
            <Input placeholder="Search by name or path..." value={searchInput} onChange={e => setSearchInput(e.target.value)} />
          </form>
          <Pagination page={page} totalPages={totalPages} total={total} perPage={perPage} onPageChange={p => load(p, perPage, search)} onPerPageChange={pp => { setPerPage(pp); load(1, pp, search); }} loading={loading} />
        </div>

        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-4 py-2 text-xs text-neutral-400">{total} mount{total !== 1 ? 's' : ''}</div>
          <div className="bg-neutral-900/40 rounded-lg p-1">
            <Table columns={columns} data={mounts} keyField="id" loading={loading} emptyText="No mounts configured" contextMenu={getMountActions} />
          </div>
        </div>
      </div>

      <SlidePanel
        open={slideOpen}
        onClose={() => !slideLoading && setSlideOpen(false)}
        title={editingMount ? "Edit Mount" : "Create Mount"}
        description={editingMount ? "Modify the properties of this mount." : "Create a new host-to-container mount mapping."}
        footer={
          <div className="flex justify-end gap-3 w-full">
            <Button variant="ghost" onClick={() => setSlideOpen(false)} disabled={slideLoading}>Cancel</Button>
            <Button onClick={handleSave} loading={slideLoading}>{editingMount ? 'Save Changes' : 'Create Mount'}</Button>
          </div>
        }
      >
        <div className="space-y-6 py-2">
          <div className="space-y-4">
            <h3 className="text-sm font-medium text-neutral-200">Basic Details</h3>
            <Input 
              label="Name" 
              placeholder="Shared Textures" 
              value={formName} 
              onChange={e => setFormName(e.target.value)} 
              required 
            />
            <Input 
              label="Description (optional)" 
              placeholder="Global textures directory" 
              value={formDescription} 
              onChange={e => setFormDescription(e.target.value)} 
            />
          </div>

          <div className="space-y-4 pt-4 border-t border-neutral-800">
            <h3 className="text-sm font-medium text-neutral-200">Path configuration</h3>
            
            <Input 
              label="Source Path" 
              placeholder="/var/lib/birdactyl/mounts/textures" 
              value={formSource} 
              onChange={e => setFormSource(e.target.value)} 
              required 
            />
            
            <Input 
              label="Target Path" 
              placeholder="/home/container/textures" 
              value={formTarget} 
              onChange={e => setFormTarget(e.target.value)} 
              required 
            />
            
            <div>
              <Checkbox 
                label="Read Only Access" 
                checked={formReadOnly} 
                onChange={(c) => setFormReadOnly(!!c)} 
              />
              <p className="text-xs text-neutral-400 mt-1 ml-6">If checked, the server cannot modify files inside this mount.</p>
            </div>

            <div>
              <Checkbox 
                label="User Mountable" 
                checked={formUserMountable} 
                onChange={(c) => setFormUserMountable(!!c)} 
              />
              <p className="text-xs text-neutral-400 mt-1 ml-6">If checked, owners will be able to opt-in and manually attach this mount to their servers themselves.</p>
            </div>

            <div>
              <Checkbox 
                label="Navigable" 
                checked={formNavigable} 
                onChange={(c) => setFormNavigable(!!c)} 
              />
              <p className="text-xs text-neutral-400 mt-1 ml-6">If checked, the mount will be visible and accessible within the server's File Manager and via SFTP.</p>
            </div>
          </div>

          <div className="space-y-4 pt-4 border-t border-neutral-800">
            <h3 className="text-sm font-medium text-neutral-200">Nodes</h3>
            <p className="text-xs text-neutral-400 mb-2">Select which nodes this mount exists on. If a node is not selected, servers on that node will not be able to use it.</p>
            <ContextMenu 
              className="w-full"
              align="start"
              trigger={
                <Button variant="ghost" className="w-full justify-between bg-neutral-900/50 border border-neutral-800 hover:bg-neutral-800/80">
                  <span>{formNodes.length === 0 ? 'Select Nodes' : `${formNodes.length} Node${formNodes.length > 1 ? 's' : ''} Selected`}</span>
                  <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
                </Button>
              }
              items={availableNodes.length > 0 ? availableNodes.map((node: Node) => ({
                label: (
                  <div 
                    className="flex items-center gap-3 w-64 px-1 py-0.5 cursor-pointer flex-1"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      setFormNodes((prev: string[]) => prev.includes(node.id) ? prev.filter(id => id !== node.id) : [...prev, node.id]);
                    }}
                  >
                    <Checkbox 
                      checked={formNodes.includes(node.id)} 
                      onChange={() => {}}
                    />
                    <span className="flex-1 truncate">{node.name}</span>
                  </div>
                )
              })) : [{ label: 'No nodes available', disabled: true }]}
            />
          </div>

          <div className="space-y-4 pt-4 border-t border-neutral-800">
            <h3 className="text-sm font-medium text-neutral-200">Packages</h3>
            <p className="text-xs text-neutral-400 mb-2">Restrict this mount so it is only available to servers using the specified packages. Select none to allow all packages.</p>
            <ContextMenu 
              className="w-full"
              align="start"
              trigger={
                <Button variant="ghost" className="w-full justify-between bg-neutral-900/50 border border-neutral-800 hover:bg-neutral-800/80">
                  <span>{formPackages.length === 0 ? 'Select Packages' : `${formPackages.length} Package${formPackages.length > 1 ? 's' : ''} Selected`}</span>
                  <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
                </Button>
              }
              items={availablePackages.length > 0 ? availablePackages.map((pkg: Package) => ({
                label: (
                  <div 
                    className="flex items-center gap-3 w-64 px-1 py-0.5 cursor-pointer flex-1"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      setFormPackages((prev: string[]) => prev.includes(pkg.id) ? prev.filter(id => id !== pkg.id) : [...prev, pkg.id]);
                    }}
                  >
                    <Checkbox 
                      checked={formPackages.includes(pkg.id)} 
                      onChange={() => {}}
                    />
                    <span className="flex-1 truncate">{pkg.name}</span>
                  </div>
                )
              })) : [{ label: 'No packages available', disabled: true }]}
            />
          </div>
        </div>
      </SlidePanel>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Delete Mount" description={`Are you sure you want to delete the mount "${deleteModal?.mount.name}"? Servers relying on this mount will lose access to the underlying directory when they restart.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button onClick={handleDelete} loading={deleteModal?.loading} variant="danger">Delete Mount</Button>
        </div>
      </Modal>
    </>
  );
}
