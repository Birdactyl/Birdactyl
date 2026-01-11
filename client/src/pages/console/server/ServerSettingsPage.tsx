import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, updateServerName, Server } from '../../../lib/api';
import { notify, Button, Input, Icons, PermissionDenied } from '../../../components';
import { DeleteServerModal, ReinstallServerModal } from '../../../components/modals';
import { useServerPermissions } from '../../../hooks/useServerPermissions';

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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-100">Settings</h1>
        <p className="text-sm text-neutral-400">Manage your server configuration and settings.</p>
      </div>

      <div className="rounded-xl bg-neutral-800/30 overflow-hidden">
        <form onSubmit={handleSaveName}>
          <div className="p-6 space-y-6">
            <div className="grid gap-6 lg:grid-cols-2">
              <div className="space-y-4">
                <div className="text-sm font-medium text-neutral-300">Server Details</div>
                <Input label="Name" value={name} onChange={e => setName(e.target.value)} placeholder="My Server" />
                <Input label="Description" value={description} onChange={e => setDescription(e.target.value)} placeholder="Optional description" />
              </div>

              <div className="space-y-4">
                <div className="text-sm font-medium text-neutral-300">Server Information</div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50">
                    <span className="text-xs text-neutral-400">Server ID</span>
                    <span className="text-xs text-neutral-200 font-mono">{server.id}</span>
                  </div>
                  <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50">
                    <span className="text-xs text-neutral-400">Node</span>
                    <span className="text-xs text-neutral-200">{server.node?.name || 'Unknown'}</span>
                  </div>
                  <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50">
                    <span className="text-xs text-neutral-400">Package</span>
                    <span className="text-xs text-neutral-200">{server.package?.name || 'Unknown'}</span>
                  </div>
                  <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50">
                    <span className="text-xs text-neutral-400">Created</span>
                    <span className="text-xs text-neutral-200">{new Date(server.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="px-6 py-4 bg-neutral-900/50 border-t border-neutral-800 flex items-center justify-between">
            {isOwner ? (
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setShowReinstall(true)}
                  className="inline-flex items-center gap-2 px-3 py-1.5 text-xs font-medium text-red-400 rounded-lg hover:bg-red-500/10 transition-colors"
                >
                  <Icons.refresh className="w-4 h-4" />
                  Reinstall
                </button>
                <button
                  type="button"
                  onClick={() => setShowDelete(true)}
                  className="inline-flex items-center gap-2 px-3 py-1.5 text-xs font-medium text-red-400 rounded-lg hover:bg-red-500/10 transition-colors"
                >
                  <Icons.trash className="w-4 h-4" />
                  Delete
                </button>
              </div>
            ) : (
              <div />
            )}
            <Button type="submit" loading={saving} disabled={!name.trim() || !hasChanges}>
              Save Changes
            </Button>
          </div>
        </form>
      </div>

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
