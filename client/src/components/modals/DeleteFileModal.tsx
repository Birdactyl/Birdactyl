import { useState, useRef, useEffect } from 'react';
import { Modal, Button } from '../';

interface Props {
  open: boolean;
  fileName: string;
  isDir: boolean;
  isBulk?: boolean;
  count?: number;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}

export default function DeleteFileModal({ open, fileName, isDir, isBulk, count, onClose, onConfirm }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
    }
  }, [open]);

  const handleConfirm = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    await onConfirm();
    onClose();
  };

  const title = isBulk ? `Delete ${count} item${count !== 1 ? 's' : ''}` : `Delete ${isDir ? 'folder' : 'file'}`;
  const description = isBulk
    ? `Are you sure you want to delete ${count} selected item${count !== 1 ? 's' : ''}? This action cannot be undone.`
    : `Are you sure you want to delete "${fileName}"? This action cannot be undone.`;

  return (
    <Modal open={open} onClose={onClose} title={title} description={description}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleConfirm} loading={loading}>Delete</Button>
      </div>
    </Modal>
  );
}
