import { ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Icons } from '../Icons';

interface BulkActionBarProps {
  count: number;
  children: ReactNode;
  onClear: () => void;
}

export default function BulkActionBar({ count, children, onClear }: BulkActionBarProps) {
  if (count <= 0) return null;

  return createPortal(
    <div className="fixed inset-x-0 bottom-0 z-[95] transition-all duration-200 ease-out animate-in slide-in-from-bottom-4 fade-in">
      <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
        <div className="rounded-t-lg border border-neutral-800 bg-neutral-900/95 shadow-2xl backdrop-blur px-4 py-3">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <span className="text-sm font-medium text-neutral-200">{count} selected</span>
              <div className="h-4 w-px bg-neutral-700" />
              <div className="flex items-center gap-2">{children}</div>
            </div>
            <button
              onClick={onClear}
              className="flex items-center gap-1.5 text-xs text-neutral-500 hover:text-neutral-300 transition-colors"
            >
              <Icons.x className="h-3.5 w-3.5" />
              Clear
            </button>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
