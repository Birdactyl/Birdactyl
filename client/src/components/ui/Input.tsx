import { useState, useRef, useCallback } from 'react';
import { Icons } from '../Icons';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  disableAutofill?: boolean;
  hideable?: boolean;
}

export default function Input({ label, disableAutofill = false, hideable = false, className = '', type, ...props }: InputProps) {
  const [isHidden, setIsHidden] = useState(hideable);
  const inputRef = useRef<HTMLInputElement>(null);
  const intervalRef = useRef<number | null>(null);
  const timeoutRef = useRef<number | null>(null);
  const isNumber = type === 'number';

  const step = useCallback((dir: 1 | -1) => {
    if (!inputRef.current) return;
    const stepVal = Number(props.step) || 1;
    const min = props.min !== undefined ? Number(props.min) : -Infinity;
    const max = props.max !== undefined ? Number(props.max) : Infinity;
    const current = Number(inputRef.current.value) || 0;
    const next = Math.min(max, Math.max(min, current + stepVal * dir));
    inputRef.current.value = String(next);
    inputRef.current.dispatchEvent(new Event('input', { bubbles: true }));
  }, [props.step, props.min, props.max]);

  const startHold = (dir: 1 | -1) => {
    step(dir);
    timeoutRef.current = window.setTimeout(() => {
      intervalRef.current = window.setInterval(() => step(dir), 50);
    }, 400);
  };

  const stopHold = () => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    if (intervalRef.current) clearInterval(intervalRef.current);
    timeoutRef.current = null;
    intervalRef.current = null;
  };

  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="block text-xs font-medium text-neutral-400">
          <span className="inline-flex items-center gap-1.5">
            {label}
            {props.required && <span className="text-red-600" aria-hidden="true">*</span>}
          </span>
        </label>
      )}
      <div className="relative">
        <input
          ref={inputRef}
          className={`w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 placeholder:text-neutral-500 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-[#0a0a0a] focus:border-neutral-500 disabled:opacity-60 disabled:cursor-not-allowed px-3 py-2 ${hideable || isNumber ? 'pr-10' : ''} ${isNumber ? '[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none' : ''} ${className}`}
          autoComplete={disableAutofill ? 'one-time-code' : props.autoComplete}
          type={hideable ? (isHidden ? 'password' : 'text') : type}
          style={{ fontSize: '13px' }}
          {...props}
        />
        {hideable && (
          <button
            type="button"
            onClick={() => setIsHidden(!isHidden)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-200 transition-colors"
          >
            {isHidden ? <Icons.eyeOff className="w-5 h-5" /> : <Icons.eye className="w-5 h-5" />}
          </button>
        )}
        {isNumber && !props.disabled && (
          <div className="absolute right-1 top-1/2 -translate-y-1/2 flex flex-col">
            <button type="button" onMouseDown={() => startHold(1)} onMouseUp={stopHold} onMouseLeave={stopHold} className="px-1.5 py-0.5 text-neutral-500 hover:text-neutral-200 transition-colors">
              <Icons.chevronUp className="w-3 h-3" />
            </button>
            <button type="button" onMouseDown={() => startHold(-1)} onMouseUp={stopHold} onMouseLeave={stopHold} className="px-1.5 py-0.5 text-neutral-500 hover:text-neutral-200 transition-colors">
              <Icons.chevronDown className="w-3 h-3" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
