import { useState, useEffect } from 'react';
import { useAsyncCallback } from '../../hooks/useAsync';
import Modal from '../ui/Modal';
import Input from '../ui/Input';
import Button from '../ui/Button';

interface Props { open: boolean; initialName: string; isDir: boolean; onClose: () => void; onRename: (name: string) => Promise<void>; }

export default function RenameFileModal({ open, initialName, isDir, onClose, onRename }: Props) {
  const [name, setName] = useState(initialName);
  const [handleRename, loading] = useAsyncCallback(async () => { if (name.trim()) await onRename(name); }, [name, onRename]);

  useEffect(() => { if (open) setName(initialName); }, [open, initialName]);
  const handleClose = () => { if (!loading) onClose(); };

  return (
    <Modal open={open} onClose={handleClose} title="Rename" description={`Rename ${isDir ? 'folder' : 'file'}`}>
      <div className="space-y-4">
        <Input label="New name" value={name} onChange={e => setName(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleRename()} placeholder="e.g. newname.txt" autoFocus disabled={loading} />
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="text" onClick={handleClose} disabled={loading}>Cancel</Button>
          <Button onClick={handleRename} disabled={!name.trim()} loading={loading}>Rename</Button>
        </div>
      </div>
    </Modal>
  );
}
