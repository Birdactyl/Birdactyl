import { useState, useRef, useEffect } from 'react';
import { reinstallServer } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
  serverId: string;
  serverName: string;
  open: boolean;
  onClose: () => void;
  onReinstalled?: () => void;
}

export default function ReinstallServerModal({ serverId, serverName, open, onClose, onReinstalled }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
    }
  }, [open]);

  const handleReinstall = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await reinstallServer(serverId);
    if (res.success) {
      notify('Reinstalling', `${serverName} is being reinstalled`, 'success');
      onReinstalled?.();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to reinstall', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Reinstall Server" description={`Are you sure you want to reinstall "${serverName}"? Some files may be deleted or modified.`}>
      <div className="flex justify-end gap-3 pt-4">
        <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
        <Button variant="danger" onClick={handleReinstall} loading={loading}>Reinstall</Button>
      </div>
    </Modal>
  );
}
