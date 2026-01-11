import { useState, useRef, useEffect } from 'react';
import { adminSuspendServers, adminUnsuspendServers, adminDeleteServers } from '../../lib/api';
import { notify, Modal, Button } from '../';

type ActionType = 'suspend' | 'unsuspend' | 'delete';

interface Props {
  type: ActionType;
  ids: string[];
  open: boolean;
  onClose: () => void;
  onComplete: () => void;
}

const config: Record<ActionType, { title: string; description: string; fn: typeof adminSuspendServers; success: string; error: string }> = {
  suspend: {
    title: 'Suspend Server',
    description: 'Suspended servers will be stopped and users cannot access them.',
    fn: adminSuspendServers,
    success: 'Servers suspended',
    error: 'Failed to suspend',
  },
  unsuspend: {
    title: 'Unsuspend Server',
    description: 'Users will be able to access these servers again.',
    fn: adminUnsuspendServers,
    success: 'Servers unsuspended',
    error: 'Failed to unsuspend',
  },
  delete: {
    title: 'Delete Server',
    description: 'This action cannot be undone. All server data will be permanently deleted.',
    fn: adminDeleteServers,
    success: 'Servers deleted',
    error: 'Failed to delete',
  },
};

export default function ConfirmActionModal({ type, ids, open, onClose, onComplete }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);
  const { title, description, fn, success, error } = config[type];

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
    const res = await fn(ids);
    if (res.success) {
      notify(success, `${res.data?.affected} server(s) affected`, 'success');
      onComplete();
      onClose();
    } else {
      notify('Failed', res.error || error, 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title={title} description={`${ids.length} server(s) selected. ${description}`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button onClick={handleConfirm} loading={loading}>Confirm</Button>
      </div>
    </Modal>
  );
}
