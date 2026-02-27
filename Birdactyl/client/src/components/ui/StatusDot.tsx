type Status = 'running' | 'stopped' | 'installing' | 'suspended' | 'failed' | string;

const colors: Record<string, string> = {
    running: 'bg-emerald-500',
    stopped: 'bg-neutral-500',
    installing: 'bg-yellow-500',
    suspended: 'bg-amber-500',
    failed: 'bg-red-500',
};

export function StatusDot({ status, className = '' }: { status: Status; className?: string }) {
    const color = colors[status] || 'bg-neutral-500';
    return (
        <span className={`inline-block h-2 w-2 rounded-full ${color} shrink-0 ${className}`} />
    );
}
