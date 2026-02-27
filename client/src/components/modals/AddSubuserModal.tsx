import { useState, useRef, useEffect } from 'react';
import { addSubuser, type Subuser } from '../../lib/api';
import { notify, Modal, Input, Button } from '../';

interface Props {
  serverId: string;
  open: boolean;
  onClose: () => void;
  onAdded?: (subuser: Subuser) => void;
}

export default function AddSubuserModal({ serverId, open, onClose, onAdded }: Props) {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
      setEmail('');
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await addSubuser(serverId, email, []);
    if (res.success && res.data) {
      notify('Added', 'Subuser added successfully', 'success');
      onAdded?.(res.data);
      onClose();
    } else {
      notify('Error', res.error || 'Failed to add subuser', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Add Subuser" description="Enter the email of the user you want to add.">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input label="Email" type="email" placeholder="user@example.com" value={email} onChange={e => setEmail(e.target.value)} required />
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button type="submit" loading={loading}>Add</Button>
        </div>
      </form>
    </Modal>
  );
}
