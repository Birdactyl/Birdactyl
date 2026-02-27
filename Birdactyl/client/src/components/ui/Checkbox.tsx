import { Icons } from '../Icons';

interface CheckboxProps {
  checked: boolean;
  onChange: (checked?: boolean) => void;
  indeterminate?: boolean;
  label?: string;
}

export default function Checkbox({ checked, onChange, indeterminate, label }: CheckboxProps) {
  const box = (
    <button
      type="button"
      onClick={() => onChange(!checked)}
      className={`h-4 w-4 rounded border transition-colors flex items-center justify-center flex-shrink-0 ${
        checked || indeterminate
          ? 'bg-neutral-100 border-neutral-100'
          : 'bg-transparent border-neutral-600 hover:border-neutral-400'
      }`}
    >
      {checked && <Icons.checkSmall className="h-3 w-3 text-neutral-900" />}
      {indeterminate && !checked && <Icons.minusSmall className="h-3 w-3 text-neutral-900" />}
    </button>
  );

  if (label) {
    return (
      <label className="flex items-center gap-2 cursor-pointer">
        {box}
        <span className="text-sm text-neutral-300">{label}</span>
      </label>
    );
  }

  return box;
}
