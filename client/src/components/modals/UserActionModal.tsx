import { useState, useRef, useEffect } from 'react';
import { adminBanUsers, adminUnbanUsers, adminDeleteUsers, adminSetAdmin, adminRevokeAdmin, adminForcePasswordReset } from '../../lib/api';
import { notify, Modal, Button } from '../';

type ActionType = 'ban' | 'unban' | 'delete' | 'setAdmin' | 'revokeAdmin' | 'forceReset';

interface Props {
  type: ActionType;
  ids: string[];
  open: boolean;
  onClose: () => void;
  onComplete: () => void;
}

const config: Record<ActionType, { title: string; description: string; fn: (ids: string[]) => Promise<{ success: boolean; data?: { affected: number }; error?: string }>; success: string; error: string }> = {
  ban: { title: 'Ban Users', description: 'They will be logged out immediately and unable to access their account.', fn: adminBanUsers, success: 'Users banned', error: 'Could not ban users' },
  unban: { title: 'Unban Users', description: 'They will be able to log in again.', fn: adminUnbanUsers, success: 'Users unbanned', error: 'Could not unban users' },
  delete: { title: 'Delete Users', description: 'This action cannot be undone. Users with servers cannot be deleted.', fn: adminDeleteUsers, success: 'Users deleted', error: 'Could not delete users' },
  setAdmin: { title: 'Grant Admin', description: 'They will have full administrative access.', fn: adminSetAdmin, success: 'Admin granted', error: 'Could not set admin' },
  revokeAdmin: { title: 'Revoke Admin', description: 'They will lose administrative privileges.', fn: adminRevokeAdmin, success: 'Admin revoked', error: 'Could not revoke admin' },
  forceReset: { title: 'Force Password Reset', description: 'They will be required to change their password on next login.', fn: adminForcePasswordReset, success: 'Password reset required', error: 'Could not force reset' },
};

export default function UserActionModal({ type, ids, open, onClose, onComplete }: Props) {
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
      notify(success, `${res.data?.affected} user(s) affected`, 'success');
      onComplete();
      onClose();
    } else {
      notify('Failed', res.error || error, 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title={title} description={`${ids.length} user(s) selected. ${description}`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button onClick={handleConfirm} loading={loading}>Confirm</Button>
      </div>
    </Modal>
  );
}
