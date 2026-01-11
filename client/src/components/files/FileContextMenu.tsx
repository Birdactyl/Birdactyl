import { createPortal } from 'react-dom';
import { Icons } from '../Icons';
import ContextMenuItem from './ContextMenuItem';
import { FileEntry } from '../../lib/api';

interface FileContextMenuProps {
  file: FileEntry;
  position: { x: number; y: number; openUp: boolean };
  onEdit?: () => void;
  onDownload: () => void;
  onCopy: () => void;
  onDuplicate: () => void;
  onMove: () => void;
  onRename: () => void;
  onCompress: () => void;
  onExtract?: () => void;
  onDelete: () => void;
  isArchive?: boolean;
  extracting?: boolean;
}

export default function FileContextMenu({
  file,
  position,
  onEdit,
  onDownload,
  onCopy,
  onDuplicate,
  onMove,
  onRename,
  onCompress,
  onExtract,
  onDelete,
  isArchive,
  extracting,
}: FileContextMenuProps) {
  return createPortal(
    <div
      className="fixed z-[9999] w-44 rounded-md border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 shadow-xl p-1 animate-dropdown-in"
      style={{
        right: `calc(100vw - ${position.x}px)`,
        ...(position.openUp ? { bottom: window.innerHeight - position.y - 80 } : { top: position.y }),
      }}
      role="menu"
    >
      {file.is_dir ? (
        <>
          <ContextMenuItem icon={<Icons.move />} label="Move" onClick={onMove} />
          <ContextMenuItem icon={<Icons.rename />} label="Rename" onClick={onRename} />
          <ContextMenuItem icon={<Icons.compress />} label="Compress" onClick={onCompress} />
          <ContextMenuItem icon={<Icons.trash />} label="Delete" onClick={onDelete} destructive />
        </>
      ) : (
        <>
          {onEdit && <ContextMenuItem icon={<Icons.edit />} label="Edit" onClick={onEdit} />}
          <ContextMenuItem icon={<Icons.download />} label="Download" onClick={onDownload} />
          <ContextMenuItem icon={<Icons.copy />} label="Copy" onClick={onCopy} />
          <ContextMenuItem icon={<Icons.duplicate />} label="Duplicate" onClick={onDuplicate} />
          <ContextMenuItem icon={<Icons.move />} label="Move" onClick={onMove} />
          <ContextMenuItem icon={<Icons.rename />} label="Rename" onClick={onRename} />
          <ContextMenuItem icon={<Icons.compress />} label="Compress" onClick={onCompress} />
          {isArchive && onExtract && (
            <ContextMenuItem
              icon={<Icons.extract />}
              label={extracting ? 'Extracting...' : 'Extract'}
              onClick={onExtract}
              disabled={extracting}
            />
          )}
          <ContextMenuItem icon={<Icons.trash />} label="Delete" onClick={onDelete} destructive />
        </>
      )}
    </div>,
    document.body
  );
}
