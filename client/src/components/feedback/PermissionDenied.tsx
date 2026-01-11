import { Icons } from '../Icons';

export default function PermissionDenied({ message }: { message?: string }) {
  return (
    <div className="rounded-lg bg-red-500/10 border border-red-500/20 px-4 py-3 flex items-center gap-3">
      <Icons.errorCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
      <div>
        <p className="text-sm font-medium text-red-400">Access Denied</p>
        <p className="text-xs text-red-400/70">{message || "You don't have permission to access this"}</p>
      </div>
    </div>
  );
}
