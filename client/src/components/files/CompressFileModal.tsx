import { useState, useEffect } from 'react';
import { useAsyncCallback } from '../../hooks/useAsync';
import Modal from '../ui/Modal';
import Button from '../ui/Button';
import { Icons } from '../Icons';
import { ContextMenu } from '../ui/ContextMenu';

interface Props { open: boolean; fileName: string; onClose: () => void; onCompress: (format: string) => Promise<void>; }

export default function CompressFileModal({ open, fileName, onClose, onCompress }: Props) {
  const [format, setFormat] = useState('zip');
  const [displayName, setDisplayName] = useState(fileName);
  const [handleCompress, loading] = useAsyncCallback(async () => { await onCompress(format); setFormat('zip'); }, [format, onCompress]);

  useEffect(() => { if (open && fileName) setDisplayName(fileName); }, [open, fileName]);
  const handleClose = () => { if (!loading) { setFormat('zip'); onClose(); } };

  return (
    <Modal open={open} onClose={handleClose} title="Compress" description={`Compress ${displayName}`}>
      <div className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-neutral-400 mb-1.5">Format</label>
          <ContextMenu
            align="start"
            className="w-full"
            trigger={
              <button className="w-full flex items-center justify-between rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 px-3 py-2 text-sm">
                <span>{format === 'tar.gz' ? '.tar.gz' : `.${format}`}</span>
                <Icons.chevronDown className="w-4 h-4" />
              </button>
            }
            items={[
              { label: '.zip', onClick: () => setFormat('zip') },
              { label: '.tar', onClick: () => setFormat('tar') },
              { label: '.tar.gz', onClick: () => setFormat('tar.gz') },
            ]}
          />
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="text" onClick={handleClose} disabled={loading}>Cancel</Button>
          <Button onClick={handleCompress} loading={loading}>Compress</Button>
        </div>
      </div>
    </Modal>
  );
}
