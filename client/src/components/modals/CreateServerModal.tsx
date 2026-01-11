import { useState, useEffect, useCallback, useRef } from 'react';
import { getAvailableNodes, getAvailablePackages, createServer, Package } from '../../lib/api';
import { notify } from '../feedback/Notification';
import Input from '../ui/Input';
import Wizard from '../feedback/Wizard';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../ui/DropdownMenu';
import { Icons } from '../Icons';

interface Node {
  id: string; name: string; fqdn: string; is_online: boolean; icon?: string;
  system_info: { memory: { total_bytes: number; available_bytes: number; usage_percent: number }; disk: { total_bytes: number; available_bytes: number; usage_percent: number }; cpu: { cores: number; usage_percent: number } };
}

const STEPS = [
  { label: 'Step 1', name: 'Name & Software' },
  { label: 'Step 2', name: 'Location' },
  { label: 'Step 3', name: 'Resources' },
  { label: 'Step 4', name: 'Review' },
];

interface Props { open: boolean; onClose: () => void; onCreated?: () => void; }

export default function CreateServerModal({ open, onClose, onCreated }: Props) {
  const [step, setStep] = useState(0);
  const [visible, setVisible] = useState(false);
  const [animate, setAnimate] = useState(false);
  const [closing, setClosing] = useState(false);
  const [ui, setUi] = useState({ loading: false, loadingData: true });
  const [data, setData] = useState<{ nodes: Node[]; packages: Package[] }>({ nodes: [], packages: [] });
  const [form, setForm] = useState({ name: '', description: '', memory: '1024', cpu: '100', disk: '5120' });
  const [selected, setSelected] = useState<{ pkg: Package | null; node: Node | null }>({ pkg: null, node: null });
  const submittingRef = useRef(false);

  const triggerClose = useCallback(() => {
    if (closing || ui.loading) return;
    setClosing(true);
    setAnimate(false);
    setTimeout(() => {
      setVisible(false);
      setClosing(false);
      document.body.style.overflow = '';
      setStep(0);
      onClose();
    }, 200);
  }, [closing, ui.loading, onClose]);

  useEffect(() => {
    if (open) {
      setClosing(false);
      setVisible(true);
      submittingRef.current = false;
      requestAnimationFrame(() => requestAnimationFrame(() => setAnimate(true)));
      document.body.style.overflow = 'hidden';
      setUi(u => ({ ...u, loadingData: true }));
      Promise.all([getAvailableNodes(), getAvailablePackages()]).then(([nodesRes, pkgsRes]) => {
        setData({ nodes: nodesRes.success && nodesRes.data ? nodesRes.data : [], packages: pkgsRes.success && pkgsRes.data ? pkgsRes.data : [] });
        setUi(u => ({ ...u, loadingData: false }));
      });
    }
  }, [open]);

  useEffect(() => {
    if (!open && visible && !closing) {
      triggerClose();
    }
  }, [open, visible, closing, triggerClose]);

  const canProceed = (): boolean => {
    switch (step) {
      case 0: return !!(form.name.trim() && selected.pkg);
      case 1: return !!selected.node;
      case 2: return parseInt(form.memory) >= 128 && parseInt(form.cpu) >= 25 && parseInt(form.disk) >= 256;
      default: return true;
    }
  };

  const handleCreate = async () => {
    if (!selected.pkg || !selected.node || submittingRef.current) return;
    submittingRef.current = true;
    setUi(u => ({ ...u, loading: true }));

    const ports = (selected.pkg.ports || []).map(p => ({ port: p.default, primary: p.primary }));
    const variables: Record<string, string> = {};
    (selected.pkg.variables || []).forEach(v => { variables[v.name] = v.default; });
    variables['SERVER_MEMORY'] = form.memory;

    const res = await createServer({
      name: form.name.trim(), description: form.description.trim() || undefined, node_id: selected.node.id, package_id: selected.pkg.id,
      memory: parseInt(form.memory), cpu: parseInt(form.cpu), disk: parseInt(form.disk), ports, variables,
    });

    if (res.success) {
      notify('Server Created', `${form.name} is being set up`, 'success');
      onCreated?.();
      setForm({ name: '', description: '', memory: '1024', cpu: '100', disk: '5120' });
      setSelected({ pkg: null, node: null });
      triggerClose();
    } else {
      notify('Error', res.error || 'Failed to create server', 'error');
      setUi(u => ({ ...u, loading: false }));
      submittingRef.current = false;
    }
  };

  const formatBytes = (bytes: number) => { const gb = bytes / (1024 * 1024 * 1024); return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / (1024 * 1024)).toFixed(0)} MB`; };

  if (!visible) return null;

  return (
    <Wizard steps={STEPS} currentStep={step} onStepChange={setStep} onClose={triggerClose} onComplete={handleCreate} canProceed={canProceed()} loading={ui.loading} completeLabel="Create Server" animate={animate} closing={closing}>
      {ui.loadingData ? (
        <div className="flex items-center justify-center h-[300px] text-neutral-500">Loading...</div>
      ) : (
        <>
          {step === 0 && (
            <div className="space-y-6">
              <Input label="Server Name" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="My Minecraft Server" />
              <Input label="Description (optional)" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} placeholder="A short description of your server" />
              <div className="w-full">
                <label className="block text-sm font-medium text-neutral-300 mb-2">Software</label>
                {data.packages.length === 0 ? <p className="text-neutral-500 text-sm">No packages available.</p> : (
                  <DropdownMenu className="w-full block">
                    <DropdownMenuTrigger asChild>
                      <button className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                        <span className="truncate text-neutral-100 flex items-center gap-2">
                          {selected.pkg ? (
                            <>
                              {selected.pkg.icon ? <img src={selected.pkg.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.cube className="w-5 h-5 text-neutral-400" />}
                              {selected.pkg.name} ({selected.pkg.docker_image})
                            </>
                          ) : 'Select a software type'}
                        </span>
                        <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
                      {data.packages.map(pkg => (
                        <DropdownMenuItem key={pkg.id} onSelect={() => setSelected(s => ({ ...s, pkg }))} className={selected.pkg?.id === pkg.id ? 'bg-neutral-700/50' : ''}>
                          <span className="truncate flex-1 flex items-center gap-2">
                            {pkg.icon ? <img src={pkg.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.cube className="w-5 h-5 text-neutral-400" />}
                            {pkg.name} ({pkg.docker_image})
                          </span>
                          {selected.pkg?.id === pkg.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="w-full">
              <label className="block text-sm font-medium text-neutral-300 mb-2">Node</label>
              {data.nodes.length === 0 ? <p className="text-neutral-500 text-sm">No online nodes available.</p> : (
                <DropdownMenu className="w-full block">
                  <DropdownMenuTrigger asChild>
                    <button className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                      <span className="truncate text-neutral-100 flex items-center gap-2">
                        {selected.node ? (
                          <>
                            {selected.node.icon ? <img src={selected.node.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.server className="w-5 h-5 text-neutral-400" />}
                            {selected.node.name} ({selected.node.fqdn})
                          </>
                        ) : 'Select a node'}
                      </span>
                      <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
                    {data.nodes.map(node => (
                      <DropdownMenuItem key={node.id} onSelect={() => setSelected(s => ({ ...s, node }))} className={selected.node?.id === node.id ? 'bg-neutral-700/50' : ''}>
                        <span className="truncate flex-1 flex items-center gap-2">
                          {node.icon ? <img src={node.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.server className="w-5 h-5 text-neutral-400" />}
                          {node.name} ({node.fqdn}){node.system_info && <span className="text-neutral-500 ml-1">— {formatBytes(node.system_info.memory.available_bytes)} free</span>}
                        </span>
                        {selected.node?.id === node.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          )}

          {step === 2 && (
            <div className="space-y-6">
              <Input label="Memory (MB)" type="number" value={form.memory} onChange={e => setForm(f => ({ ...f, memory: e.target.value }))} min={128} />
              <Input label="CPU (%)" type="number" value={form.cpu} onChange={e => setForm(f => ({ ...f, cpu: e.target.value }))} min={25} max={400} />
              <p className="text-xs text-neutral-500 -mt-4">100% = 1 CPU core</p>
              <Input label="Disk (MB)" type="number" value={form.disk} onChange={e => setForm(f => ({ ...f, disk: e.target.value }))} min={256} />
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <div className="rounded-lg bg-neutral-900/50 border border-neutral-800 overflow-hidden">
                <div className="px-4 py-3 border-b border-neutral-800">
                  <h4 className="text-sm font-medium text-neutral-100">Server Summary</h4>
                </div>
                <div className="p-4 space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Name</span>
                    <span className="text-neutral-100 font-medium">{form.name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Software</span>
                    <span className="text-neutral-100">{selected.pkg?.name || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Docker Image</span>
                    <span className="text-neutral-100 font-mono text-xs">{selected.pkg?.docker_image || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Node</span>
                    <span className="text-neutral-100">{selected.node?.name || '-'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Location</span>
                    <span className="text-neutral-100 font-mono text-xs">{selected.node?.fqdn || '-'}</span>
                  </div>
                </div>
              </div>

              <div className="rounded-lg bg-neutral-900/50 border border-neutral-800 overflow-hidden">
                <div className="px-4 py-3 border-b border-neutral-800">
                  <h4 className="text-sm font-medium text-neutral-100">Resources</h4>
                </div>
                <div className="p-4 space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Memory</span>
                    <span className="text-neutral-100">{form.memory} MB</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">CPU</span>
                    <span className="text-neutral-100">{form.cpu}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-neutral-400">Disk</span>
                    <span className="text-neutral-100">{form.disk} MB</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </Wizard>
  );
}
