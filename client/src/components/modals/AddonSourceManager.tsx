import { useState } from 'react';
import { AddonSource, AddonSourceMapping } from '../../lib/api';
import { Icons } from '../';

interface Props {
  sources: AddonSource[];
  onChange: (sources: AddonSource[]) => void;
}

const TEMPLATES: Record<string, AddonSource> = {
  'modrinth-mods-fabric': {
    id: 'modrinth',
    name: 'Modrinth Mods',
    icon: 'https://modrinth.com/favicon.ico',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amod%22%5D%2C%5B%22categories%3Afabric%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22fabric%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: 'mods',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-mods-forge': {
    id: 'modrinth',
    name: 'Modrinth Mods',
    icon: 'https://modrinth.com/favicon.ico',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amod%22%5D%2C%5B%22categories%3Aforge%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22forge%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: 'mods',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-mods-neoforge': {
    id: 'modrinth',
    name: 'Modrinth Mods',
    icon: 'https://modrinth.com/favicon.ico',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amod%22%5D%2C%5B%22categories%3Aneoforge%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22neoforge%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: 'mods',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-plugins-paper': {
    id: 'modrinth',
    name: 'Modrinth Plugins',
    icon: 'https://modrinth.com/favicon.ico',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Aplugin%22%5D%2C%5B%22categories%3Apaper%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22paper%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: 'plugins',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-plugins-velocity': {
    id: 'modrinth',
    name: 'Modrinth Plugins',
    icon: 'https://modrinth.com/favicon.ico',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Aplugin%22%5D%2C%5B%22categories%3Avelocity%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22velocity%22%5D',
    install_path: 'plugins',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'curseforge-mods-fabric': {
    id: 'curseforge',
    name: 'CurseForge Mods',
    icon: 'https://www.curseforge.com/favicon.ico',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=6&modLoaderType=4&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=4&gameVersion={{MC_VERSION}}',
    install_path: 'mods',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
  'curseforge-mods-forge': {
    id: 'curseforge',
    name: 'CurseForge Mods',
    icon: 'https://www.curseforge.com/favicon.ico',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=6&modLoaderType=1&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=1&gameVersion={{MC_VERSION}}',
    install_path: 'mods',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
  'curseforge-mods-neoforge': {
    id: 'curseforge',
    name: 'CurseForge Mods',
    icon: 'https://www.curseforge.com/favicon.ico',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=6&modLoaderType=6&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=6&gameVersion={{MC_VERSION}}',
    install_path: 'mods',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
  'modrinth-modpacks-fabric': {
    id: 'modrinth-modpacks',
    name: 'Modrinth Modpacks',
    icon: 'https://modrinth.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amodpack%22%5D%2C%5B%22categories%3Afabric%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22fabric%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: '',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-modpacks-forge': {
    id: 'modrinth-modpacks',
    name: 'Modrinth Modpacks',
    icon: 'https://modrinth.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amodpack%22%5D%2C%5B%22categories%3Aforge%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22forge%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: '',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'modrinth-modpacks-neoforge': {
    id: 'modrinth-modpacks',
    name: 'Modrinth Modpacks',
    icon: 'https://modrinth.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.modrinth.com/v2/search?query={{query}}&facets=%5B%5B%22project_type%3Amodpack%22%5D%2C%5B%22categories%3Aneoforge%22%5D%2C%5B%22versions%3A{{MC_VERSION}}%22%5D%5D&limit={{limit}}&offset={{offset}}',
    versions_url: 'https://api.modrinth.com/v2/project/{{id}}/version?loaders=%5B%22neoforge%22%5D&game_versions=%5B%22{{MC_VERSION}}%22%5D',
    install_path: '',
    mapping: { results: 'hits', id: 'project_id', name: 'title', description: 'description', icon: 'icon_url', author: 'author', downloads: 'downloads', version_id: 'id', version_name: 'version_number', download_url: 'files.0.url', file_name: 'files.0.filename' }
  },
  'curseforge-modpacks-fabric': {
    id: 'curseforge-modpacks',
    name: 'CurseForge Modpacks',
    icon: 'https://www.curseforge.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=4471&modLoaderType=4&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=4&gameVersion={{MC_VERSION}}',
    install_path: '',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
  'curseforge-modpacks-forge': {
    id: 'curseforge-modpacks',
    name: 'CurseForge Modpacks',
    icon: 'https://www.curseforge.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=4471&modLoaderType=1&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=1&gameVersion={{MC_VERSION}}',
    install_path: '',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
  'curseforge-modpacks-neoforge': {
    id: 'curseforge-modpacks',
    name: 'CurseForge Modpacks',
    icon: 'https://www.curseforge.com/favicon.ico',
    type: 'modpack',
    search_url: 'https://api.curseforge.com/v1/mods/search?gameId=432&classId=4471&modLoaderType=6&gameVersion={{MC_VERSION}}&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}',
    versions_url: 'https://api.curseforge.com/v1/mods/{{id}}/files?modLoaderType=6&gameVersion={{MC_VERSION}}',
    install_path: '',
    mapping: { results: 'data', id: 'id', name: 'name', description: 'summary', icon: 'logo.url', author: 'authors.0.name', downloads: 'downloadCount', version_id: 'id', version_name: 'displayName', download_url: 'downloadUrl', file_name: 'fileName' }
  },
};

const TEMPLATE_GROUPS = [
  { label: 'Modrinth Mods', items: [
    { key: 'modrinth-mods-fabric', label: 'Fabric' },
    { key: 'modrinth-mods-forge', label: 'Forge' },
    { key: 'modrinth-mods-neoforge', label: 'NeoForge' },
  ]},
  { label: 'Modrinth Plugins', items: [
    { key: 'modrinth-plugins-paper', label: 'Paper' },
    { key: 'modrinth-plugins-velocity', label: 'Velocity' },
  ]},
  { label: 'Modrinth Modpacks', items: [
    { key: 'modrinth-modpacks-fabric', label: 'Fabric' },
    { key: 'modrinth-modpacks-forge', label: 'Forge' },
    { key: 'modrinth-modpacks-neoforge', label: 'NeoForge' },
  ]},
  { label: 'CurseForge Mods', items: [
    { key: 'curseforge-mods-fabric', label: 'Fabric' },
    { key: 'curseforge-mods-forge', label: 'Forge' },
    { key: 'curseforge-mods-neoforge', label: 'NeoForge' },
  ]},
  { label: 'CurseForge Modpacks', items: [
    { key: 'curseforge-modpacks-fabric', label: 'Fabric' },
    { key: 'curseforge-modpacks-forge', label: 'Forge' },
    { key: 'curseforge-modpacks-neoforge', label: 'NeoForge' },
  ]},
];

const defaultMapping: AddonSourceMapping = {
  results: 'data', id: 'id', name: 'name', description: 'description', icon: 'icon', author: 'author', downloads: 'downloads',
  version_id: 'id', version_name: 'name', download_url: 'download_url', file_name: 'file_name'
};

export default function AddonSourceManager({ sources, onChange }: Props) {
  const [mode, setMode] = useState<'template' | 'json'>('template');
  const [jsonText, setJsonText] = useState('');
  const [jsonError, setJsonError] = useState('');
  const [editing, setEditing] = useState(-1);

  const addFromTemplate = (key: string) => {
    const template = TEMPLATES[key];
    if (!template) return;
    const exists = sources.some(s => s.id === template.id);
    if (exists) {
      const newId = `${template.id}-${Date.now()}`;
      onChange([...sources, { ...template, id: newId }]);
    } else {
      onChange([...sources, { ...template }]);
    }
  };

  const addFromJson = () => {
    try {
      const parsed = JSON.parse(jsonText);
      if (!parsed.id || !parsed.name || !parsed.search_url) {
        setJsonError('Missing required fields: id, name, search_url');
        return;
      }
      onChange([...sources, { ...parsed, mapping: parsed.mapping || defaultMapping }]);
      setJsonText('');
      setJsonError('');
    } catch {
      setJsonError('Invalid JSON');
    }
  };

  const edit = (index: number) => {
    setEditing(index);
    setMode('json');
    setJsonText(JSON.stringify(sources[index], null, 2));
  };

  const updateFromJson = () => {
    if (editing < 0) return;
    try {
      const parsed = JSON.parse(jsonText);
      if (!parsed.id || !parsed.name || !parsed.search_url) {
        setJsonError('Missing required fields: id, name, search_url');
        return;
      }
      onChange(sources.map((s, i) => i === editing ? { ...parsed, mapping: parsed.mapping || defaultMapping } : s));
      setEditing(-1);
      setJsonText('');
      setJsonError('');
    } catch {
      setJsonError('Invalid JSON');
    }
  };

  const remove = (index: number) => {
    onChange(sources.filter((_, i) => i !== index));
    if (editing === index) {
      setEditing(-1);
      setJsonText('');
    }
  };

  const cancelEdit = () => {
    setEditing(-1);
    setJsonText('');
    setJsonError('');
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 mb-2">
        <button
          onClick={() => { setMode('template'); cancelEdit(); }}
          className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${mode === 'template' ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-400 hover:text-neutral-100'}`}
        >
          Templates
        </button>
        <button
          onClick={() => setMode('json')}
          className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${mode === 'json' ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-400 hover:text-neutral-100'}`}
        >
          Custom JSON
        </button>
      </div>

      {mode === 'template' ? (
        <div className="p-4 rounded-lg bg-neutral-900/50 border border-neutral-800 space-y-4">
          <p className="text-xs text-neutral-400">Select a template to add a pre-configured addon source. Uses {"{{MC_VERSION}}"} variable for version filtering.</p>
          {TEMPLATE_GROUPS.map(group => (
            <div key={group.label}>
              <div className="text-xs font-medium text-neutral-500 mb-2">{group.label}</div>
              <div className="flex flex-wrap gap-2">
                {group.items.map(item => (
                  <button
                    key={item.key}
                    onClick={() => addFromTemplate(item.key)}
                    className="px-3 py-1.5 text-sm font-medium text-neutral-300 bg-neutral-800 hover:bg-neutral-700 rounded-lg transition-colors"
                  >
                    {item.label}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="p-4 rounded-lg bg-neutral-900/50 border border-neutral-800 space-y-3">
          <div>
            <label className="block text-xs font-medium text-neutral-400 mb-1.5">
              {editing >= 0 ? 'Edit Addon Source JSON' : 'Addon Source JSON'}
            </label>
            <textarea
              value={jsonText}
              onChange={e => { setJsonText(e.target.value); setJsonError(''); }}
              placeholder='{"id": "custom", "name": "Custom Source", "search_url": "...", "install_path": "mods", "mapping": {...}}'
              rows={8}
              className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 px-3 py-2 text-sm font-mono resize-none"
              spellCheck={false}
            />
            {jsonError && <p className="mt-1 text-xs text-red-400">{jsonError}</p>}
          </div>
          <div className="flex gap-2">
            {editing >= 0 ? (
              <>
                <button
                  onClick={updateFromJson}
                  className="flex-1 py-2 text-sm font-medium text-neutral-300 bg-neutral-800 hover:bg-neutral-700 rounded-lg transition-colors"
                >
                  Update Source
                </button>
                <button
                  onClick={cancelEdit}
                  className="px-4 py-2 text-sm font-medium text-neutral-400 hover:text-neutral-100 rounded-lg transition-colors"
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                onClick={addFromJson}
                disabled={!jsonText.trim()}
                className="w-full py-2 text-sm font-medium text-neutral-300 bg-neutral-800 hover:bg-neutral-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Add Source
              </button>
            )}
          </div>
        </div>
      )}

      {sources.length > 0 ? (
        <div className="space-y-2">
          <div className="text-xs font-medium text-neutral-400">Added Sources</div>
          {sources.map((source, i) => (
            <div key={i} className={`flex items-center justify-between p-3 rounded-lg ${editing === i ? 'bg-amber-500/20 ring-1 ring-amber-500' : 'bg-neutral-800/50'}`}>
              <div className="flex items-center gap-3">
                {source.icon ? (
                  <img src={source.icon} alt="" className="w-6 h-6 rounded" />
                ) : (
                  <div className="w-6 h-6 rounded bg-neutral-700 flex items-center justify-center">
                    <Icons.cube className="w-3 h-3 text-neutral-400" />
                  </div>
                )}
                <div>
                  <span className="text-sm font-medium text-neutral-100">{source.name}</span>
                  <span className="text-xs text-neutral-500 ml-2">({source.id})</span>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <button onClick={() => edit(i)} className="text-neutral-400 hover:text-neutral-100 transition-colors p-1">
                  <Icons.edit className="w-4 h-4" />
                </button>
                <button onClick={() => remove(i)} className="text-neutral-400 hover:text-red-400 transition-colors p-1">
                  <Icons.x className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-sm text-neutral-500 text-center py-4">No addon sources added yet. Add sources to enable addon browsing for servers using this package.</p>
      )}
    </div>
  );
}
