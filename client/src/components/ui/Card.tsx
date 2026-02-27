import { ReactNode } from 'react';
import { Icons } from '../Icons';

interface CardProps {
  title?: string;
  description?: string;
  children: ReactNode;
  className?: string;
}

export function Card({ title, description, children, className = '' }: CardProps) {
  return (
    <div className={`rounded-xl bg-neutral-800/30 ${className}`}>
      {(title || description) && (
        <div className="px-6 pt-6 pb-3">
          {title && <h3 className="text-lg font-semibold text-neutral-100">{title}</h3>}
          {description && <p className="mt-1 text-sm text-neutral-400">{description}</p>}
        </div>
      )}
      <div className={title || description ? 'px-6 pb-6 pt-2' : 'p-6'}>{children}</div>
    </div>
  );
}

interface StatCardProps {
  label: string;
  value: string | number;
  max?: string;
  percent?: number;
  icon?: string;
}

export function StatCard({ label, value, max, percent, icon }: StatCardProps) {
  const IconComponent = icon ? (Icons as Record<string, React.ComponentType<{ className?: string }>>)[icon] : null;
  
  return (
    <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-800/30 shadow-xs cursor-default relative overflow-hidden">
      {percent !== undefined && (
        <div className="absolute inset-0 flex items-end opacity-10 dark:opacity-5 pointer-events-none">
          <div 
            className="bg-gradient-to-t from-sky-500 to-transparent w-full transition-all duration-500" 
            style={{ height: `${Math.min(100, percent)}%` }}
          />
        </div>
      )}
      
      <div className="p-4 flex items-start justify-between relative">
        <div>
          <div className="text-xs font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wide mb-1">
            {label}
          </div>
          <div className="text-xl font-bold text-neutral-900 dark:text-neutral-100 tabular-nums">
            {value}
          </div>
          {max && <div className="text-xs text-neutral-400 dark:text-neutral-500 mt-0.5">{max}</div>}
        </div>
        
        {IconComponent && (
          <div className="h-10 w-10 rounded-lg bg-neutral-100 dark:bg-neutral-800/50 grid place-items-center text-neutral-600 dark:text-neutral-400 shrink-0">
            <IconComponent className="h-5 w-5" />
          </div>
        )}
      </div>
    </div>
  );
}
