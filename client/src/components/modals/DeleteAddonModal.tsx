import { useState, useRef, useEffect } from 'react';
import { deleteAddon } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
  serverId: string;
  sourceId: string;
  fileName: string;
  open: boolean;
  onClose: () => void;
  onDeleted?: () => void;
}

export default function DeleteAddonModal({ serverId, sourceId, fileName, open, onClose, onDeleted }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
    }
  }, [open]);

  const handleDelete = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await deleteAddon(serverId, sourceId, fileName);
    if (res.success) {
      notify('Deleted', `${fileName} has been removed`, 'success');
      onDeleted?.();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to delete', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Delete addon" description={`Are you sure you want to delete "${fileName}"? This cannot be undone.`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
      </div>
    </Modal>
  );
}
