import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, updateServerName, Server } from '../../../lib/api';
import { notify, Button, Input, Icons, PermissionDenied } from '../../../components';
import { DeleteServerModal, ReinstallServerModal } from '../../../components/modals';
import { useServerPermissions } from '../../../hooks/useServerPermissions';

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
        <div className="px-6 py-3.5 bg-neutral-900/50 border-t border-neutral-800 flex items-center justify-end">
          {footer}
        </div>
      )}
    </div>
  );
}

function InfoRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/30 border border-neutral-800/50">
      <span className="text-xs text-neutral-500">{label}</span>
      <span className={`text-xs text-neutral-200 ${mono ? 'font-mono' : ''}`}>{value}</span>
    </div>
  );
}

export default function ServerSettingsPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [saving, setSaving] = useState(false);
  const [showReinstall, setShowReinstall] = useState(false);
  const [showDelete, setShowDelete] = useState(false);
  const { can, loading: permsLoading, isOwner } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => {
      if (res.success && res.data) {
        setServer(res.data);
        setName(res.data.name);
        setDescription(res.data.description || '');
      }
      setLoading(false);
    });
  }, [id]);

  const handleSaveName = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !name.trim()) return;
    setSaving(true);
    const res = await updateServerName(id, name.trim(), description.trim());
    if (res.success) {
      setServer(s => s ? { ...s, name: name.trim(), description: description.trim() } : null);
      notify('Saved', 'Server settings updated', 'success');
    } else {
      notify('Error', res.error || 'Failed to update settings', 'error');
    }
    setSaving(false);
  };

  const hasChanges = name !== server?.name || description !== (server?.description || '');

  if (loading || permsLoading || !server) return null;
  if (!can('settings.view')) return <PermissionDenied message="You don't have permission to view settings" />;

  return (
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Server Details"
        description="Update your server's name and description."
        footer={
          <Button type="submit" form="serverDetailsForm" loading={saving} disabled={!name.trim() || !hasChanges}>
            Save Changes
          </Button>
        }
      >
        <form id="serverDetailsForm" onSubmit={handleSaveName} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input label="Name" value={name} onChange={e => setName(e.target.value)} placeholder="My Server" />
          <Input label="Description" value={description} onChange={e => setDescription(e.target.value)} placeholder="Optional description" />
        </form>
      </SectionCard>

      <SectionCard
        title="Server Information"
        description="Details about this server's configuration."
      >
        <div className="space-y-2">
          <InfoRow label="Server ID" value={server.id} mono />
          <InfoRow label="Node" value={server.node?.name || 'Unknown'} />
          <InfoRow label="Package" value={server.package?.name || 'Unknown'} />
          <InfoRow label="Docker Image" value={server.docker_image} mono />
          <InfoRow label="Created" value={new Date(server.created_at).toLocaleDateString()} />
        </div>
      </SectionCard>

      {isOwner && (
        <div className="rounded-xl border border-red-500/20 overflow-hidden">
          <div className="px-6 py-5">
            <h3 className="text-base font-semibold text-red-400">Danger Zone</h3>
            <p className="mt-1 text-sm text-neutral-400">Irreversible and destructive actions for this server.</p>
            <div className="mt-5 space-y-3">
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
                <div>
                  <div className="text-sm font-medium text-neutral-200">Reinstall Server</div>
                  <div className="text-xs text-neutral-500 mt-0.5">Wipe all files and reinstall the server from scratch.</div>
                </div>
                <Button variant="danger" onClick={() => setShowReinstall(true)} className="shrink-0" style={{ minWidth: 140 }}>
                  <Icons.refresh className="w-4 h-4 mr-1.5" />Reinstall
                </Button>
              </div>
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
                <div>
                  <div className="text-sm font-medium text-neutral-200">Delete Server</div>
                  <div className="text-xs text-neutral-500 mt-0.5">Permanently delete this server and all its data.</div>
                </div>
                <Button variant="danger" onClick={() => setShowDelete(true)} className="shrink-0" style={{ minWidth: 140 }}>
                  <Icons.trash className="w-4 h-4 mr-1.5" />Delete
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}

      <ReinstallServerModal
        open={showReinstall}
        serverId={server.id}
        serverName={server.name}
        onClose={() => setShowReinstall(false)}
      />

      <DeleteServerModal
        open={showDelete}
        serverId={server.id}
        serverName={server.name}
        onClose={() => setShowDelete(false)}
        redirectOnDelete
      />
    </div>
  );
}
