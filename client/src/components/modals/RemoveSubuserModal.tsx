import { useState, useRef, useEffect } from 'react';
import { removeSubuser } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
  serverId: string;
  subuserId: string;
  username: string;
  open: boolean;
  onClose: () => void;
  onRemoved?: () => void;
}

export default function RemoveSubuserModal({ serverId, subuserId, username, open, onClose, onRemoved }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
    }
  }, [open]);

  const handleRemove = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await removeSubuser(serverId, subuserId);
    if (res.success) {
      notify('Removed', 'Subuser removed', 'success');
      onRemoved?.();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to remove subuser', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Remove Subuser" description={`Remove ${username} from this server? They will lose all access.`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleRemove} loading={loading}>Remove</Button>
      </div>
    </Modal>
  );
}
