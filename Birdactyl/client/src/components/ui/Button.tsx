import React, { ReactNode } from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  loading?: boolean;
  leftIcon?: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'ghost' | 'text' | 'danger';
}

const variants = {
  primary: "bg-black dark:bg-white text-white dark:text-black hover:bg-neutral-800 dark:hover:bg-neutral-200 ring-1 ring-neutral-900 dark:ring-neutral-100",
  secondary: "bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 hover:bg-neutral-200 dark:hover:bg-neutral-700 ring-1 ring-neutral-200 dark:ring-neutral-700",
  ghost: "bg-transparent text-neutral-500 dark:text-neutral-400 hover:bg-neutral-100/80 dark:hover:bg-neutral-800/80 hover:text-neutral-900 dark:hover:text-neutral-100",
  text: "bg-transparent text-neutral-400 hover:text-neutral-100",
  danger: "bg-red-500 dark:bg-red-600 text-white hover:bg-red-600 dark:hover:bg-red-700 ring-1 ring-red-600 dark:ring-red-700",
};

export default function Button({ children, loading = false, leftIcon, variant = 'primary', className = '', disabled, ...props }: ButtonProps) {
  return (
    <button
      className={`
        transition-all focus:outline-none focus:ring-2 focus:ring-offset-2
        focus:ring-neutral-400 dark:focus:ring-neutral-600
        ring-offset-[#0a0a0a]
        rounded-lg cursor-pointer tracking-tight inline-flex
        items-center justify-center whitespace-nowrap gap-1.5
        font-semibold text-xs px-3 py-1.5 shrink-0
        disabled:opacity-60 disabled:cursor-not-allowed disabled:pointer-events-none
        ${variants[variant]}
        ${loading ? 'pointer-events-none' : ''}
        ${className}
      `}
      aria-busy={loading}
      disabled={disabled || loading}
      {...props}
    >
      <span className="relative inline-flex items-center justify-center">
        <span
          className={`
            flex items-center justify-center gap-2 transition duration-200 ease-out will-change-transform
            ${loading ? 'opacity-0 blur-[2px] scale-95 pointer-events-none' : 'opacity-100 blur-0 scale-100'}
          `}
          aria-hidden={loading}
        >
          <span className="font-semibold text-sm inline-flex items-center">
            {leftIcon && <span className="w-4 h-4 mr-1.5 flex items-center">{leftIcon}</span>}
            {children}
          </span>
        </span>

        {loading && (
          <span className="absolute inset-0 flex items-center justify-center transition duration-200 ease-out" role="status" aria-live="polite">
            <span className="inline-flex shrink-0 h-4 w-4">
              <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" className="size-full">
                <g fill="currentColor" opacity="0.5" className="motion-safe:animate-spin motion-safe:origin-center" style={{ animationDuration: '12s', animationDirection: 'reverse' }}>
                  <circle cx="8" cy="2.75" r="0.75"></circle>
                  <circle cx="13.25" cy="8" r="0.75"></circle>
                  <circle cx="2.75" cy="8" r="0.75"></circle>
                  <circle cx="4.29999" cy="4.29001" r="0.75"></circle>
                  <circle cx="11.7" cy="4.29001" r="0.75"></circle>
                  <circle cx="4.29999" cy="11.72" r="0.75"></circle>
                  <circle cx="11.7" cy="11.72" r="0.75"></circle>
                  <circle cx="8" cy="13.25" r="0.75"></circle>
                </g>
                <circle className="motion-safe:animate-spin motion-safe:origin-center" cx="8" cy="8" r="5.25" pathLength="360" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeDashoffset="100" strokeDasharray="90 270" strokeWidth="1.5" style={{ animationDuration: '1.4s' }}></circle>
              </svg>
            </span>
          </span>
        )}
      </span>
    </button>
  );
}
