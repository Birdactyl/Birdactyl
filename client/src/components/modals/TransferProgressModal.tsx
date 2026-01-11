import { type TransferStatus } from '../../lib/api';
import { Modal, Button, Icons } from '../';

const stageLabels: Record<TransferStatus['stage'], string> = {
  pending: 'Starting transfer...',
  stopping: 'Stopping server...',
  archiving: 'Creating archive...',
  downloading: 'Transferring...',
  uploading: 'Transferring to target...',
  importing: 'Finalizing...',
  cleanup: 'Cleaning up...',
  complete: 'Transfer complete!',
  failed: 'Transfer failed',
};

interface Props {
  transfer: TransferStatus;
  open: boolean;
  onClose: () => void;
}

export default function TransferProgressModal({ transfer, open, onClose }: Props) {
  const isDone = transfer.stage === 'complete' || transfer.stage === 'failed';

  return (
    <Modal open={open} onClose={onClose} title="Server Transfer" description={transfer.server_name}>
      <div className="space-y-4">
        <div className="flex items-center justify-between text-sm">
          <span className="text-neutral-400">{transfer.from_node_name}</span>
          <Icons.arrowRight className="w-4 h-4 text-neutral-500" />
          <span className="text-neutral-300">{transfer.to_node_name}</span>
        </div>
        <div className="space-y-2">
          <div className="flex items-center justify-between text-xs">
            <span className={transfer.stage === 'failed' ? 'text-red-400' : transfer.stage === 'complete' ? 'text-emerald-400' : 'text-neutral-300'}>
              {stageLabels[transfer.stage]}
            </span>
            <span className="text-neutral-500">{transfer.progress}%</span>
          </div>
          <div className="h-2 bg-neutral-800 rounded-full overflow-hidden">
            <div
              className={`h-full transition-all duration-300 ${transfer.stage === 'failed' ? 'bg-red-500' : transfer.stage === 'complete' ? 'bg-emerald-500' : 'bg-blue-500'}`}
              style={{ width: `${transfer.progress}%` }}
            />
          </div>
        </div>
        {transfer.error && <p className="text-xs text-red-400">{transfer.error}</p>}
        <div className="flex justify-end pt-2">
          <Button onClick={onClose}>{isDone ? 'Close' : 'Minimize'}</Button>
        </div>
      </div>
    </Modal>
  );
}
