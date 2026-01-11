import { useState, useEffect, useRef } from 'react';
import { adminUpdateServerResources, adminTransferServer, type Server, type Node, type TransferStatus } from '../../lib/api';
import { notify, Modal, Input, Button, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../';

type EditTab = 'details' | 'resources' | 'transfer';

interface Props {
  server: Server | null;
  nodes: Node[];
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  onTransferStarted: (status: TransferStatus) => void;
}

export default function EditServerModal({ server, nodes, open, onClose, onSaved, onTransferStarted }: Props) {
  const [tab, setTab] = useState<EditTab>('details');
  const [name, setName] = useState('');
  const [ownerId, setOwnerId] = useState('');
  const [memory, setMemory] = useState('');
  const [cpu, setCpu] = useState('');
  const [disk, setDisk] = useState('');
  const [targetNodeId, setTargetNodeId] = useState('');
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (server && open) {
      submittingRef.current = false;
      setLoading(false);
      setName(server.name);
      setOwnerId(server.user_id);
      setMemory(String(server.memory));
      setCpu(String(server.cpu));
      setDisk(String(server.disk));
      setTargetNodeId('');
      setTab('details');
    }
  }, [server, open]);

  if (!server) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);

    if (tab === 'transfer') {
      if (!targetNodeId || targetNodeId === server.node_id) {
        notify('Info', 'Select a different node to transfer', 'info');
        setLoading(false);
        submittingRef.current = false;
        return;
      }
      const res = await adminTransferServer(server.id, targetNodeId);
      if (res.success && res.data) {
        const initialStatus: TransferStatus = {
          id: res.data.transfer_id,
          server_id: server.id,
          server_name: server.name,
          from_node_id: server.node_id,
          from_node_name: server.node?.name || '',
          to_node_id: targetNodeId,
          to_node_name: nodes.find(n => n.id === targetNodeId)?.name || '',
          stage: 'pending',
          progress: 0,
          started_at: new Date().toISOString()
        };
        onTransferStarted(initialStatus);
        onClose();
      } else {
        notify('Error', res.error || 'Failed to start transfer', 'error');
        setLoading(false);
        submittingRef.current = false;
      }
      return;
    }

    const data: { name?: string; user_id?: string; memory?: number; cpu?: number; disk?: number } = {};
    if (name !== server.name) data.name = name;
    if (ownerId !== server.user_id) data.user_id = ownerId;
    const mem = parseInt(memory);
    const cpuVal = parseInt(cpu);
    const diskVal = parseInt(disk);
    if (mem && mem !== server.memory) data.memory = mem;
    if (cpuVal && cpuVal !== server.cpu) data.cpu = cpuVal;
    if (diskVal && diskVal !== server.disk) data.disk = diskVal;

    if (Object.keys(data).length === 0) {
      notify('Info', 'No changes to save', 'info');
      setLoading(false);
      submittingRef.current = false;
      return;
    }

    const res = await adminUpdateServerResources(server.id, data);
    if (res.success) {
      notify('Success', 'Server updated', 'success');
      onSaved();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to update server', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  const availableNodes = nodes.filter(n => n.id !== server.node_id && n.is_online);
  const offlineNodes = nodes.filter(n => n.id !== server.node_id && !n.is_online);

  return (
    <Modal open={open} onClose={onClose} title="Edit Server" description={server.name}>
      <div className="space-y-4">
        <div className="flex gap-1 p-1 bg-neutral-800/50 rounded-lg">
          <button type="button" onClick={() => setTab('details')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'details' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Details</button>
          <button type="button" onClick={() => setTab('resources')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'resources' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Resources</button>
          <button type="button" onClick={() => setTab('transfer')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'transfer' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Transfer</button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          {tab === 'details' && (
            <>
              <Input label="Server Name" value={name} onChange={e => setName(e.target.value)} required />
              <Input label="Owner ID" value={ownerId} onChange={e => setOwnerId(e.target.value)} placeholder="Full or short ID" />
              <div className="pt-2 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Status</span>
                  {server.is_suspended ? (
                    <span className="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 ring-1 ring-inset ring-amber-500/20">Suspended</span>
                  ) : server.status === 'running' ? (
                    <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400">Online</span>
                  ) : (
                    <span className="px-2 py-0.5 rounded bg-neutral-500/10 text-neutral-400">Offline</span>
                  )}
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Owner</span>
                  <span className="text-neutral-300">{server.user?.username || '—'}</span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Node</span>
                  <span className="text-neutral-300">{server.node?.name || '—'}</span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Server ID</span>
                  <span className="text-neutral-500 font-mono text-[10px]">{server.id}</span>
                </div>
              </div>
            </>
          )}
          {tab === 'resources' && (
            <div className="grid grid-cols-3 gap-3">
              <Input label="Memory (MB)" type="number" value={memory} onChange={e => setMemory(e.target.value)} min={128} />
              <Input label="CPU (%)" type="number" value={cpu} onChange={e => setCpu(e.target.value)} min={25} />
              <Input label="Disk (MB)" type="number" value={disk} onChange={e => setDisk(e.target.value)} min={256} />
            </div>
          )}
          {tab === 'transfer' && (
            <div className="space-y-4">
              <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
                <div className="flex items-start gap-2">
                  <Icons.infoCircle className="w-4 h-4 text-amber-400 mt-0.5 flex-shrink-0" />
                  <div className="text-xs text-amber-200">
                    <p className="font-medium">Transfer will:</p>
                    <ul className="mt-1 space-y-0.5 text-amber-300/80">
                      <li>• Stop the server if running</li>
                      <li>• Archive and move all server files</li>
                      <li>• Assign new port allocations</li>
                    </ul>
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="block text-xs font-medium text-neutral-400">Current Node</label>
                <div className="px-3 py-2 rounded-lg bg-neutral-800/50 text-sm text-neutral-300">{server.node?.name || '—'}</div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="block text-xs font-medium text-neutral-400">Target Node</label>
                {availableNodes.length === 0 ? (
                  <p className="text-neutral-500 text-sm">No other online nodes available.</p>
                ) : (
                  <DropdownMenu className="w-full block">
                    <DropdownMenuTrigger asChild>
                      <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                        <span className="truncate text-neutral-100">
                          {targetNodeId ? nodes.find(n => n.id === targetNodeId)?.name + ' (' + nodes.find(n => n.id === targetNodeId)?.fqdn + ')' : 'Select a node'}
                        </span>
                        <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
                      {availableNodes.map(node => (
                        <DropdownMenuItem key={node.id} onSelect={() => setTargetNodeId(node.id)} className={targetNodeId === node.id ? 'bg-neutral-700/50' : ''}>
                          <span className="truncate flex-1">{node.name} ({node.fqdn})</span>
                          {targetNodeId === node.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
                {offlineNodes.length > 0 && (
                  <p className="text-xs text-neutral-500">{offlineNodes.length} offline node(s) hidden</p>
                )}
              </div>
            </div>
          )}
          <div className="flex justify-end gap-3 pt-4 border-t border-neutral-800">
            <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
            <Button type="submit" loading={loading}>{tab === 'transfer' ? 'Transfer Server' : 'Save Changes'}</Button>
          </div>
        </form>
      </div>
    </Modal>
  );
}
