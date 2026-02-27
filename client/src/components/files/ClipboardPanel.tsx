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
    <div className="fixed inset-x-0 bottom-0 z-[95] transition-all duration-200 ease-out animate-in slide-in-from-bottom-4 fade-in">
      <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
        <div className="rounded-t-lg border border-neutral-800 bg-neutral-900/95 shadow-2xl backdrop-blur px-4 py-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <span className="text-sm font-medium text-neutral-200">{items.length} in clipboard</span>
              <div className="h-4 w-px bg-neutral-700" />
              <div className="flex items-center gap-2">
                <Button onClick={onPaste} disabled={pasting} loading={pasting} variant="secondary">
                  <Icons.clipboardCheck className="h-4 w-4" />
                  Paste
                </Button>
                <Button variant="ghost" onClick={onClear} className="text-neutral-500 hover:text-neutral-300">
                  Clear
                </Button>
              </div>
            </div>
            <div className="flex items-center gap-1 max-w-[40%] overflow-x-auto scrollbar-hide">
              {items.slice(0, 3).map(path => (
                <span key={path} className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-neutral-800 text-xs text-neutral-400 truncate max-w-[120px]">
                  {path.split('/').pop()}
                  <button onClick={() => onRemove(path)} className="hover:text-neutral-200">
                    <Icons.x className="h-3 w-3" />
                  </button>
                </span>
              ))}
              {items.length > 3 && (
                <span className="text-xs text-neutral-500">+{items.length - 3} more</span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
