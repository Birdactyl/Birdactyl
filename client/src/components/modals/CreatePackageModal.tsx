import { useState, useEffect, useCallback, useRef } from 'react';
import { createPortal } from 'react-dom';
import { adminCreatePackage, adminUpdatePackage, Package } from '../../lib/api';
import { notify } from '../feedback/Notification';
import { usePackageForm } from '../../hooks/usePackageForm';
import PortManager from './PortManager';
import VariableManager from './VariableManager';
import AddonSourceManager from './AddonSourceManager';
import { Input, Wizard, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, Icons, Button, Checkbox } from '../';

const STEPS = [
  { label: 'Step 1', name: 'Basic Info' },
  { label: 'Step 2', name: 'Docker & Startup' },
  { label: 'Step 3', name: 'Ports' },
  { label: 'Step 4', name: 'Variables' },
  { label: 'Step 5', name: 'Addon Sources' },
  { label: 'Step 6', name: 'Review' },
];

const EXAMPLE_PACKAGE = `{
  "name": "Minecraft Paper",
  "version": "1.21.4",
  "author": "Birdactyl",
  "description": "High performance Minecraft server using PaperMC",
  "docker_image": "eclipse-temurin:21-jre-alpine",
  "install_image": "eclipse-temurin:21-jdk-alpine",
  "startup": "java -Xms128M -Xmx{{SERVER_MEMORY}}M -jar server.jar --nogui",
  "install_script": "cd /home/container\\nPAPER_BUILD=$(curl -s https://api.papermc.io/v2/projects/paper/versions/{{MC_VERSION}}/builds | jq -r '.builds[-1].build')\\ncurl -o server.jar https://api.papermc.io/v2/projects/paper/versions/{{MC_VERSION}}/builds/\${PAPER_BUILD}/downloads/paper-{{MC_VERSION}}-\${PAPER_BUILD}.jar\\necho \\"eula=true\\" > eula.txt",
  "stop_signal": "SIGINT",
  "stop_command": "stop",
  "stop_timeout": 30,
  "ports": [
    { "name": "Game", "default": 25565, "protocol": "tcp", "primary": true },
    { "name": "RCON", "default": 25575, "protocol": "tcp" }
  ],
  "variables": [
    { "name": "SERVER_MEMORY", "default": "1024", "description": "Server memory in MB", "user_editable": false },
    { "name": "MC_VERSION", "default": "1.21.4", "description": "Minecraft version", "user_editable": false },
    { "name": "MOTD", "default": "A Birdactyl Server", "description": "Server message of the day", "user_editable": true }
  ],
  "addon_sources": [
    {
      "id": "modrinth",
      "name": "Modrinth",
      "icon": "https://modrinth.com/favicon.ico",
      "search_url": "https://api.modrinth.com/v2/search?query={{query}}&limit={{limit}}&offset={{offset}}&facets=[[%22project_type:plugin%22],[%22categories:paper%22],[%22versions:{{MC_VERSION}}%22]]",
      "versions_url": "https://api.modrinth.com/v2/project/{{id}}/version?game_versions=[%22{{MC_VERSION}}%22]",
      "install_path": "/plugins",
      "mapping": {
        "results": "hits",
        "id": "project_id",
        "name": "title",
        "description": "description",
        "icon": "icon_url",
        "author": "author",
        "downloads": "downloads",
        "version_id": "id",
        "version_name": "version_number",
        "download_url": "files.0.url",
        "file_name": "files.0.filename"
      }
    },
    {
      "id": "curseforge",
      "name": "CurseForge",
      "icon": "https://www.curseforge.com/favicon.ico",
      "api_key": "curseforge",
      "search_url": "https://api.curseforge.com/v1/mods/search?gameId=432&classId=6&searchFilter={{query}}&pageSize={{limit}}&index={{offset}}",
      "versions_url": "https://api.curseforge.com/v1/mods/{{id}}/files",
      "install_path": "/plugins",
      "mapping": {
        "results": "data",
        "id": "id",
        "name": "name",
        "description": "summary",
        "icon": "logo.url",
        "author": "authors.0.name",
        "downloads": "downloadCount",
        "version_id": "id",
        "version_name": "displayName",
        "download_url": "downloadUrl",
        "file_name": "fileName"
      }
    }
  ]
}`;

