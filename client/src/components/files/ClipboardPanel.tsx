import { createPortal } from 'react-dom';
import { Icons } from '../Icons';
import Button from '../ui/Button';

interface Props {
  items: string[];
  pasting: boolean;
  onPaste: () => void;
  onClear: () => void;
  onRemove: (path: string) => void;
}

export default function ClipboardPanel({ items, pasting, onPaste, onClear, onRemove }: Props) {
  if (items.length === 0) return null;

  return createPortal(
    <div className="fixed right-4 bottom-4 z-[101] transition-all duration-150 ease-out opacity-100 translate-y-0">
      <div className="w-[320px] max-h-[40vh] rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white/95 dark:bg-neutral-900/95 shadow-xl backdrop-blur flex flex-col">
        <div className="px-3 py-2 border-b border-neutral-200 dark:border-neutral-800 flex items-center justify-between">
          <div className="text-xs font-medium text-neutral-800 dark:text-neutral-100">
            Clipboard <span className="text-neutral-500 dark:text-neutral-400 font-normal">({items.length})</span>
          </div>
          <div className="flex items-center gap-1">
            <Button onClick={onPaste} disabled={pasting} loading={pasting}>
              <Icons.clipboardCheck className="h-4 w-4 mr-1" />
              Paste here
            </Button>
            <Button variant="ghost" onClick={onClear} className="px-2 py-1">Clear</Button>
          </div>
        </div>
        <div className="px-2 py-2 overflow-auto">
          <ul className="space-y-1">
            {items.map(path => (
              <li key={path} className="flex items-center justify-between gap-2 px-2 py-1 rounded-md hover:bg-neutral-100 dark:hover:bg-neutral-800 transition-colors">
                <div className="text-[12px] text-neutral-700 dark:text-neutral-300 truncate" title={path}>
                  {path.split('/').pop()}
                </div>
                <button onClick={() => onRemove(path)} className="text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100 p-1" aria-label="Remove">
                  <Icons.x className="h-4 w-4" />
                </button>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>,
    document.body
  );
}
