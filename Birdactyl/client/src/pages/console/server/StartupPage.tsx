import { useState, useEffect } from 'react';

import { useParams } from 'react-router-dom';
import { Server, getServer, updateServerVariables } from '../../../lib/api';
import Input from '../../../components/ui/Input';
import Button from '../../../components/ui/Button';
import { notify } from '../../../components/feedback/Notification';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { PermissionDenied, Icons, FloatingBar } from '../../../components';

function SectionCard({ title, description, children, footer, badge }: {
  title: string; description?: string; children: React.ReactNode; footer?: React.ReactNode; badge?: React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-neutral-800 overflow-hidden">
      <div className="px-6 py-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-base font-semibold text-neutral-100">{title}</h3>
            {description && <p className="mt-1 text-sm text-neutral-400">{description}</p>}
          </div>
          {badge}
        </div>
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

export default function StartupPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
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
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Startup Command"
        description="The command executed when your server boots up."
        badge={
          !pkg.startup_editable ? (
            <span className="text-[10px] font-medium text-neutral-500 bg-neutral-800 px-2 py-1 rounded-md shrink-0">Read Only</span>
          ) : undefined
        }
      >
        <div className="rounded-lg border border-neutral-800 overflow-hidden">
          <div className="flex items-center gap-2 px-3 py-2 bg-neutral-900/60 border-b border-neutral-800">
            <Icons.monitor className="w-3.5 h-3.5 text-neutral-500" />
            <span className="text-[11px] font-medium text-neutral-500 select-none">startup.sh</span>
          </div>
          <textarea
            value={form.startup}
            onChange={e => setForm(f => ({ ...f, startup: e.target.value }))}
            disabled={!pkg.startup_editable}
            rows={3}
            className="w-full bg-[#0a0a0a] px-4 py-3 text-[13px] font-mono text-neutral-200 placeholder:text-neutral-600 resize-none focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
            placeholder="java -Xms128M -Xmx{{SERVER_MEMORY}}M -jar server.jar"
          />
        </div>
      </SectionCard>

      <SectionCard
        title="Docker Image"
        description="The container image used to run your server."
        badge={
          !pkg.docker_image_editable ? (
            <span className="text-[10px] font-medium text-neutral-500 bg-neutral-800 px-2 py-1 rounded-md shrink-0">Read Only</span>
          ) : undefined
        }
      >
        <div className="flex items-start gap-4 p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
          <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-800 shrink-0 mt-0.5">
            <Icons.cube className="w-5 h-5 text-neutral-400" />
          </div>
          <div className="flex-1 min-w-0">
            <Input
              value={form.dockerImage}
              onChange={e => setForm(f => ({ ...f, dockerImage: e.target.value }))}
              disabled={!pkg.docker_image_editable}
              placeholder="ghcr.io/pterodactyl/yolks:java_21"
            />
            <p className="text-xs text-neutral-500 mt-1.5">Use a default image from the package or provide your own registry reference.</p>
          </div>
        </div>
      </SectionCard>

      <SectionCard
        title="Environment Variables"
        description="Override environment variables exposed by your package. Defaults are shown below each field."
      >
        {pkg.variables && pkg.variables.length > 0 ? (
          <div className="space-y-3">
            {pkg.variables.map((v) => {
              const isEditable = v.user_editable;
              const currentValue = form.variables[v.name] ?? v.default;
              const isModified = form.variables[v.name] !== undefined && form.variables[v.name] !== v.default;
              const friendlyName = v.name.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());

              return (
                <div key={v.name} className={`p-4 rounded-lg border transition-colors ${isModified ? 'border-sky-500/30 bg-sky-500/[0.02]' : 'border-neutral-800 bg-neutral-900/30'
                  }`}>
                  <div className="flex items-start justify-between gap-3 mb-3">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-neutral-200">{friendlyName}</span>
                        {!isEditable && (
                          <span className="text-[10px] font-medium text-neutral-500 bg-neutral-800 px-1.5 py-0.5 rounded">Locked</span>
                        )}
                        {isModified && (
                          <span className="text-[10px] font-medium text-sky-400 bg-sky-500/10 px-1.5 py-0.5 rounded">Modified</span>
                        )}
                      </div>
                      {v.description && <p className="text-xs text-neutral-500 mt-0.5">{v.description}</p>}
                    </div>
                    <code className="text-[11px] font-mono text-neutral-500 bg-neutral-800 px-1.5 py-0.5 rounded shrink-0">{v.name}</code>
                  </div>
                  <input
                    value={currentValue}
                    onChange={e => setForm(f => ({ ...f, variables: { ...f.variables, [v.name]: e.target.value } }))}
                    disabled={!isEditable}
                    className="w-full rounded-lg border border-neutral-800 bg-neutral-800/60 px-3 py-2 text-[13px] font-mono text-neutral-100 placeholder:text-neutral-600 transition hover:border-neutral-700 focus:outline-none focus:ring-1 focus:ring-neutral-600 focus:border-neutral-600 disabled:opacity-50 disabled:cursor-not-allowed"
                  />
                  <div className="mt-1.5 text-[11px] text-neutral-600 font-mono">
                    Default: <code>{v.default || '\u2014'}</code>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-neutral-800 py-8 text-center">
            <Icons.cube className="w-6 h-6 text-neutral-700 mx-auto mb-2" />
            <p className="text-sm text-neutral-500">No variables defined for this package.</p>
          </div>
        )}
      </SectionCard>

      <FloatingBar show={hasChanges}>
        <div className="text-sm text-neutral-300">You have unsaved changes.</div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={handleReset}>Reset</Button>
          <Button onClick={handleSave} loading={saving}>Save Changes</Button>
        </div>
      </FloatingBar>
    </div>
  );
}
