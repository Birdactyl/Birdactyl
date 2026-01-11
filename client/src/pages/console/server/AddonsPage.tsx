import { useState, useEffect, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, Server, getAddonSources, searchAddons, getAddonVersions, listInstalledAddons, installAddon, installModpack, AddonSource, Addon, AddonVersion, InstalledAddon } from '../../../lib/api';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Icons, Modal, Input, PermissionDenied, Table, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../../../components';
import { DeleteAddonModal } from '../../../components/modals';
import { notify } from '../../../components/feedback/Notification';

export default function AddonsPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [sources, setSources] = useState<AddonSource[]>([]);
  const [activeSource, setActiveSource] = useState<AddonSource | null>(null);
  const [tab, setTab] = useState<'browse' | 'installed'>('browse');
  const [search, setSearch] = useState('');
  const [addons, setAddons] = useState<Addon[]>([]);
  const [installed, setInstalled] = useState<InstalledAddon[]>([]);
  const [loading, setLoading] = useState({ sources: true, search: false, installed: false });
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [versionModal, setVersionModal] = useState<{ addon: Addon; versions: AddonVersion[]; loading: boolean; installing: string | null } | null>(null);
  const [deleteFile, setDeleteFile] = useState<InstalledAddon | null>(null);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [hasMore, setHasMore] = useState(false);
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => res.success && res.data && setServer(res.data));
    loadSources();
  }, [id]);

  useEffect(() => {
    if (!id || !activeSource) return;
    setLoading(s => ({ ...s, installed: true }));
    listInstalledAddons(id, activeSource.id).then(res => {
      if (res.success && res.data) setInstalled(res.data.filter(f => !f.is_dir));
      setLoading(s => ({ ...s, installed: false }));
    });
  }, [id, activeSource]);

  const loadSources = async () => {
    if (!id) return;
    const res = await getAddonSources(id);
    if (res.success && res.data) {
      setSources(res.data);
      if (res.data.length > 0) setActiveSource(res.data[0]);
    }
    setLoading(s => ({ ...s, sources: false }));
  };

  const handleSearch = async (e?: React.FormEvent, newPage = 1, newPerPage = perPage) => {
    e?.preventDefault();
    if (!id || !activeSource || !search.trim()) return;
    setLoading(s => ({ ...s, search: true }));
    setExpanded(new Set());
    const offset = (newPage - 1) * newPerPage;
    const res = await searchAddons(id, activeSource.id, search.trim(), newPerPage, offset);
    if (res.success && res.data) {
      setAddons(res.data);
      setHasMore(res.data.length === newPerPage);
      setPage(newPage);
    }
    setLoading(s => ({ ...s, search: false }));
  };

  const handlePageChange = (newPage: number) => handleSearch(undefined, newPage, perPage);
  const handlePerPageChange = (newPerPage: number) => { setPerPage(newPerPage); handleSearch(undefined, 1, newPerPage); };

  const loadInstalled = async () => {
    if (!id || !activeSource) return;
    setLoading(s => ({ ...s, installed: true }));
    const res = await listInstalledAddons(id, activeSource.id);
    if (res.success && res.data) setInstalled(res.data.filter(f => !f.is_dir));
    setLoading(s => ({ ...s, installed: false }));
  };

  const openVersions = async (addon: Addon) => {
    if (!id || !activeSource) return;
    setVersionModal({ addon, versions: [], loading: true, installing: null });
    const res = await getAddonVersions(id, activeSource.id, addon.id);
    if (res.success && res.data) {
      setVersionModal(s => s && { ...s, versions: res.data!, loading: false });
    } else {
      setVersionModal(null);
      notify('Error', 'Failed to load versions', 'error');
    }
  };

  const handleInstall = async (version: AddonVersion) => {
    if (!id || !activeSource || !versionModal) return;
    setVersionModal(s => s && { ...s, installing: version.id });
    
    if (activeSource.type === 'modpack') {
      const res = await installModpack(id, version.download_url || '', activeSource.id);
      if (res.success && res.data) {
        const result = res.data;
        if (result.files_failed > 0) {
          notify('Installed', `${result.files_installed} files installed, ${result.files_failed} failed`, 'info');
        } else {
          notify('Installed', `${result.name} installed (${result.files_installed} files)`, 'success');
        }
        setVersionModal(null);
      } else {
        notify('Error', res.error || 'Failed to install modpack', 'error');
        setVersionModal(s => s && { ...s, installing: null });
      }
    } else {
      const res = await installAddon(
        id,
        activeSource.id,
        version.download_url,
        version.file_name || `${versionModal.addon.name}.jar`,
        version.mod_id?.toString(),
        version.id?.toString()
      );
      if (res.success) {
        notify('Installed', `${versionModal.addon.name} has been installed`, 'success');
        setVersionModal(null);
        loadInstalled();
      } else {
        notify('Error', res.error || 'Failed to install addon', 'error');
        setVersionModal(s => s && { ...s, installing: null });
      }
    }
  };

  const toggleExpand = (addonId: string) => {
    const next = new Set(expanded);
    next.has(addonId) ? next.delete(addonId) : next.add(addonId);
    setExpanded(next);
  };

  const filteredInstalled = useMemo(() => {
    if (!search.trim() || tab !== 'installed') return installed;
    return installed.filter(f => f.name.toLowerCase().includes(search.toLowerCase()));
  }, [installed, search, tab]);

  const browseColumns = useMemo(() => [
    {
      key: 'expand',
      header: '',
      className: 'w-8',
      render: (addon: Addon) => (
        <button onClick={() => toggleExpand(addon.id)} className="text-neutral-400 hover:text-neutral-200 transition">
          <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(addon.id) ? 'rotate-90' : ''}`} />
        </button>
      ),
    },
    {
      key: 'addon',
      header: 'Addon',
      render: (addon: Addon) => {
        const addonNameLower = addon.name.toLowerCase().replace(/\s+/g, '');
        const isAddonInstalled = installed.some(f => {
          const fileName = f.name.toLowerCase().replace(/\.jar$/, '');
          return fileName === addonNameLower || fileName.startsWith(addonNameLower + '-') || fileName.startsWith(addonNameLower + '_') || fileName.includes(addonNameLower);
        });
        return (
          <div className="flex items-center gap-3">
            {addon.icon ? (
              <img src={addon.icon} alt="" className="w-8 h-8 rounded-lg object-cover" />
            ) : (
              <div className="w-8 h-8 rounded-lg bg-neutral-700 flex items-center justify-center">
                <Icons.cube className="w-4 h-4 text-neutral-400" />
              </div>
            )}
            <div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-neutral-100">{addon.name}</span>
                {isAddonInstalled && <Icons.check className="w-4 h-4 text-green-500" />}
              </div>
              {addon.author && <div className="text-xs text-neutral-500">by {addon.author}</div>}
            </div>
          </div>
        );
      },
    },
    {
      key: 'downloads',
      header: 'Downloads',
      render: (addon: Addon) => (
        <span className="text-sm text-neutral-400">
          {addon.downloads !== undefined ? Number(addon.downloads).toLocaleString() : '—'}
        </span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right' as const,
      render: (addon: Addon) => (
        <Button variant="ghost" onClick={() => openVersions(addon)}>
          <Icons.download className="w-4 h-4" />
          Install
        </Button>
      ),
    },
  ], [expanded, installed]);

  const installedColumns = useMemo(() => [
    {
      key: 'file',
      header: 'File',
      render: (file: InstalledAddon) => (
        <div className="flex items-center gap-3">
          {file.icon ? (
            <img src={file.icon} alt="" className="w-8 h-8 rounded-lg object-cover" />
          ) : (
            <div className="w-8 h-8 rounded-lg bg-neutral-700 flex items-center justify-center">
              <Icons.cube className="w-4 h-4 text-neutral-400" />
            </div>
          )}
          <span className="text-sm text-neutral-100">{file.name}</span>
        </div>
      ),
    },
    {
      key: 'size',
      header: 'Size',
      render: (file: InstalledAddon) => (
        <span className="text-sm text-neutral-400">{(file.size / 1024).toFixed(1)} KB</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right' as const,
      render: (file: InstalledAddon) =>
        can('file.delete') ? (
          <button
            onClick={() => setDeleteFile(file)}
            className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
          >
            <Icons.trash className="w-4 h-4" />
          </button>
        ) : null,
    },
  ], [can]);

  if (permsLoading || loading.sources) return null;
  if (!can('file.list')) return <PermissionDenied message="You don't have permission to manage addons" />;

  if (sources.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-1 text-sm text-neutral-400">
          <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
          <span>/</span>
          <span className="font-semibold text-neutral-100">Addons</span>
        </div>
        <div className="rounded-xl bg-neutral-800/30 p-12 text-center">
          <div className="w-12 h-12 rounded-full bg-neutral-700/50 flex items-center justify-center mx-auto mb-4">
            <Icons.cube className="w-6 h-6 text-neutral-400" />
          </div>
          <h3 className="text-sm font-medium text-neutral-100 mb-1">No addon sources</h3>
          <p className="text-xs text-neutral-400">This package doesn't have any addon sources configured.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Addons</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Addons</h1>
          <p className="text-sm text-neutral-400">Browse and install addons for your server.</p>
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                {activeSource?.icon && <img src={activeSource.icon} alt="" className="w-4 h-4 rounded" />}
                {activeSource?.name || 'Select source'}
                <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              {sources.map(src => (
                <DropdownMenuItem
                  key={src.id}
                  onSelect={() => { setActiveSource(src); setAddons([]); setSearch(''); setExpanded(new Set()); setPage(1); setTab('browse'); }}
                >
                  {src.icon && <img src={src.icon} alt="" className="w-4 h-4 rounded" />}
                  {src.name}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          {activeSource?.type !== 'modpack' && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                  {tab === 'browse' ? 'Browse' : 'Installed'}
                  <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onSelect={() => setTab('browse')}>Browse</DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setTab('installed')}>Installed</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
        <form onSubmit={handleSearch} className="w-full">
          <Input
            placeholder={tab === 'browse' ? `Search ${activeSource?.name || 'addons'}...` : 'Filter installed...'}
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </form>
        {tab === 'browse' && addons.length > 0 && (
          <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-neutral-400">
            <div className="flex items-center gap-2">
              <span>Rows</span>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    {perPage}
                    <Icons.chevronDown className="w-3 h-3 text-neutral-400" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  {[10, 20, 50, 100].map(n => (
                    <DropdownMenuItem key={n} onSelect={() => handlePerPageChange(n)}>{n}</DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
            <div className="flex items-center gap-1">
              <button onClick={() => handlePageChange(page - 1)} disabled={page <= 1 || loading.search} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800 hover:text-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed">&#8249;</button>
              <span className="px-2 text-neutral-300 font-semibold min-w-[3rem] text-center">{page}</span>
              <button onClick={() => handlePageChange(page + 1)} disabled={!hasMore || loading.search} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800 hover:text-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed">&#8250;</button>
            </div>
          </div>
        )}
      </div>

      <div className="bg-neutral-900/40 rounded-lg p-1">
        {tab === 'browse' ? (
          <Table
            columns={browseColumns}
            data={addons}
            keyField="id"
            loading={loading.search}
            emptyText="Search for addons to get started"
            expandable={{
              isExpanded: addon => expanded.has(addon.id),
              render: addon => (
                <div className="space-y-3">
                  {addon.description && <p className="text-sm text-neutral-300">{addon.description}</p>}
                  <div className="flex items-center gap-4 text-xs text-neutral-500">
                    {addon.author && <span>Author: <span className="text-neutral-300">{addon.author}</span></span>}
                    {addon.downloads !== undefined && <span>Downloads: <span className="text-neutral-300">{Number(addon.downloads).toLocaleString()}</span></span>}
                  </div>
                </div>
              ),
            }}
          />
        ) : (
          <Table
            columns={installedColumns}
            data={filteredInstalled}
            keyField="name"
            loading={loading.installed}
            emptyText={search.trim() ? 'No matching addons' : 'No addons installed yet'}
          />
        )}
      </div>

      <Modal open={!!versionModal} onClose={() => !versionModal?.installing && setVersionModal(null)} title={`Install ${versionModal?.addon.name || ''}`} description={activeSource?.type === 'modpack' ? 'Select a version. This will download all mods and apply configs.' : 'Select a version to install'}>
        <div className="space-y-1 pt-2 max-h-72 overflow-y-auto">
          {versionModal?.loading ? (
            <div className="py-8 text-center"><span className="inline-block w-5 h-5 border-2 border-neutral-400 border-t-transparent rounded-full animate-spin" /></div>
          ) : versionModal?.versions.length === 0 ? (
            <div className="py-8 text-center text-neutral-400">No versions available</div>
          ) : (
            versionModal?.versions.slice(0, 10).map(v => (
              <div key={v.id} className="flex items-center justify-between px-3 py-2.5 rounded-lg hover:bg-neutral-800 transition-colors">
                <span className="text-sm text-neutral-100 font-medium">{v.name}</span>
                <Button variant="ghost" onClick={() => handleInstall(v)} disabled={!!versionModal.installing} loading={versionModal.installing === v.id}>Install</Button>
              </div>
            ))
          )}
        </div>
      </Modal>

      <DeleteAddonModal
        open={!!deleteFile}
        serverId={id || ''}
        sourceId={activeSource?.id || ''}
        fileName={deleteFile?.name || ''}
        onClose={() => setDeleteFile(null)}
        onDeleted={() => deleteFile && setInstalled(i => i.filter(f => f.name !== deleteFile.name))}
      />
    </div>
  );
}
