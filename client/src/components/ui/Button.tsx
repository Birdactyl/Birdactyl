import { ReactNode } from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  loading?: boolean;
  variant?: 'primary' | 'secondary' | 'ghost' | 'text' | 'danger';
}

const variants = {
  primary: 'border border-transparent bg-black dark:bg-white text-white dark:text-black hover:bg-neutral-800 dark:hover:bg-neutral-200',
  secondary: 'border border-transparent bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 hover:bg-neutral-200 dark:hover:bg-neutral-700',
  ghost: 'border border-transparent bg-transparent text-neutral-500 dark:text-neutral-400 hover:bg-neutral-100/80 dark:hover:bg-neutral-800/80 hover:text-neutral-900 dark:hover:text-neutral-100',
  text: 'border border-transparent bg-transparent text-neutral-400 hover:text-neutral-100',
  danger: 'border border-transparent bg-red-500 dark:bg-red-600 text-white hover:bg-red-600 dark:hover:bg-red-700',
};

export default function Button({ children, loading = false, variant = 'primary', className = '', disabled, ...props }: ButtonProps) {
  return (
    <button
      className={`transition-all rounded-lg cursor-pointer inline-flex items-center justify-center whitespace-nowrap font-semibold disabled:opacity-60 disabled:cursor-not-allowed disabled:pointer-events-none text-xs px-3 py-1.5 ${variants[variant]} ${loading ? 'pointer-events-none' : ''} ${className}`}
      aria-busy={loading}
      disabled={disabled || loading}
      {...props}
    >
      <span className="relative inline-flex items-center justify-center">
        <span
          className={`flex items-center justify-center gap-2 transition duration-200 ease-out will-change-transform ${
            loading ? 'opacity-0 blur-[2px] scale-95' : 'opacity-100 blur-0 scale-100'
          }`}
          aria-hidden={loading}
        >
          <span className="font-semibold text-sm inline-flex items-center">{children}</span>
        </span>
        {loading && (
          <span
            className="absolute inset-0 flex items-center justify-center transition duration-200 ease-out opacity-100 blur-0 scale-100"
            role="status"
            aria-live="polite"
          >
            <span className="inline-block rounded-full border-2 border-current border-t-transparent animate-spin h-4 w-4"></span>
            <span className="sr-only">Loading...</span>
          </span>
        )}
      </span>
    </button>
  );
}
