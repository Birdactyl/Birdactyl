import { useRef, useEffect, ReactNode } from 'react';

interface BulkActionBarProps {
  count: number;
  children: ReactNode;
  onClear: () => void;
}

export default function BulkActionBar({ count, children, onClear }: BulkActionBarProps) {
  const cachedCount = useRef(count);
  
  useEffect(() => {
    if (count > 0) cachedCount.current = count;
  }, [count]);

  return (
    <div className={`overflow-hidden transition-all duration-200 ${count > 0 ? 'max-h-20 opacity-100' : 'max-h-0 opacity-0'}`}>
      <div className="flex items-center gap-4 px-4 py-3 rounded-xl bg-neutral-800/50">
        <span className="text-sm text-neutral-300">{cachedCount.current} selected</span>
        <div className="h-4 w-px bg-neutral-700" />
        <div className="flex items-center gap-2">{children}</div>
        <button onClick={onClear} className="ml-auto text-xs text-neutral-500 hover:text-neutral-300 transition-colors">Clear selection</button>
      </div>
    </div>
  );
}
