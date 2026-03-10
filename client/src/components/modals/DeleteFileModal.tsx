import { useState, useRef, useEffect } from 'react';
import { Modal, Button } from '../';

interface Props {
  open: boolean;
  fileName: string;
  isDir: boolean;
  isBulk?: boolean;
  count?: number;
  isPermanent?: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}

export default function DeleteFileModal({ open, fileName, isDir, isBulk, count, isPermanent, onClose, onConfirm }: Props) {
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

  const titleAction = isPermanent ? 'Permanently Delete' : 'Delete';
  const title = isBulk ? `${titleAction} ${count} item${count !== 1 ? 's' : ''}` : `${titleAction} ${isDir ? 'folder' : 'file'}`;
  
  const descriptionPrefix = isBulk
    ? `Are you sure you want to delete ${count} selected item${count !== 1 ? 's' : ''}?`
    : `Are you sure you want to delete "${fileName}"?`;
    
  const descriptionSuffix = isPermanent 
    ? 'This action cannot be undone.'
    : 'This will be moved to the Recycle Bin (.trash).';

  const description = `${descriptionPrefix} ${descriptionSuffix}`;

  return (
    <Modal open={open} onClose={onClose} title={title} description={description}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleConfirm} loading={loading}>Delete</Button>
      </div>
    </Modal>
  );
}
