import { useState, useEffect, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { useParams } from 'react-router-dom';
import { getServer, Server, listBackups, createBackup, deleteBackup, restoreBackup, getBackupDownloadUrl, Backup } from '../../../lib/api';
import { formatBytes } from '../../../lib/utils';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Icons, Modal, Input, Checkbox, PermissionDenied } from '../../../components';
import ContextMenuItem from '../../../components/files/ContextMenuItem';
import { notify } from '../../../components/feedback/Notification';

function BackupContextMenu({ backup, position, onDownload, onRestore, onDelete, canRestore }: {
  backup: Backup;
  position: { x: number; y: number; openUp: boolean };
  onDownload: () => void;
  onRestore: () => void;
  onDelete: () => void;
  canRestore: boolean;
}) {
  return createPortal(
    <div
      className="fixed z-[9999] w-44 rounded-md border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 shadow-xl p-1 animate-dropdown-in"
      style={{
        right: `calc(100vw - ${position.x}px)`,
        ...(position.openUp ? { bottom: window.innerHeight - position.y - 80 } : { top: position.y }),
      }}
      role="menu"
    >
      <ContextMenuItem icon={<Icons.download />} label="Download" onClick={onDownload} disabled={!backup.completed} />
      {canRestore && <ContextMenuItem icon={<Icons.refresh />} label="Restore" onClick={onRestore} disabled={!backup.completed} />}
      <ContextMenuItem icon={<Icons.trash />} label="Delete" onClick={onDelete} destructive />
    </div>,
    document.body
  );
}

