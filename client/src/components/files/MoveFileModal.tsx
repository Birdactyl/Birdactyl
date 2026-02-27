import { useState, useEffect } from 'react';
import { useAsyncCallback } from '../../hooks/useAsync';
import Modal from '../ui/Modal';
import Input from '../ui/Input';
import Button from '../ui/Button';

interface Props { open: boolean; initialPath: string; onClose: () => void; onMove: (path: string) => Promise<void>; }

export default function MoveFileModal({ open, initialPath, onClose, onMove }: Props) {
  const [path, setPath] = useState(initialPath);
  const [handleMove, loading] = useAsyncCallback(async () => { if (path.trim()) await onMove(path); }, [path, onMove]);

  useEffect(() => { if (open) setPath(initialPath); }, [open, initialPath]);
  const handleClose = () => { if (!loading) onClose(); };

  return (
    <Modal open={open} onClose={handleClose} title="Move" description="Enter a full path like /folder/newname">
      <div className="space-y-4">
        <Input label="Destination path" value={path} onChange={e => setPath(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleMove()} placeholder="e.g. /backups/file.txt" autoFocus disabled={loading} />
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="text" onClick={handleClose} disabled={loading}>Cancel</Button>
          <Button onClick={handleMove} disabled={!path.trim()} loading={loading}>Move</Button>
        </div>
      </div>
    </Modal>
  );
}
