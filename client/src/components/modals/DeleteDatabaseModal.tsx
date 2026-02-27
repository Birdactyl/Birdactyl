import { useState, useRef, useEffect } from 'react';
import { deleteServerDatabase, ServerDatabase } from '../../lib/api';
import { notify } from '../feedback/Notification';
import { Modal, Button } from '../';

interface Props {
  serverId: string;
  database: ServerDatabase | null;
  open: boolean;
  onClose: () => void;
  onDeleted?: (dbId: string) => void;
}

export default function DeleteDatabaseModal({ serverId, database, open, onClose, onDeleted }: Props) {
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
    }
  }, [open]);

  const handleDelete = async () => {
    if (!database || submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await deleteServerDatabase(serverId, database.id);
    if (res.success) {
      notify('Success', 'Database deleted', 'success');
      onDeleted?.(database.id);
      onClose();
    } else {
      notify('Error', res.error || 'Failed to delete database', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Delete Database" description="Are you sure you want to delete this database? This action cannot be undone.">
      <div className="space-y-4">
        {database && <p className="text-sm text-neutral-300">Database: <code className="text-violet-400">{database.database_name}</code></p>}
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
        </div>
      </div>
    </Modal>
  );
}
