import { Icons } from '../Icons';

type Variant = 'start' | 'restart' | 'stop' | 'kill';

const config: Record<Variant, { icon: keyof typeof Icons; label: string; colors: string }> = {
  start: { icon: 'play', label: 'Start', colors: 'bg-emerald-700 enabled:hover:bg-emerald-800 text-emerald-50 dark:bg-emerald-800/20 dark:enabled:hover:bg-emerald-800/30 dark:text-emerald-300 focus:ring-emerald-600/40 active:ring-emerald-600/40' },
  restart: { icon: 'refresh', label: 'Restart', colors: 'bg-transparent enabled:hover:bg-neutral-200/50 shadow-xs border border-neutral-200 dark:border-transparent text-neutral-900 dark:bg-neutral-800/20 dark:enabled:hover:bg-neutral-800/30 dark:text-neutral-300 focus:ring-neutral-600/40 active:ring-neutral-600/40' },
  stop: { icon: 'stopFilled', label: 'Stop', colors: 'bg-red-700 enabled:hover:bg-red-800 text-red-50 dark:bg-red-800/20 dark:enabled:hover:bg-red-800/30 dark:text-red-300 focus:ring-red-600/40 active:ring-red-600/40 border-0' },
  kill: { icon: 'stopFilled', label: 'Kill', colors: 'bg-orange-700 enabled:hover:bg-orange-800 text-orange-50 dark:bg-orange-800/20 dark:enabled:hover:bg-orange-800/30 dark:text-orange-300 focus:ring-orange-600/40 active:ring-orange-600/40 border-0' },
};

export function PowerButton({ variant, onClick, disabled }: { variant: Variant; onClick: () => void; disabled: boolean }) {
  const { icon, label, colors } = config[variant];
  const Icon = Icons[icon];
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`transition-all rounded-lg cursor-pointer inline-flex items-center justify-center whitespace-nowrap outline-none font-semibold text-xs px-3.5 py-1.5 ring-0 focus:ring-2 focus:ring-offset-2 active:ring-2 active:ring-offset-2 ring-offset-white dark:ring-offset-neutral-950 disabled:opacity-60 disabled:cursor-not-allowed ${colors}`}
    >
      <Icon className="w-4 h-4 mr-2" />
      <span className="font-semibold text-sm">{label}</span>
    </button>
  );
}