export default function BackupsPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [backups, setBackups] = useState<Backup[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [createModal, setCreateModal] = useState({ open: false, name: '', loading: false });
  const [deleteModal, setDeleteModal] = useState<{ backup: Backup; loading: boolean } | null>(null);
  const [restoreModal, setRestoreModal] = useState<{ backup: Backup; loading: boolean } | null>(null);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; openUp: boolean; backup: Backup } | null>(null);
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => res.success && res.data && setServer(res.data));
    loadBackups();
  }, [id]);

  useEffect(() => {
    const hasInProgress = backups.some(b => !b.completed);
    if (!hasInProgress) return;
    const interval = setInterval(loadBackups, 2000);
    return () => clearInterval(interval);
  }, [backups, id]);

  useEffect(() => {
    if (!contextMenu) return;
    const h = () => setContextMenu(null);
    document.addEventListener('click', h);
    return () => document.removeEventListener('click', h);
  }, [contextMenu]);

  const loadBackups = async () => {
    if (!id) return;
    const res = await listBackups(id);
    if (res.success && res.data) {
      setBackups(res.data);
    }
    setLoading(false);
  };

  const handleCreate = async () => {
    if (!id) return;
    setCreateModal(s => ({ ...s, loading: true }));
    const res = await createBackup(id, createModal.name || undefined);
    if (res.success) {
      notify('Backup started', 'Your backup is being created', 'success');
      setCreateModal({ open: false, name: '', loading: false });
      loadBackups();
    } else {
      notify('Error', res.error || 'Failed to create backup', 'error');
      setCreateModal(s => ({ ...s, loading: false }));
    }
  };

  const handleDownload = (backup: Backup) => {
    if (!id) return;
    setContextMenu(null);
    const a = document.createElement('a');
    a.href = getBackupDownloadUrl(id, backup.id);
    a.download = backup.name + '.tar.gz';
    a.click();
  };

  const handleDelete = async () => {
    if (!deleteModal || !id) return;
    setDeleteModal(s => s && { ...s, loading: true });
    const res = await deleteBackup(id, deleteModal.backup.id);
    if (res.success) {
      notify('Deleted', 'Backup has been deleted', 'success');
      setBackups(b => b.filter(x => x.id !== deleteModal.backup.id));
      setSelected(s => { const n = new Set(s); n.delete(deleteModal.backup.id); return n; });
      setDeleteModal(null);
    } else {
      notify('Error', res.error || 'Failed to delete backup', 'error');
      setDeleteModal(s => s && { ...s, loading: false });
    }
  };

  const handleRestore = async () => {
    if (!restoreModal || !id) return;
    setRestoreModal(s => s && { ...s, loading: true });
    const res = await restoreBackup(id, restoreModal.backup.id);
    if (res.success) {
      notify('Restored', 'Backup has been restored successfully', 'success');
      setRestoreModal(null);
    } else {
      notify('Error', res.error || 'Failed to restore backup', 'error');
      setRestoreModal(s => s && { ...s, loading: false });
    }
  };

  const handleBulkDelete = async () => {
    if (!id) return;
    const toDelete = [...selected];
    for (const backupId of toDelete) {
      await deleteBackup(id, backupId);
    }
    notify('Deleted', `${toDelete.length} backup(s) deleted`, 'success');
    setBackups(b => b.filter(x => !selected.has(x.id)));
    setSelected(new Set());
  };

  const openContextMenu = (e: React.MouseEvent<HTMLButtonElement>, backup: Backup) => {
    e.stopPropagation();
    if (contextMenu?.backup === backup) { setContextMenu(null); return; }
    const rect = e.currentTarget.getBoundingClientRect();
    const menuH = 80;
    const openUp = window.innerHeight - rect.bottom < menuH && rect.top > window.innerHeight - rect.bottom;
    setContextMenu({ x: rect.right - 48, y: openUp ? rect.top : rect.bottom + 4, openUp, backup });
  };

  const toggleSelect = (id: string) => setSelected(s => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n; });
  const toggleAll = () => setSelected(s => s.size === backups.length ? new Set() : new Set(backups.map(b => b.id)));
  const allSelected = useMemo(() => backups.length > 0 && selected.size === backups.length, [backups, selected]);
  const someSelected = useMemo(() => selected.size > 0 && selected.size < backups.length, [backups, selected]);

  const formatDateTime = (ts: number) => new Date(ts * 1000).toLocaleString();

  if (permsLoading) return null;
  if (!can('backup.list')) return <PermissionDenied message="You don't have permission to view backups" />;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">Backups</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Backups</h1>
          <p className="text-sm text-neutral-400">Create and restore backups of your server.</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={loadBackups}><Icons.refresh className="h-4 w-4" /></Button>
          {can('backup.create') && <Button onClick={() => setCreateModal(s => ({ ...s, open: true }))}><Icons.plus className="h-4 w-4 mr-1.5" />New Backup</Button>}
        </div>
      </div>

      <div className="bg-neutral-900/40 rounded-lg p-1">
        <table className="min-w-full table-fixed border-separate border-spacing-0">
          <colgroup>
            <col style={{ width: '40px' }} />
            <col style={{ width: '40%' }} />
            <col style={{ width: '15%' }} />
            <col style={{ width: '20%' }} />
            <col style={{ width: '12%' }} />
            <col style={{ width: '10%' }} />
          </colgroup>
          <thead className="bg-transparent">
            <tr>
              <th className="pl-4 py-2 text-left"><Checkbox checked={allSelected} indeterminate={someSelected} onChange={toggleAll} /></th>
              <th className="pl-2 pr-6 py-2 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">Name</th>
              <th className="pl-3 pr-6 py-2 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">Size</th>
              <th className="pl-3 pr-6 py-2 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">Created</th>
              <th className="pl-3 pr-6 py-2 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">Status</th>
              <th className="px-6 py-2"><span className="sr-only">Actions</span></th>
            </tr>
          </thead>
        </table>
        <div className="mt-1 rounded-lg border border-neutral-800 overflow-hidden">
          <table className="min-w-full table-fixed border-separate border-spacing-0">
            <colgroup>
              <col style={{ width: '40px' }} />
              <col style={{ width: '40%' }} />
              <col style={{ width: '15%' }} />
              <col style={{ width: '20%' }} />
              <col style={{ width: '12%' }} />
              <col style={{ width: '10%' }} />
            </colgroup>
            <tbody className="bg-neutral-900/50 divide-y divide-neutral-700">
              {loading ? (
                <tr><td colSpan={6} className="px-6 py-8 text-center text-sm text-neutral-500">&nbsp;</td></tr>
              ) : backups.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-8 text-center text-sm text-neutral-500">No backups yet</td></tr>
              ) : backups.map(backup => (
                <tr key={backup.id} className={`hover:bg-neutral-800/50 ${selected.has(backup.id) ? 'bg-neutral-800/30' : ''}`}>
                  <td className="pl-4 py-3" onClick={e => e.stopPropagation()}>
                    <Checkbox checked={selected.has(backup.id)} onChange={() => toggleSelect(backup.id)} />
                  </td>
                  <td className="pl-2 pr-6 py-3">
                    <div className="flex items-center gap-3 text-sm">
                      <Icons.archive className="w-5 h-5 text-blue-500 shrink-0" />
                      <span className="text-neutral-100 truncate">{backup.name}</span>
                    </div>
                  </td>
                  <td className="pl-3 pr-6 py-3 text-sm text-neutral-400">{backup.completed ? formatBytes(backup.size) : '—'}</td>
                  <td className="pl-3 pr-6 py-3 text-sm text-neutral-400">{formatDateTime(backup.created_at)}</td>
                  <td className="pl-3 pr-6 py-3">
                    {backup.completed ? (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/20 text-emerald-400">Completed</span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-medium bg-amber-500/20 text-amber-400">
                        <span className="inline-block w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin" />
                        In progress
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-3 text-right">
                    <Button variant="ghost" onClick={e => openContextMenu(e, backup)}>
                      <Icons.ellipsis className="w-5 h-5" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {contextMenu && (
        <BackupContextMenu
          backup={contextMenu.backup}
          position={contextMenu}
          onDownload={() => handleDownload(contextMenu.backup)}
          onRestore={() => { setRestoreModal({ backup: contextMenu.backup, loading: false }); setContextMenu(null); }}
          onDelete={() => { setDeleteModal({ backup: contextMenu.backup, loading: false }); setContextMenu(null); }}
          canRestore={can('backup.restore')}
        />
      )}

      <Modal open={createModal.open} onClose={() => !createModal.loading && setCreateModal(s => ({ ...s, open: false }))} title="Create backup" description="Optionally name your backup">
        <div className="space-y-4 pt-2">
          <div className="space-y-2">
            <label className="text-sm font-medium text-neutral-300">Name</label>
            <Input
              placeholder="Optional"
              value={createModal.name}
              onChange={e => setCreateModal(s => ({ ...s, name: e.target.value }))}
            />
            <p className="text-xs text-neutral-500">Leave blank to auto-generate.</p>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <Button variant="ghost" onClick={() => setCreateModal(s => ({ ...s, open: false }))} disabled={createModal.loading}>Cancel</Button>
            <Button onClick={handleCreate} loading={createModal.loading}>Create</Button>
          </div>
        </div>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Delete backup" description={`Are you sure you want to delete "${deleteModal?.backup.name}"? This cannot be undone.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button variant="danger" onClick={handleDelete} loading={deleteModal?.loading}>Delete</Button>
        </div>
      </Modal>

      <Modal open={!!restoreModal} onClose={() => !restoreModal?.loading && setRestoreModal(null)} title="Restore backup" description={`Are you sure you want to restore "${restoreModal?.backup.name}"? This will replace all current server files. The server must be stopped.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setRestoreModal(null)} disabled={restoreModal?.loading}>Cancel</Button>
          <Button variant="danger" onClick={handleRestore} loading={restoreModal?.loading}>Restore</Button>
        </div>
      </Modal>

      {selected.size > 0 && createPortal(
        <div className="fixed inset-x-0 bottom-0 z-[95]">
          <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
            <div className="rounded-t-lg border border-neutral-200 dark:border-neutral-800 bg-white/95 dark:bg-neutral-900/95 shadow-2xl backdrop-blur px-3 py-2">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm text-neutral-700 dark:text-neutral-300"><span className="font-medium">{selected.size}</span> selected</div>
                <div className="flex items-center gap-2">
                  <Button variant="danger" onClick={handleBulkDelete}><Icons.trash className="h-4 w-4 mr-1" />Delete</Button>
                  <Button variant="ghost" onClick={() => setSelected(new Set())}>Clear</Button>
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
