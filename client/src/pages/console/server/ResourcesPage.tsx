import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Server, getServer, updateServerResources } from '../../../lib/api';
import { useAsyncCallback } from '../../../hooks/useAsync';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Input, PermissionDenied } from '../../../components';

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
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server.name}</span>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Resources</span>
      </div>

      <div>
        <h1 className="text-xl font-semibold text-neutral-100">Resources</h1>
        <p className="text-sm text-neutral-400">Set limits for this server. Changes may require a restart.</p>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-6 pt-6 pb-3">
          <h3 className="text-lg font-semibold text-neutral-100">Allocation</h3>
        </div>
        <div className="px-6 pb-6 pt-2">
          <form onSubmit={handleSave} className="space-y-4 max-w-md">
            <div>
              <Input
                label="Memory (MiB)"
                inputMode="numeric"
                value={form.memory}
                onChange={e => setForm(f => ({ ...f, memory: e.target.value }))}
              />
              <p className="text-xs text-neutral-500 mt-1">Current: {server.memory} MiB (~{(server.memory / 1024).toFixed(2)} GiB)</p>
            </div>
            <div>
              <Input
                label="Disk (MiB)"
                inputMode="numeric"
                value={form.disk}
                onChange={e => setForm(f => ({ ...f, disk: e.target.value }))}
              />
              <p className="text-xs text-neutral-500 mt-1">Current: {server.disk} MiB (~{(server.disk / 1024).toFixed(2)} GiB)</p>
            </div>
            <div>
              <Input
                label="CPU (%)"
                inputMode="numeric"
                value={form.cpu}
                onChange={e => setForm(f => ({ ...f, cpu: e.target.value }))}
              />
              <p className="text-xs text-neutral-500 mt-1">100% = 1 core</p>
            </div>
            <div className="flex gap-2 pt-2">
              <Button type="submit" loading={loading} disabled={!hasChanges}>Save</Button>
              <Button type="button" variant="ghost" onClick={handleReset} disabled={!hasChanges}>Reset</Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
