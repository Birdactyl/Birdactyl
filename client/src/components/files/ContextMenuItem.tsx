import { ReactNode } from 'react';

interface ContextMenuItemProps {
  icon: ReactNode;
  label: string;
  onClick: () => void;
  disabled?: boolean;
  destructive?: boolean;
}

export default function ContextMenuItem({ icon, label, onClick, disabled, destructive }: ContextMenuItemProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`w-full inline-flex items-center justify-start gap-2 text-left px-2 py-1.5 rounded-md text-xs transition-colors disabled:opacity-50 ${
        destructive
          ? 'text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30'
          : 'text-neutral-700 hover:bg-neutral-100 dark:text-neutral-200 dark:hover:bg-neutral-800'
      }`}
      role="menuitem"
    >
      {icon}
      <span className="text-sm font-medium">{label}</span>
    </button>
  );
}
