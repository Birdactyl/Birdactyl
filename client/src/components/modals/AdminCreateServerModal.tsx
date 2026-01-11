import { useState } from 'react';
import { adminCreateServer, type Node, type Package } from '../../lib/api';
import { notify, Modal, Input, Button, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../';

interface Props {
  nodes: Node[];
  packages: Package[];
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}

export default function AdminCreateServerModal({ nodes, packages, open, onClose, onCreated }: Props) {
  const [name, setName] = useState('');
  const [userId, setUserId] = useState('');
  const [nodeId, setNodeId] = useState('');
  const [packageId, setPackageId] = useState('');
  const [memory, setMemory] = useState('1024');
  const [cpu, setCpu] = useState('100');
  const [disk, setDisk] = useState('5120');
  const [loading, setLoading] = useState(false);

  const resetForm = () => {
    setName('');
    setUserId('');
    setNodeId('');
    setPackageId('');
    setMemory('1024');
    setCpu('100');
    setDisk('5120');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nodeId || !packageId) return;
    setLoading(true);
    const res = await adminCreateServer({
      name,
      node_id: nodeId,
      package_id: packageId,
      memory: parseInt(memory) || 1024,
      cpu: parseInt(cpu) || 100,
      disk: parseInt(disk) || 5120,
      user_id: userId || undefined,
    });
    if (res.success) {
      notify('Server Created', `${name} is being set up`, 'success');
      onCreated();
      resetForm();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to create server', 'error');
    }
    setLoading(false);
  };

  const onlineNodes = nodes.filter(n => n.is_online);

  return (
    <Modal open={open} onClose={onClose} title="Create Server" description="Create a new server. Leave User ID blank to create for yourself.">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input label="Server Name" value={name} onChange={e => setName(e.target.value)} placeholder="My Server" />
        <Input label="User ID (optional)" value={userId} onChange={e => setUserId(e.target.value)} placeholder="Leave blank for yourself" />
        <div className="flex flex-col gap-1.5">
          <label className="block text-xs font-medium text-neutral-400">Node</label>
          <DropdownMenu className="w-full block">
            <DropdownMenuTrigger asChild>
              <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                <span className="truncate text-neutral-100 flex items-center gap-2">
                  {nodeId ? (
                    <>
                      {nodes.find(n => n.id === nodeId)?.icon ? <img src={nodes.find(n => n.id === nodeId)?.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.server className="w-5 h-5 text-neutral-400" />}
                      {nodes.find(n => n.id === nodeId)?.name}
                    </>
                  ) : 'Select a node'}
                </span>
                <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
              {onlineNodes.map(node => (
                <DropdownMenuItem key={node.id} onSelect={() => setNodeId(node.id)} className={nodeId === node.id ? 'bg-neutral-700/50' : ''}>
                  <span className="truncate flex-1 flex items-center gap-2">
                    {node.icon ? <img src={node.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.server className="w-5 h-5 text-neutral-400" />}
                    {node.name} ({node.fqdn})
                  </span>
                  {nodeId === node.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="block text-xs font-medium text-neutral-400">Package</label>
          <DropdownMenu className="w-full block">
            <DropdownMenuTrigger asChild>
              <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                <span className="truncate text-neutral-100 flex items-center gap-2">
                  {packageId ? (
                    <>
                      {packages.find(p => p.id === packageId)?.icon ? <img src={packages.find(p => p.id === packageId)?.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.cube className="w-5 h-5 text-neutral-400" />}
                      {packages.find(p => p.id === packageId)?.name} ({packages.find(p => p.id === packageId)?.docker_image})
                    </>
                  ) : 'Select a package'}
                </span>
                <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
              {packages.map(pkg => (
                <DropdownMenuItem key={pkg.id} onSelect={() => setPackageId(pkg.id)} className={packageId === pkg.id ? 'bg-neutral-700/50' : ''}>
                  <span className="truncate flex-1 flex items-center gap-2">
                    {pkg.icon ? <img src={pkg.icon} alt="" className="w-5 h-5 rounded object-cover" /> : <Icons.cube className="w-5 h-5 text-neutral-400" />}
                    {pkg.name} ({pkg.docker_image})
                  </span>
                  {packageId === pkg.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="grid grid-cols-3 gap-3">
          <Input label="Memory (MB)" type="number" value={memory} onChange={e => setMemory(e.target.value)} min={128} />
          <Input label="CPU (%)" type="number" value={cpu} onChange={e => setCpu(e.target.value)} min={25} />
          <Input label="Disk (MB)" type="number" value={disk} onChange={e => setDisk(e.target.value)} min={256} />
        </div>
        <div className="flex justify-end gap-3 pt-4 border-t border-neutral-800">
          <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button type="submit" loading={loading} disabled={!name || !nodeId || !packageId}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
