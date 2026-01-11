import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { startLoading, finishLoading } from '../../lib/pageLoader';
import { Icons } from '../../components';
import { getResources, Resources, getServers, Server } from '../../lib/api';
import { getUser } from '../../lib/auth';

function ResourceRing({ percent, color, size = 80 }: { percent: number; color: string; size?: number }) {
  const strokeWidth = 6;
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (Math.min(percent, 100) / 100) * circumference;

  return (
    <svg width={size} height={size} className="transform -rotate-90">
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        className="text-neutral-800"
      />
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeDasharray={circumference}
        strokeDashoffset={offset}
        strokeLinecap="round"
        className="transition-all duration-500"
      />
    </svg>
  );
}

function QuickServer({ server }: { server: Server }) {
  const isOnline = server.status === 'running';
  
  return (
    <Link
      to={`/console/server/${server.id}`}
      className="flex items-center gap-3 p-3 rounded-lg bg-neutral-900/50 hover:bg-neutral-800/50 transition-colors group"
    >
      <div className={`w-2 h-2 rounded-full ${isOnline ? 'bg-emerald-500' : 'bg-neutral-500'}`} />
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-neutral-100 truncate">{server.name}</div>
        <div className="text-xs text-neutral-500">{server.node?.name || 'Unknown node'}</div>
      </div>
      <Icons.chevronRight className="w-4 h-4 text-neutral-600 group-hover:text-neutral-400 transition-colors" />
    </Link>
  );
}

export default function DashboardPage() {
  const [ready, setReady] = useState(false);
  const [resources, setResources] = useState<Resources | null>(null);
  const [servers, setServers] = useState<Server[]>([]);
  const user = getUser();

  useEffect(() => {
    startLoading();
    Promise.all([
      getResources(),
      getServers()
    ]).then(([resourcesRes, serversRes]) => {
      if (resourcesRes.success && resourcesRes.data) setResources(resourcesRes.data);
      if (serversRes.success && serversRes.data) setServers(serversRes.data.slice(0, 5));
      setReady(true);
      finishLoading();
    });
  }, []);

  if (!ready) return null;

  const memPercent = resources ? Math.round((resources.used.ram / resources.limits.ram) * 100) : 0;
  const cpuPercent = resources ? Math.round((resources.used.cpu / resources.limits.cpu) * 100) : 0;
  const diskPercent = resources ? Math.round((resources.used.disk / resources.limits.disk) * 100) : 0;
  const serverPercent = resources ? Math.round((resources.used.servers / resources.limits.servers) * 100) : 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-100">Welcome back, {user?.username || 'User'}</h1>
        <p className="text-sm text-neutral-400 mt-1">Here's an overview of your resources and servers.</p>
      </div>

      <div className="grid gap-6">
        <div className="rounded-xl bg-neutral-800/30 p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="text-sm font-medium text-neutral-300">Resource Usage</div>
            <div className="text-xs text-neutral-500">{resources?.used.servers ?? 0} of {resources?.limits.servers ?? 0} servers</div>
          </div>
          
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
            <div className="flex flex-col items-center">
              <div className="relative">
                <ResourceRing percent={memPercent} color="#38bdf8" />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-lg font-semibold text-neutral-100">{memPercent}%</span>
                </div>
              </div>
              <div className="mt-3 text-center">
                <div className="text-xs font-medium text-neutral-300">Memory</div>
                <div className="text-xs text-neutral-500">{resources?.used.ram ?? 0} / {resources?.limits.ram ?? 0} MB</div>
              </div>
            </div>

            <div className="flex flex-col items-center">
              <div className="relative">
                <ResourceRing percent={cpuPercent} color="#a78bfa" />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-lg font-semibold text-neutral-100">{cpuPercent}%</span>
                </div>
              </div>
              <div className="mt-3 text-center">
                <div className="text-xs font-medium text-neutral-300">CPU</div>
                <div className="text-xs text-neutral-500">{resources?.used.cpu ?? 0} / {resources?.limits.cpu ?? 0}%</div>
              </div>
            </div>

            <div className="flex flex-col items-center">
              <div className="relative">
                <ResourceRing percent={diskPercent} color="#fbbf24" />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-lg font-semibold text-neutral-100">{diskPercent}%</span>
                </div>
              </div>
              <div className="mt-3 text-center">
                <div className="text-xs font-medium text-neutral-300">Disk</div>
                <div className="text-xs text-neutral-500">{resources?.used.disk ?? 0} / {resources?.limits.disk ?? 0} MB</div>
              </div>
            </div>

            <div className="flex flex-col items-center">
              <div className="relative">
                <ResourceRing percent={serverPercent} color="#34d399" />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-lg font-semibold text-neutral-100">{serverPercent}%</span>
                </div>
              </div>
              <div className="mt-3 text-center">
                <div className="text-xs font-medium text-neutral-300">Servers</div>
                <div className="text-xs text-neutral-500">{resources?.used.servers ?? 0} / {resources?.limits.servers ?? 0}</div>
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-xl bg-neutral-800/30 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="text-sm font-medium text-neutral-300">Servers</div>
            <Link to="/console/servers" className="text-xs text-neutral-500 hover:text-neutral-300 transition-colors">
              View all
            </Link>
          </div>
          
          {servers.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Icons.server className="w-8 h-8 text-neutral-700 mb-3" />
              <div className="text-sm text-neutral-400">No servers yet</div>
              <Link to="/console/servers" className="text-xs text-sky-400 hover:text-sky-300 mt-2 transition-colors">
                Create your first server
              </Link>
            </div>
          ) : (
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {servers.map(server => (
                <QuickServer key={server.id} server={server} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
