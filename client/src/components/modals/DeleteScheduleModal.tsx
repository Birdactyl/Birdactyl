import { useState, useRef, useEffect } from 'react';
import { deleteSchedule, type Schedule } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
    serverId: string;
    schedule: Schedule | null;
    open: boolean;
    onClose: () => void;
    onDeleted?: (scheduleId: string) => void;
}

export default function DeleteScheduleModal({ serverId, schedule, open, onClose, onDeleted }: Props) {
    const [loading, setLoading] = useState(false);
    const submittingRef = useRef(false);

    useEffect(() => {
        if (open) {
            submittingRef.current = false;
            setLoading(false);
        }
    }, [open]);

    const handleDelete = async () => {
        if (!schedule || submittingRef.current) return;
        submittingRef.current = true;
        setLoading(true);
        const res = await deleteSchedule(serverId, schedule.id);
        if (res.success) {
            notify('Deleted', 'Schedule deleted', 'success');
            onDeleted?.(schedule.id);
            onClose();
        } else {
            notify('Error', res.error || 'Failed to delete', 'error');
            setLoading(false);
            submittingRef.current = false;
        }
    };

    return (
        <Modal open={open} onClose={onClose} title="Delete Schedule" description={`Are you sure you want to delete "${schedule?.name}"? This action cannot be undone.`}>
            <div className="flex justify-end gap-3 pt-4">
                <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
                <Button variant="danger" onClick={handleDelete} loading={loading}>Delete</Button>
            </div>
        </Modal>
    );
}
