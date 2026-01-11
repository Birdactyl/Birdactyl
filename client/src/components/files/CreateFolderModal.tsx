import { useState } from 'react';
import { useAsyncCallback } from '../../hooks/useAsync';
import Modal from '../ui/Modal';
import Input from '../ui/Input';
import Button from '../ui/Button';

interface Props { open: boolean; onClose: () => void; onCreate: (name: string) => Promise<void>; }

export default function CreateFolderModal({ open, onClose, onCreate }: Props) {
  const [name, setName] = useState('');
  const [handleCreate, loading] = useAsyncCallback(async () => {
    if (!name.trim()) return;
    await onCreate(name);
    setName('');
  }, [name, onCreate]);

  const handleClose = () => { if (!loading) { setName(''); onClose(); } };

  return (
    <Modal open={open} onClose={handleClose} title="Create folder" description="Enter a name for the new folder">
      <div className="space-y-4">
        <Input label="Folder name" value={name} onChange={e => setName(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleCreate()} placeholder="e.g. logs" autoFocus disabled={loading} />
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="text" onClick={handleClose} disabled={loading}>Cancel</Button>
          <Button onClick={handleCreate} disabled={!name.trim()} loading={loading}>Create</Button>
        </div>
      </div>
    </Modal>
  );
}
