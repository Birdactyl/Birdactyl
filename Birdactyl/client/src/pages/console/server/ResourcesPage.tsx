import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Server, getServer, updateServerResources } from '../../../lib/api';
import { useAsyncCallback } from '../../../hooks/useAsync';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Icons, PermissionDenied } from '../../../components';

function SectionCard({ title, description, children, footer }: {
  title: string; description?: string; children: React.ReactNode; footer?: React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-neutral-800 overflow-hidden">
      <div className="px-6 py-5">
        <h3 className="text-base font-semibold text-neutral-100">{title}</h3>
        {description && <p className="mt-1 text-sm text-neutral-400">{description}</p>}
        <div className="mt-5">{children}</div>
      </div>
      {footer && (
        <div className="px-6 py-3.5 bg-neutral-900/50 border-t border-neutral-800 flex items-center justify-end gap-2">
          {footer}
        </div>
      )}
    </div>
  );
}

function ResourceField({ label, hint, unit, value, onChange, icon }: {
  label: string; hint: string; unit: string; value: string; onChange: (v: string) => void; icon: keyof typeof Icons;
}) {
  const IconComponent = Icons[icon];
  return (
    <div className="flex items-start gap-4 p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
      <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-800 shrink-0 mt-0.5">
        <IconComponent className="w-5 h-5 text-neutral-400" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-2">
          <label className="text-sm font-medium text-neutral-200">{label}</label>
          <span className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide">{unit}</span>
        </div>
        <input
          type="text"
          inputMode="numeric"
          value={value}
          onChange={e => onChange(e.target.value)}
          className="w-full rounded-lg border border-neutral-800 bg-neutral-800/60 px-3 py-2 text-sm font-mono text-neutral-100 placeholder:text-neutral-600 transition hover:border-neutral-700 focus:outline-none focus:ring-1 focus:ring-neutral-600 focus:border-neutral-600 tabular-nums"
        />
        <p className="text-xs text-neutral-500 mt-1.5">{hint}</p>
      </div>
    </div>
  );
}

export default function ResourcesPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [form, setForm] = useState({ memory: '', cpu: '', disk: '' });
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => {
      if (res.success && res.data) {
        setServer(res.data);
        setForm({ memory: String(res.data.memory), cpu: String(res.data.cpu), disk: String(res.data.disk) });
      }
    });
  }, [id]);

  const hasChanges = server && (form.memory !== String(server.memory) || form.cpu !== String(server.cpu) || form.disk !== String(server.disk));

  const handleReset = () => { if (server) setForm({ memory: String(server.memory), cpu: String(server.cpu), disk: String(server.disk) }); };

  const [handleSave, loading] = useAsyncCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    const res = await updateServerResources(id, parseInt(form.memory), parseInt(form.cpu), parseInt(form.disk));
    if (res.success && res.data) setServer(res.data);
  }, [id, form]);

  if (permsLoading) return null;
  if (!can('settings.resources')) return <PermissionDenied message="You don't have permission to view resources" />;
  if (!server) return <div className="text-neutral-400">Loading...</div>;

  return (
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Resource Allocation"
        description="Adjust the memory, CPU, and disk limits for this server. Changes may require a restart to take effect."
        footer={
          <>
            <Button type="button" variant="ghost" onClick={handleReset} disabled={!hasChanges}>Reset</Button>
            <Button type="submit" form="resourcesForm" loading={loading} disabled={!hasChanges}>Save Changes</Button>
          </>
        }
      >
        <form id="resourcesForm" onSubmit={handleSave} className="space-y-3">
          <ResourceField
            label="Memory"
            unit="MiB"
            hint={`Currently allocated ${server.memory} MiB (~${(server.memory / 1024).toFixed(1)} GiB)`}
            value={form.memory}
            onChange={v => setForm(f => ({ ...f, memory: v }))}
            icon="pieChart"
          />
          <ResourceField
            label="CPU"
            unit="%"
            hint="100% equals 1 full CPU core. 200% = 2 cores, etc."
            value={form.cpu}
            onChange={v => setForm(f => ({ ...f, cpu: v }))}
            icon="cpu"
          />
          <ResourceField
            label="Disk"
            unit="MiB"
            hint={`Currently allocated ${server.disk} MiB (~${(server.disk / 1024).toFixed(1)} GiB)`}
            value={form.disk}
            onChange={v => setForm(f => ({ ...f, disk: v }))}
            icon="disk"
          />
        </form>
      </SectionCard>

      <SectionCard
        title="Current Usage"
        description="A summary of the resources currently assigned to this server."
      >
        <div className="grid grid-cols-3 gap-3">
          <div className="p-3 rounded-lg border border-neutral-800 bg-neutral-900/30 text-center">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-1">Memory</div>
            <div className="text-lg font-semibold text-neutral-100 tabular-nums">{server.memory}</div>
            <div className="text-xs text-neutral-500">MiB</div>
          </div>
          <div className="p-3 rounded-lg border border-neutral-800 bg-neutral-900/30 text-center">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-1">CPU</div>
            <div className="text-lg font-semibold text-neutral-100 tabular-nums">{server.cpu}%</div>
            <div className="text-xs text-neutral-500">{Math.ceil(server.cpu / 100)} core{Math.ceil(server.cpu / 100) !== 1 ? 's' : ''}</div>
          </div>
          <div className="p-3 rounded-lg border border-neutral-800 bg-neutral-900/30 text-center">
            <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-1">Disk</div>
            <div className="text-lg font-semibold text-neutral-100 tabular-nums">{(server.disk / 1024).toFixed(1)}</div>
            <div className="text-xs text-neutral-500">GiB</div>
          </div>
        </div>
      </SectionCard>
    </div>
  );
}
