import { useState, useEffect } from 'react';
import { Icons, Input, Table, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, Button, Modal } from '../../../components';
import { notify } from '../../../components/feedback/Notification';
import { api } from '../../../lib/api';

interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  description: string;
  html_url: string;
  stargazers_count: number;
  owner: { login: string; avatar_url: string };
  updated_at: string;
}

interface GitHubRelease {
  id: number;
  tag_name: string;
  name: string;
  published_at: string;
  assets: { name: string; browser_download_url: string }[];
}

interface PluginFile {
  name: string;
  size: number;
  type?: 'java' | 'go';
  repo?: string;
  owner_name?: string;
  owner_avatar?: string;
  description?: string;
}

export default function MarketplacePage() {
  const [tab, setTab] = useState<'browse' | 'installed'>('browse');
  const [plugins, setPlugins] = useState<GitHubRepo[]>([]);
  const [installed, setInstalled] = useState<PluginFile[]>([]);
  const [loading, setLoading] = useState({ browse: true, installed: false });
  const [search, setSearch] = useState('');
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [totalCount, setTotalCount] = useState(0);
  const [installModal, setInstallModal] = useState<{ repo: GitHubRepo; releases: GitHubRelease[]; loading: boolean; installing: string | null } | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ file: PluginFile; loading: boolean } | null>(null);
  const [buildTools, setBuildTools] = useState<{ maven: boolean; go: boolean }>({ maven: false, go: false });

  useEffect(() => {
    api.get<{ maven_available: boolean; go_available: boolean }>('/admin/plugins/config').then(res => {
      if (res.success && res.data) setBuildTools({ maven: res.data.maven_available, go: res.data.go_available });
    });
  }, []);

  useEffect(() => {
    if (tab === 'browse') fetchPlugins();
    else loadInstalled();
  }, [tab, page, perPage]);

  const fetchPlugins = async () => {
    setLoading(s => ({ ...s, browse: true }));
    try {
      const res = await fetch(`https://api.github.com/search/repositories?q=topic:birdactyl-plugin&sort=stars&order=desc&per_page=${perPage}&page=${page}`);
      const data = await res.json();
      setPlugins(data.items || []);
      setTotalCount(data.total_count || 0);
    } catch {
      setPlugins([]);
    }
    setLoading(s => ({ ...s, browse: false }));
  };

  const loadInstalled = async () => {
    setLoading(s => ({ ...s, installed: true }));
    const res = await api.get<{ files: PluginFile[] }>('/admin/plugins/files');
    if (res.success && res.data) setInstalled(res.data.files || []);
    setLoading(s => ({ ...s, installed: false }));
  };

  const filtered = search.trim()
    ? plugins.filter(p => p.name.toLowerCase().includes(search.toLowerCase()) || p.description?.toLowerCase().includes(search.toLowerCase()))
    : plugins;

  const filteredInstalled = search.trim()
    ? installed.filter(f => f.name.toLowerCase().includes(search.toLowerCase()))
    : installed;

  const totalPages = Math.ceil(totalCount / perPage) || 1;

  const toggleExpand = (id: number) => {
    const next = new Set(expanded);
    next.has(id) ? next.delete(id) : next.add(id);
    setExpanded(next);
  };

  const openInstallModal = async (repo: GitHubRepo) => {
    setInstallModal({ repo, releases: [], loading: true, installing: null });
    try {
      const res = await fetch(`https://api.github.com/repos/${repo.full_name}/releases?per_page=10`);
      const releases = await res.json();
      setInstallModal(s => s && { ...s, releases: Array.isArray(releases) ? releases : [], loading: false });
    } catch {
      setInstallModal(s => s && { ...s, releases: [], loading: false });
    }
  };

  const installFromSource = async () => {
    if (!installModal) return;
    setInstallModal(s => s && { ...s, installing: 'source' });
    const res = await api.post<{ file: string }>('/admin/plugins/install-source', {
      repo: installModal.repo.full_name,
      owner_name: installModal.repo.owner.login,
      owner_avatar: installModal.repo.owner.avatar_url,
      description: installModal.repo.description || '',
    });
    if (res.success) {
      notify('Installed', `${installModal.repo.name} compiled and installed`, 'success');
      setInstallModal(null);
    } else {
      notify('Error', res.error || 'Failed to install', 'error');
      setInstallModal(s => s && { ...s, installing: null });
    }
  };

  const installFromRelease = async (release: GitHubRelease, asset: { name: string; browser_download_url: string }) => {
    if (!installModal) return;
    setInstallModal(s => s && { ...s, installing: `release-${release.id}-${asset.name}` });
    const res = await api.post<{ file: string }>('/admin/plugins/install-release', {
      url: asset.browser_download_url,
      filename: asset.name,
      repo: installModal.repo.full_name,
      owner_name: installModal.repo.owner.login,
      owner_avatar: installModal.repo.owner.avatar_url,
      description: installModal.repo.description || '',
    });
    if (res.success) {
      notify('Installed', `${asset.name} installed`, 'success');
      setInstallModal(null);
    } else {
      notify('Error', res.error || 'Failed to install', 'error');
      setInstallModal(s => s && { ...s, installing: null });
    }
  };

  const handleDelete = async () => {
    if (!deleteModal) return;
    setDeleteModal(s => s && { ...s, loading: true });
    const res = await api.delete(`/admin/plugins/file/${encodeURIComponent(deleteModal.file.name)}`);
    if (res.success) {
      notify('Deleted', `${deleteModal.file.name} removed`, 'success');
      setDeleteModal(null);
      loadInstalled();
    } else {
      notify('Error', res.error || 'Failed to delete', 'error');
      setDeleteModal(s => s && { ...s, loading: false });
    }
  };

  const browseColumns = [
    {
      key: 'expand',
      header: '',
      className: 'w-8',
      render: (repo: GitHubRepo) => (
        <button onClick={() => toggleExpand(repo.id)} className="text-neutral-400 hover:text-neutral-200 transition">
          <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(repo.id) ? 'rotate-90' : ''}`} />
        </button>
      ),
    },
    {
      key: 'plugin',
      header: 'Plugin',
      render: (repo: GitHubRepo) => (
        <div className="flex items-center gap-3">
          <img src={repo.owner.avatar_url} alt="" className="w-8 h-8 rounded-lg" />
          <div>
            <span className="text-sm font-medium text-neutral-100">{repo.name}</span>
            <div className="text-xs text-neutral-500">by {repo.owner.login}</div>
          </div>
        </div>
      ),
    },
    {
      key: 'stars',
      header: 'Stars',
      render: (repo: GitHubRepo) => (
        <span className="text-sm text-neutral-400">{repo.stargazers_count.toLocaleString()}</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right' as const,
      render: (repo: GitHubRepo) => (
        <div className="flex items-center justify-end gap-2">
          <Button variant="ghost" onClick={() => openInstallModal(repo)}>
            <Icons.download className="w-4 h-4" />
            Install
          </Button>
          <a href={repo.html_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 rounded-lg transition-colors">
            <Icons.globe className="w-4 h-4" />
            View
          </a>
        </div>
      ),
    },
  ];

  const installedColumns = [
    {
      key: 'file',
      header: 'Plugin',
      render: (file: PluginFile) => (
        <div className="flex items-center gap-3">
          {file.owner_avatar ? (
            <img src={file.owner_avatar} alt="" className="w-8 h-8 rounded-lg" />
          ) : (
            <div className="w-8 h-8 rounded-lg bg-neutral-700 flex items-center justify-center">
              <Icons.cube className="w-4 h-4 text-neutral-400" />
            </div>
          )}
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-neutral-100">{file.name.replace('.jar', '')}</span>
              <span className={`text-[10px] px-1.5 py-0.5 rounded ${file.type === 'go' ? 'bg-cyan-500/20 text-cyan-400' : 'bg-orange-500/20 text-orange-400'}`}>
                {file.type === 'go' ? 'Go' : 'Java'}
              </span>
            </div>
            {file.owner_name && <div className="text-xs text-neutral-500">by {file.owner_name}</div>}
          </div>
        </div>
      ),
    },
    {
      key: 'size',
      header: 'Size',
      render: (file: PluginFile) => (
        <span className="text-sm text-neutral-400">{(file.size / 1024 / 1024).toFixed(2)} MB</span>
      ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right' as const,
      render: (file: PluginFile) => (
        <button onClick={() => setDeleteModal({ file, loading: false })} className="p-1.5 text-neutral-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors">
          <Icons.trash className="w-4 h-4" />
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-neutral-100">Plugin Marketplace</h1>
        <p className="text-sm text-neutral-400">Browse and manage Birdactyl plugins.</p>
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
              {tab === 'browse' ? 'Browse' : 'Installed'}
              <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem onSelect={() => { setTab('browse'); setSearch(''); }}>Browse</DropdownMenuItem>
            <DropdownMenuItem onSelect={() => { setTab('installed'); setSearch(''); }}>Installed</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <form onSubmit={e => e.preventDefault()} className="flex-1 sm:max-w-xl">
          <Input placeholder={tab === 'browse' ? 'Search plugins...' : 'Filter installed...'} value={search} onChange={e => setSearch(e.target.value)} />
        </form>
        <button onClick={() => tab === 'browse' ? fetchPlugins() : loadInstalled()} disabled={loading.browse || loading.installed} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 transition-colors disabled:opacity-50 flex-shrink-0">
          <Icons.refresh className={`w-4 h-4 ${(loading.browse || loading.installed) ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {tab === 'browse' && (
        <div className="flex flex-wrap items-center gap-3 text-xs text-neutral-400">
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
                  <DropdownMenuItem key={n} onSelect={() => { setPerPage(n); setPage(1); }}>{n}</DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div className="flex items-center gap-1">
            <button onClick={() => setPage(p => p - 1)} disabled={page <= 1 || loading.browse} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800 hover:text-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed">‹</button>
            <span className="px-2 text-neutral-300 font-semibold min-w-[3rem] text-center">{page} / {totalPages}</span>
            <button onClick={() => setPage(p => p + 1)} disabled={page >= totalPages || loading.browse} className="h-9 w-9 rounded-lg inline-flex items-center justify-center text-lg font-medium text-neutral-400 hover:bg-neutral-800 hover:text-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed">›</button>
          </div>
        </div>
      )}

      <div className="bg-neutral-900/40 rounded-lg p-1">
        {tab === 'browse' ? (
          <Table
            columns={browseColumns}
            data={filtered}
            keyField="id"
            loading={loading.browse}
            emptyText="No plugins found"
            expandable={{
              isExpanded: repo => expanded.has(repo.id),
              render: repo => (
                <div className="space-y-2">
                  {repo.description && <p className="text-sm text-neutral-300">{repo.description}</p>}
                  <div className="text-xs text-neutral-500">Updated: {new Date(repo.updated_at).toLocaleDateString()}</div>
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
            emptyText="No plugins installed"
          />
        )}
      </div>

      <Modal open={!!installModal} onClose={() => !installModal?.installing && setInstallModal(null)} title={`Install ${installModal?.repo.name || ''}`} description="Choose installation method">
        <div className="space-y-4 pt-2">
          <div className={`rounded-lg border p-3 ${!buildTools.maven && !buildTools.go ? 'border-red-800/50 bg-red-900/10' : 'border-neutral-800'}`}>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm font-medium text-neutral-100">Build from source</div>
                {!buildTools.maven && !buildTools.go ? (
                  <div className="text-xs text-red-400">No build tools available (Maven/Go not installed)</div>
                ) : (
                  <div className="text-xs text-neutral-500">
                    Clone and compile • 
                    <span className={buildTools.maven ? 'text-green-400' : 'text-red-400'}> Maven {buildTools.maven ? '✓' : '✗'}</span>
                    <span className={buildTools.go ? 'text-green-400' : 'text-red-400'}> Go {buildTools.go ? '✓' : '✗'}</span>
                  </div>
                )}
              </div>
              <Button variant="ghost" onClick={installFromSource} disabled={!!installModal?.installing || (!buildTools.maven && !buildTools.go)} loading={installModal?.installing === 'source'}>
                Build
              </Button>
            </div>
          </div>
          <div className="text-xs text-neutral-500 text-center">— or install from release —</div>
          <div className="space-y-1 max-h-64 overflow-y-auto">
            {installModal?.loading ? (
              <div className="py-8 text-center"><span className="inline-block w-5 h-5 border-2 border-neutral-400 border-t-transparent rounded-full animate-spin" /></div>
            ) : installModal?.releases.length === 0 ? (
              <div className="py-4 text-center text-neutral-500 text-sm">No releases available</div>
            ) : (
              installModal?.releases.map(release => (
                <div key={release.id} className="rounded-lg border border-neutral-800 p-3">
                  <div className="text-sm font-medium text-neutral-100 mb-2">{release.tag_name} {release.name && release.name !== release.tag_name && `- ${release.name}`}</div>
                  <div className="text-xs text-neutral-500 mb-2">{new Date(release.published_at).toLocaleDateString()}</div>
                  {release.assets.filter(a => a.name.endsWith('.jar')).length === 0 ? (
                    <div className="text-xs text-neutral-500">No .jar assets</div>
                  ) : (
                    <div className="space-y-1">
                      {release.assets.filter(a => a.name.endsWith('.jar')).map(asset => (
                        <div key={asset.name} className="flex items-center justify-between">
                          <span className="text-xs text-neutral-300 font-mono">{asset.name}</span>
                          <Button variant="ghost" onClick={() => installFromRelease(release, asset)} disabled={!!installModal?.installing} loading={installModal?.installing === `release-${release.id}-${asset.name}`}>
                            Install
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Delete Plugin" description={`Are you sure you want to delete "${deleteModal?.file.name}"? This cannot be undone.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button variant="danger" onClick={handleDelete} loading={deleteModal?.loading}>Delete</Button>
        </div>
      </Modal>
    </div>
  );
}