interface Props {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
  editPackage?: Package | null;
}

export default function CreatePackageModal({ open, onClose, onCreated, editPackage }: Props) {
  const [mode, setMode] = useState<'wizard' | 'import'>('wizard');
  const [step, setStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [animate, setAnimate] = useState(false);
  const [closing, setClosing] = useState(false);
  const [jsonText, setJsonText] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const submittingRef = useRef(false);
  const { data, update, toJson, fromJson, toApiData, reset } = usePackageForm(editPackage);

  const triggerClose = useCallback(() => {
    if (closing || loading) return;
    setClosing(true);
    setAnimate(false);
    setTimeout(() => {
      setVisible(false);
      setClosing(false);
      document.body.style.overflow = '';
      setStep(0);
      setMode('wizard');
      onClose();
    }, 200);
  }, [closing, loading, onClose]);

  useEffect(() => {
    if (open) {
      setClosing(false);
      setVisible(true);
      submittingRef.current = false;
      setLoading(false);
      requestAnimationFrame(() => requestAnimationFrame(() => setAnimate(true)));
      document.body.style.overflow = 'hidden';
      if (editPackage) {
        reset(editPackage);
      }
    }
  }, [open, editPackage, reset]);

  useEffect(() => {
    if (!open && visible && !closing) {
      triggerClose();
    }
  }, [open, visible, closing, triggerClose]);

  useEffect(() => {
    if (!visible) return;
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') triggerClose(); };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [visible, triggerClose]);

  const canProceed = () => {
    switch (step) {
      case 0: return data.name.trim() !== '';
      case 1: return data.dockerImage.trim() !== '' && data.startup.trim() !== '';
      default: return true;
    }
  };

  const handleCreate = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = editPackage
      ? await adminUpdatePackage(editPackage.id, toApiData())
      : await adminCreatePackage(toApiData());

    if (res.success) {
      notify('Success', editPackage ? 'Package updated successfully' : 'Package created successfully', 'success');
      onCreated();
      triggerClose();
    } else {
      notify('Error', res.error || 'Failed to save package', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  const handleFileImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const content = ev.target?.result as string;
      setJsonText(content);
      fromJson(content);
      setMode('import');
    };
    reader.readAsText(file);
    e.target.value = '';
  };

  const renderStep = () => {
    switch (step) {
      case 0:
        return (
          <div className="space-y-4">
            <Input label="Package Name" value={data.name} onChange={e => update('name', e.target.value)} placeholder="Minecraft Paper" required />
            <div className="grid grid-cols-2 gap-4">
              <Input label="Version" value={data.version} onChange={e => update('version', e.target.value)} placeholder="1.21" />
              <Input label="Author" value={data.author} onChange={e => update('author', e.target.value)} placeholder="Birdactyl" />
            </div>
            <Input label="Icon URL (optional)" value={data.icon} onChange={e => update('icon', e.target.value)} placeholder="https://example.com/icon.png" />
            <div>
              <label className="block text-xs font-medium text-neutral-400 mb-1.5">Description</label>
              <textarea
                value={data.description}
                onChange={e => update('description', e.target.value)}
                placeholder="High performance Minecraft server using Paper"
                rows={3}
                className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 px-3 py-2 text-sm resize-none"
              />
            </div>
          </div>
        );

      case 1:
        return (
          <div className="space-y-4">
            <Input label="Docker Image" value={data.dockerImage} onChange={e => update('dockerImage', e.target.value)} placeholder="eclipse-temurin:21-jre" required />
            <Input label="Install Image (optional)" value={data.installImage} onChange={e => update('installImage', e.target.value)} placeholder="eclipse-temurin:21-jdk" />
            <div>
              <label className="block text-xs font-medium text-neutral-400 mb-1.5">
                Startup Command <span className="text-red-500">*</span>
              </label>
              <textarea
                value={data.startup}
                onChange={e => update('startup', e.target.value)}
                placeholder="java -Xms128M -Xmx{{SERVER_MEMORY}}M -jar server.jar --nogui"
                rows={2}
                className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 px-3 py-2 text-sm font-mono resize-none"
              />
              <p className="mt-1 text-xs text-neutral-500">Use {"{{VARIABLE}}"} syntax for variable substitution</p>
            </div>
            <div>
              <label className="block text-xs font-medium text-neutral-400 mb-1.5">Install Script (optional)</label>
              <textarea
                value={data.installScript}
                onChange={e => update('installScript', e.target.value)}
                placeholder="cd /home/container&#10;curl -o server.jar https://..."
                rows={4}
                className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 px-3 py-2 text-sm font-mono resize-none"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="block text-xs font-medium text-neutral-400">Stop Signal</label>
                <DropdownMenu className="w-full">
                  <DropdownMenuTrigger asChild className="w-full">
                    <button type="button" className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 text-left transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-[#0a0a0a] px-3 py-2 flex items-center justify-between" style={{ fontSize: '13px' }}>
                      {data.stopSignal}
                      <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent>
                    <DropdownMenuItem onSelect={() => update('stopSignal', 'SIGTERM')}>SIGTERM</DropdownMenuItem>
                    <DropdownMenuItem onSelect={() => update('stopSignal', 'SIGINT')}>SIGINT</DropdownMenuItem>
                    <DropdownMenuItem onSelect={() => update('stopSignal', 'SIGKILL')}>SIGKILL</DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
              <Input label="Stop Command (optional)" value={data.stopCommand} onChange={e => update('stopCommand', e.target.value)} placeholder="stop" />
            </div>
            <div className="grid grid-cols-1 gap-4">
              <Input label="Stop Timeout (seconds)" type="number" value={data.stopTimeout} onChange={e => update('stopTimeout', e.target.value)} placeholder="30" />
            </div>
            <div className="flex items-center gap-6 pt-2">
              <Checkbox checked={data.startupEditable} onChange={() => update('startupEditable', !data.startupEditable)} label="Allow users to edit startup command" />
              <Checkbox checked={data.dockerImageEditable} onChange={() => update('dockerImageEditable', !data.dockerImageEditable)} label="Allow users to edit Docker image" />
            </div>
          </div>
        );

      case 2:
        return <PortManager ports={data.ports} onChange={ports => update('ports', ports)} />;

      case 3:
        return <VariableManager variables={data.variables} onChange={variables => update('variables', variables)} />;

      case 4:
        return <AddonSourceManager sources={data.addonSources} onChange={sources => update('addonSources', sources)} />;

      case 5:
        return (
          <div className="space-y-4">
            <div className="rounded-lg bg-neutral-900/50 border border-neutral-800 overflow-hidden">
              <div className="px-4 py-3 border-b border-neutral-800">
                <h4 className="text-sm font-medium text-neutral-100">Package Summary</h4>
              </div>
              <div className="p-4 space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-neutral-400">Name</span>
                  <span className="text-neutral-100 font-medium">{data.name || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Version</span>
                  <span className="text-neutral-100">{data.version || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Author</span>
                  <span className="text-neutral-100">{data.author || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Docker Image</span>
                  <span className="text-neutral-100 font-mono text-xs">{data.dockerImage || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Ports</span>
                  <span className="text-neutral-100">{data.ports.length} configured</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Variables</span>
                  <span className="text-neutral-100">{data.variables.length} configured</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">Addon Sources</span>
                  <span className="text-neutral-100">{data.addonSources.length} configured</span>
                </div>
              </div>
            </div>

            {data.description && (
              <div className="rounded-lg bg-neutral-900/50 border border-neutral-800 p-4">
                <div className="text-xs font-medium text-neutral-400 mb-2">Description</div>
                <p className="text-sm text-neutral-300">{data.description}</p>
              </div>
            )}

            <div className="rounded-lg bg-neutral-900/50 border border-neutral-800 p-4">
              <div className="text-xs font-medium text-neutral-400 mb-2">Startup Command</div>
              <code className="text-xs text-neutral-300 font-mono break-all">{data.startup || '-'}</code>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  const modeToggle = (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <button
          onClick={() => setMode('wizard')}
          className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${mode === 'wizard' ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-400 hover:text-neutral-100'}`}
        >
          Wizard
        </button>
        <button
          onClick={() => { setJsonText(toJson()); setMode('import'); }}
          className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${mode === 'import' ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-400 hover:text-neutral-100'}`}
        >
          Import JSON
        </button>
      </div>
      <div>
        <input ref={fileInputRef} type="file" accept=".json" onChange={handleFileImport} className="hidden" />
        <button
          onClick={() => fileInputRef.current?.click()}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 rounded-lg transition-colors"
        >
          <Icons.upload className="w-4 h-4" />
          Import from File
        </button>
      </div>
    </div>
  );

  if (!visible) return null;

  if (mode === 'import') {
    return createPortal(
      <div className={`fixed inset-0 z-[9999] flex transition-opacity duration-300 ${animate && !closing ? 'opacity-100' : 'opacity-0'}`}>
        <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={triggerClose} />
        
        <div className={`relative m-auto w-full max-w-3xl bg-neutral-900 rounded-2xl shadow-2xl border border-neutral-800 overflow-hidden transition-all duration-300 ${animate && !closing ? 'scale-100 translate-y-0' : 'scale-95 translate-y-4'}`}>
          <div className="flex items-center justify-between px-6 py-4 border-b border-neutral-800">
            <h2 className="text-sm font-medium text-neutral-100">Import Package</h2>
            <button
              onClick={triggerClose}
              className="p-2 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-800 transition-colors"
            >
              <Icons.x className="w-5 h-5" />
            </button>
          </div>

          <div className="px-6 pt-6">{modeToggle}</div>

          <div className="px-6 py-8 min-h-[320px] max-h-[60vh] overflow-y-auto">
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-xs font-medium text-neutral-400">Package JSON</label>
                  <button
                    onClick={() => { setJsonText(EXAMPLE_PACKAGE); fromJson(EXAMPLE_PACKAGE); }}
                    className="text-xs text-amber-400 hover:text-violet-300 transition-colors"
                  >
                    Load Example
                  </button>
                </div>
                <textarea
                  value={jsonText}
                  onChange={e => { setJsonText(e.target.value); fromJson(e.target.value); }}
                  placeholder="Paste package JSON here..."
                  rows={12}
                  className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 px-3 py-2 text-sm font-mono resize-none"
                  spellCheck={false}
                />
              </div>
            </div>
          </div>

          <div className="flex items-center justify-between px-6 py-4 bg-neutral-900/50 border-t border-neutral-800">
            <button
              onClick={triggerClose}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-neutral-400 hover:text-neutral-100 rounded-lg transition-colors"
            >
              <Icons.chevronLeft className="w-4 h-4" />
              Cancel
            </button>
            <Button onClick={handleCreate} loading={loading} disabled={!data.name.trim()}>
              {editPackage ? 'Update Package' : 'Create Package'}
            </Button>
          </div>
        </div>
      </div>,
      document.body
    );
  }

  return (
    <Wizard
      steps={STEPS}
      currentStep={step}
      onStepChange={setStep}
      onClose={triggerClose}
      onComplete={handleCreate}
      canProceed={canProceed()}
      canFinish={data.name.trim() !== '' && data.dockerImage.trim() !== '' && data.startup.trim() !== ''}
      loading={loading}
      completeLabel={editPackage ? 'Update Package' : 'Create Package'}
      headerContent={modeToggle}
      animate={animate}
      closing={closing}
    >
      {renderStep()}
    </Wizard>
  );
}
