import { useEffect, useState } from 'react';
import { startLoading, finishLoading } from '../../lib/pageLoader';
import { getResources, Resources } from '../../lib/api';
import { getUser } from '../../lib/auth';
import ServersPage from './ServersPage';

export default function DashboardPage() {
  const [resources, setResources] = useState<Resources | null>(null);
  const [ready, setReady] = useState(false);
  const user = getUser();

  useEffect(() => {
    startLoading();
    getResources().then(res => {
      if (res.success && res.data) setResources(res.data);
      setReady(true);
      finishLoading();
    });
  }, []);

  if (!ready) return null;

  return (
    <div className="space-y-8">
      <div className="rounded-xl bg-transparent border border-neutral-800">
        <div className="flex flex-col sm:flex-row sm:items-center gap-4 px-5 sm:px-8 py-6 sm:py-0 sm:h-48">
          <div className="flex flex-col gap-0 shrink-0">
            <div className="border mb-4 border-neutral-800 w-14 pl-1.5 rounded-lg h-[42px] pt-2 hidden sm:block">
              <svg width="45" height="24" viewBox="0 0 45 30" fill="none" xmlns="http://www.w3.org/2000/svg" style={{ transform: 'scaleX(-1)' }}>
                <path d="M0 7.50004L0 22.4913C0 26.6334 3.35786 29.9912 7.49999 29.9912C11.6421 29.9912 15 26.6334 15 22.4913L15 7.50004C15 3.35791 11.6421 5.53131e-05 7.49999 5.53131e-05C3.35786 5.53131e-05 0 3.35791 0 7.50004Z" fill="#525252" />
                <path d="M44.9998 15.0001C44.9998 23.283 38.2828 30 29.9998 30C21.7169 30 14.9999 23.283 14.9999 15.0001L44.9998 15.0001Z" fill="#a3a3a3" />
                <path d="M44.9998 15.0001C44.9998 6.71707 38.2828 8.27909e-05 29.9998 8.27909e-05C21.7169 8.27909e-05 14.9999 6.71707 14.9999 15.0001L44.9998 15.0001Z" fill="#d4d4d4" />
              </svg>
            </div>
            <span className="text-base sm:text-lg max-w-[300px] truncate font-semibold text-neutral-200 shrink-0">
              Welcome back, {user?.username || 'User'}
            </span>
            <span className="text-sm max-w-[600px] text-neutral-400">
              Manage your servers and infrastructure from here.
            </span>
          </div>

          {resources && resources.enabled && (
            <div className="sm:ml-auto grid grid-cols-2 sm:flex sm:items-center sm:divide-x sm:divide-neutral-800 gap-3 sm:gap-0">
              <div className="sm:px-4 sm:first:pl-0 sm:last:pr-0">
                <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Memory</div>
                <div className="text-sm font-medium text-neutral-300 tabular-nums">{resources.used.ram} / {resources.limits.ram} MiB</div>
              </div>
              <div className="sm:px-4 sm:first:pl-0 sm:last:pr-0">
                <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">CPU</div>
                <div className="text-sm font-medium text-neutral-300 tabular-nums">{resources.used.cpu}% / {resources.limits.cpu}%</div>
              </div>
              <div className="sm:px-4 sm:first:pl-0 sm:last:pr-0">
                <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Disk</div>
                <div className="text-sm font-medium text-neutral-300 tabular-nums">{resources.used.disk} / {resources.limits.disk} MiB</div>
              </div>
              <div className="sm:px-4 sm:first:pl-0 sm:last:pr-0">
                <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">Servers</div>
                <div className="text-sm font-medium text-neutral-300 tabular-nums">{resources.used.servers} / {resources.limits.servers}</div>
              </div>
            </div>
          )}
        </div>
      </div>

      <ServersPage />
    </div>
  );
}
