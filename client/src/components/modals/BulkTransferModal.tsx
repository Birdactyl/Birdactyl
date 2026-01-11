import { useState, useRef, useEffect } from 'react';
import { adminTransferServer, type Node } from '../../lib/api';
import { notify, Modal, Button, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../';

interface Props {
  serverIds: string[];
  nodes: Node[];
  getServerNodeId: (id: string) => string | undefined;
  open: boolean;
  onClose: () => void;
  onComplete: () => void;
}

export default function BulkTransferModal({ serverIds, nodes, getServerNodeId, open, onClose, onComplete }: Props) {
  const [targetNodeId, setTargetNodeId] = useState('');
  const [concurrency, setConcurrency] = useState(1);
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState({ current: 0, total: serverIds.length, percent: 0 });
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
      setTargetNodeId('');
      setProgress({ current: 0, total: serverIds.length, percent: 0 });
    }
  }, [open, serverIds.length]);

  const handleTransfer = async () => {
    if (!targetNodeId || submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);

    const ids = serverIds.filter(id => getServerNodeId(id) !== targetNodeId);
    let started = 0;
    let failed = 0;

    const startOne = async (id: string) => {
      const res = await adminTransferServer(id, targetNodeId);
      if (!res.success) failed++;
      started++;
      setProgress({ current: started, total: ids.length, percent: Math.round((started / ids.length) * 100) });
    };

    for (let i = 0; i < ids.length; i += concurrency) {
      const batch = ids.slice(i, i + concurrency);
      await Promise.all(batch.map(startOne));
    }

    notify('Bulk Transfer Started', `${started - failed} transfers started, ${failed} failed to start`, failed > 0 ? 'error' : 'success');
    onComplete();
    onClose();
  };

  const onlineNodes = nodes.filter(n => n.is_online);

  return (
    <Modal open={open} onClose={onClose} title="Bulk Transfer" description={`${serverIds.length} server(s) selected`}>
      <div className="space-y-4">
        {!loading ? (
          <>
            <div className="flex flex-col gap-1.5">
              <label className="block text-xs font-medium text-neutral-400">Target Node</label>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                    <span className="truncate text-neutral-100">
                      {targetNodeId ? nodes.find(n => n.id === targetNodeId)?.name : 'Select a node'}
                    </span>
                    <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  {onlineNodes.map(node => (
                    <DropdownMenuItem key={node.id} onSelect={() => setTargetNodeId(node.id)}>
                      {node.name} ({node.fqdn})
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="block text-xs font-medium text-neutral-400">Concurrent Transfers</label>
              <div className="flex gap-2">
                {[1, 5, 10, 20, 50].map(n => (
                  <button key={n} type="button" onClick={() => setConcurrency(n)} className={`px-3 py-1.5 text-xs rounded-lg border transition-colors ${concurrency === n ? 'border-blue-500 bg-blue-500/20 text-blue-400' : 'border-neutral-700 text-neutral-400 hover:border-neutral-500'}`}>{n}</button>
                ))}
              </div>
              <p className="text-xs text-neutral-500">Higher values = faster but more load on nodes</p>
            </div>
            <div className="flex justify-end gap-3 pt-4 border-t border-neutral-800">
              <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
              <Button onClick={handleTransfer} disabled={!targetNodeId || loading}>Transfer All</Button>
            </div>
          </>
        ) : (
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-neutral-300">Transferring {progress.current} of {progress.total}</span>
              <span className="text-neutral-500">{progress.percent}%</span>
            </div>
            <div className="h-2 bg-neutral-800 rounded-full overflow-hidden">
              <div className="h-full bg-blue-500 transition-all duration-300" style={{ width: `${progress.percent}%` }} />
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
}
