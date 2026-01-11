import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useParams } from 'react-router-dom';
import { Server, getServer, updateServerVariables } from '../../../lib/api';
import { Card } from '../../../components/ui/Card';
import Input from '../../../components/ui/Input';
import Button from '../../../components/ui/Button';
import { notify } from '../../../components/feedback/Notification';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { PermissionDenied } from '../../../components';

export default function StartupPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [tab, setTab] = useState<'startup' | 'variables'>('startup');
  const [form, setForm] = useState({ startup: '', dockerImage: '', variables: {} as Record<string, string> });
  const [original, setOriginal] = useState({ startup: '', dockerImage: '', variables: {} as Record<string, string> });
  const [saving, setSaving] = useState(false);
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => {
      if (res.success && res.data) {
        setServer(res.data);
        const pkg = res.data.package;
        const serverVars = typeof res.data.variables === 'string' ? JSON.parse(res.data.variables) : res.data.variables || {};
        const state = {
          startup: res.data.startup || pkg?.startup || '',
          dockerImage: res.data.docker_image || pkg?.docker_image || '',
          variables: serverVars
        };
        setForm(state);
        setOriginal(state);
      }
    });
  }, [id]);

  const hasChanges = form.startup !== original.startup || form.dockerImage !== original.dockerImage || JSON.stringify(form.variables) !== JSON.stringify(original.variables);
  const handleReset = () => setForm(original);
  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    const res = await updateServerVariables(id, form.variables, form.startup, form.dockerImage);
    setSaving(false);
    if (res.success) {
      setOriginal(form);
      notify('Saved', 'Startup configuration updated', 'success');
    } else {
      notify('Error', res.error || 'Failed to save', 'error');
    }
  };

  if (permsLoading) return null;
  if (!can('startup.view')) return <PermissionDenied message="You don't have permission to view startup settings" />;
  if (!server || !server.package) return <div className="text-neutral-400">Loading...</div>;

  const pkg = server.package;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server.name}</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">Startup</span>
      </div>

      <div className="space-y-1">
        <h1 className="text-2xl font-semibold text-neutral-100">Startup</h1>
        <p className="text-sm text-neutral-400">Control the command, container image, and variables used when your server boots.</p>
      </div>

      <div className="-mx-2 mb-3 overflow-x-auto px-2">
        <div className="inline-flex items-center rounded-lg bg-neutral-800 p-0.5 whitespace-nowrap">
          <button
            type="button"
            onClick={() => setTab('startup')}
            className={`inline-flex items-center cursor-pointer justify-center whitespace-nowrap rounded-lg px-3 h-7 text-[13px] font-medium transition-colors ${
              tab === 'startup'
                ? 'bg-neutral-700 text-neutral-100 shadow-xs border border-neutral-700'
                : 'text-neutral-400 hover:text-neutral-100 border border-transparent'
            }`}
          >
            Startup
          </button>
          <button
            type="button"
            onClick={() => setTab('variables')}
            className={`inline-flex items-center cursor-pointer justify-center whitespace-nowrap rounded-lg px-3 h-7 text-[13px] font-medium transition-colors ${
              tab === 'variables'
                ? 'bg-neutral-700 text-neutral-100 shadow-xs border border-neutral-700'
                : 'text-neutral-400 hover:text-neutral-100 border border-transparent'
            }`}
          >
            Variables
          </button>
        </div>
      </div>

      {tab === 'startup' ? (
        <Card title="Command & image" description="Update the startup command and container image.">
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="block text-xs font-medium text-neutral-400">Startup command</label>
              <textarea
                value={form.startup}
                onChange={e => setForm(f => ({ ...f, startup: e.target.value }))}
                disabled={!pkg.startup_editable}
                rows={4}
                className="w-full rounded-lg border border-neutral-800 bg-neutral-800/80 px-3 py-2 text-[13px] text-neutral-100 placeholder:text-neutral-500 shadow-xs transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 focus:border-neutral-500 disabled:opacity-60 disabled:cursor-not-allowed"
                placeholder="Enter the command executed when the server boots"
              />
            </div>
            <div className="space-y-2">
              <label className="block text-xs font-medium text-neutral-400">Docker image</label>
              <Input
                value={form.dockerImage}
                onChange={e => setForm(f => ({ ...f, dockerImage: e.target.value }))}
                disabled={!pkg.docker_image_editable}
                placeholder="ghcr.io/pterodactyl/yolks:java_21"
              />
              <p className="text-[11px] text-neutral-500">Select one of the default images, or provide your own registry reference.</p>
            </div>
          </div>
        </Card>
      ) : (
        <Card title="Variables" description="Override environment variables exposed by your package. Defaults are shown for reference.">
          <div className="space-y-4">
            {pkg.variables?.map((v) => (
              <div key={v.name} className="rounded-lg border border-neutral-800 bg-neutral-900 p-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-1">
                    <label className="block text-xs font-medium text-neutral-400">{v.name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}</label>
                    {v.description && <p className="text-xs text-neutral-400 whitespace-pre-wrap">{v.description}</p>}
                  </div>
                  <span className="rounded bg-neutral-800 px-1.5 py-0.5 text-[10px] font-medium text-neutral-300 font-mono">
                    <code>{v.name}</code>
                  </span>
                </div>
                <input
                  value={form.variables[v.name] ?? v.default}
                  onChange={e => setForm(f => ({ ...f, variables: { ...f.variables, [v.name]: e.target.value } }))}
                  disabled={!v.user_editable}
                  className="w-full rounded-lg border border-neutral-800 bg-neutral-800/80 text-[13px] text-neutral-100 placeholder:text-neutral-500 shadow-xs transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 focus:border-neutral-500 disabled:opacity-60 disabled:cursor-not-allowed px-3 py-2 mt-3"
                />
                <div className="mt-2 text-[11px] text-neutral-500 font-mono">
                  Default: <code>{v.default || '—'}</code>
                </div>
              </div>
            ))}
            {(!pkg.variables || pkg.variables.length === 0) && (
              <p className="text-sm text-neutral-500 text-center py-4">No variables defined for this package.</p>
            )}
          </div>
        </Card>
      )}

      {hasChanges && createPortal(
        <div className="fixed inset-x-0 bottom-0 z-[95] transition-all duration-200 ease-out">
          <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
            <div className="rounded-t-lg border border-neutral-800 bg-neutral-900/95 px-3 py-2 shadow-2xl backdrop-blur">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm text-neutral-300">You have unsaved changes.</div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" onClick={handleReset}>
                    Reset
                  </Button>
                  <Button onClick={handleSave} loading={saving}>
                    Save changes
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>,
        document.body
      )}
    </div>
  );
}
