import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { deleteServer } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
  serverId: string;
  serverName: string;
  open: boolean;
  onClose: () => void;
  onDeleted?: () => void;
  redirectOnDelete?: boolean;
}

export default function DeleteServerModal({ serverId, serverName, open, onClose, onDeleted, redirectOnDelete }: Props) {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
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
    const res = await deleteServer(serverId);
    if (res.success) {
      notify('Server Deleted', serverName, 'success');
      onDeleted?.();
      onClose();
      if (redirectOnDelete) navigate('/console/servers', { replace: true });
    } else {
      notify('Error', res.error || 'Failed to delete', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Delete Server" description={`Are you sure you want to delete "${serverName}"? This action cannot be undone.`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
      </div>
    </Modal>
  );
}
