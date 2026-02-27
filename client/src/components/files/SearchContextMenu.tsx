import { createPortal } from 'react-dom';
import { Icons } from '../Icons';
import ContextMenuItem from './ContextMenuItem';

interface Props {
  position: { x: number; y: number; openUp: boolean };
  onOpenDirectory: () => void;
}

export default function SearchContextMenu({ position, onOpenDirectory }: Props) {
  return createPortal(
    <div
      className="fixed z-[9999] w-44 rounded-md border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 shadow-xl p-1 animate-dropdown-in"
      style={{
        right: `calc(100vw - ${position.x}px)`,
        ...(position.openUp ? { bottom: window.innerHeight - position.y - 80 } : { top: position.y }),
      }}
      role="menu"
    >
      <ContextMenuItem icon={<Icons.folderOpen />} label="Open Directory" onClick={onOpenDirectory} />
    </div>,
    document.body
  );
}
