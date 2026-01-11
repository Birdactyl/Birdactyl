import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useParams } from 'react-router-dom';
import { getServer, getSubusers, updateSubuser, type Server, type Subuser } from '../../../lib/api';
import { PermissionGroups, PermissionLabels } from '../../../lib/permissions';
import { notify, Button, Icons, Table, Checkbox, PermissionDenied } from '../../../components';
import { AddSubuserModal, RemoveSubuserModal } from '../../../components/modals';

export default function SubusersPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [subusers, setSubusers] = useState<Subuser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [editedPerms, setEditedPerms] = useState<Record<string, string[]>>({});
  const [saving, setSaving] = useState(false);
  const [showAdd, setShowAdd] = useState(false);
  const [removeSubuser, setRemoveSubuser] = useState<Subuser | null>(null);

  useEffect(() => {
    if (!id) return;
    Promise.all([
      getServer(id),
      getSubusers(id)
    ]).then(([serverRes, subusersRes]) => {
      if (serverRes.success && serverRes.data) setServer(serverRes.data);
      if (subusersRes.success && subusersRes.data) {
        setSubusers(subusersRes.data);
      } else if (subusersRes.error === 'Permission denied') {
        setError("You don't have permission to manage subusers");
      }
      setLoading(false);
    });
  }, [id]);

  const toggleExpand = (subuserId: string) => {
    const next = new Set(expanded);
    if (next.has(subuserId)) {
      next.delete(subuserId);
      setEditedPerms(p => { const n = { ...p }; delete n[subuserId]; return n; });
    } else {
      next.add(subuserId);
      const sub = subusers.find(s => s.id === subuserId);
      if (sub) setEditedPerms(p => ({ ...p, [subuserId]: [...sub.permissions] }));
    }
    setExpanded(next);
  };

  const togglePerm = (subuserId: string, perm: string) => {
    setEditedPerms(p => {
      const current = p[subuserId] || [];
      return { ...p, [subuserId]: current.includes(perm) ? current.filter(x => x !== perm) : [...current, perm] };
    });
  };

  const toggleGroup = (subuserId: string, group: string[]) => {
    setEditedPerms(p => {
      const current = p[subuserId] || [];
      const allSelected = group.every(g => current.includes(g));
      return { ...p, [subuserId]: allSelected ? current.filter(x => !group.includes(x)) : [...new Set([...current, ...group])] };
    });
  };

  const hasChanges = Object.entries(editedPerms).some(([subId, perms]) => {
    const original = subusers.find(s => s.id === subId)?.permissions || [];
    return JSON.stringify([...perms].sort()) !== JSON.stringify([...original].sort());
  });

  const handleReset = () => {
    const reset: Record<string, string[]> = {};
    expanded.forEach(subId => {
      const sub = subusers.find(s => s.id === subId);
      if (sub) reset[subId] = [...sub.permissions];
    });
    setEditedPerms(reset);
  };

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    const updates = Object.entries(editedPerms).filter(([subId, perms]) => {
      const original = subusers.find(s => s.id === subId)?.permissions || [];
      return JSON.stringify([...perms].sort()) !== JSON.stringify([...original].sort());
    });

    for (const [subId, perms] of updates) {
      const res = await updateSubuser(id, subId, perms);
      if (res.success && res.data) {
        setSubusers(subs => subs.map(s => s.id === subId ? res.data! : s));
      }
    }
    setSaving(false);
    notify('Saved', 'Permissions updated', 'success');
  };

  const handleRemoved = (subuserId: string) => {
    setSubusers(s => s.filter(x => x.id !== subuserId));
    setExpanded(e => { const n = new Set(e); n.delete(subuserId); return n; });
    setEditedPerms(p => { const n = { ...p }; delete n[subuserId]; return n; });
  };

  if (loading || !server) return error ? <PermissionDenied message={error} /> : <div className="text-neutral-400">Loading...</div>;

  const columns = [
    { key: 'expand', header: '', className: 'w-8', render: (sub: Subuser) => (
      <button onClick={() => toggleExpand(sub.id)} className="text-neutral-400 hover:text-neutral-200 transition">
        <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(sub.id) ? 'rotate-90' : ''}`} />
      </button>
    )},
    { key: 'user', header: 'User', render: (sub: Subuser) => (
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-neutral-700 flex items-center justify-center text-sm font-medium text-neutral-200">
          {sub.user?.username?.[0]?.toUpperCase() || '?'}
        </div>
        <div>
          <div className="text-sm font-medium text-neutral-100">{sub.user?.username}</div>
          <div className="text-xs text-neutral-500">{sub.user?.email}</div>
        </div>
      </div>
    )},
    { key: 'perms', header: 'Permissions', render: (sub: Subuser) => {
      const count = (editedPerms[sub.id] || sub.permissions).length;
      return <span className="text-sm text-neutral-400">{count} permission{count !== 1 ? 's' : ''}</span>;
    }},
    { key: 'actions', header: '', align: 'right' as const, render: (sub: Subuser) => (
      <button onClick={() => setRemoveSubuser(sub)} className="text-neutral-400 hover:text-red-400 transition p-1">
        <Icons.trash className="w-4 h-4" />
      </button>
    )},
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server.name}</span>
        <span>/</span>
        <span className="font-semibold text-neutral-100">Subusers</span>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-100">Subusers</h1>
          <p className="text-sm text-neutral-400">Manage who has access to this server and what they can do.</p>
        </div>
        <Button onClick={() => setShowAdd(true)}><Icons.plus className="w-4 h-4" />Add User</Button>
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{subusers.length} subuser{subusers.length !== 1 ? 's' : ''}</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table
            columns={columns}
            data={subusers}
            keyField="id"
            emptyText="No subusers yet"
            expandable={{
              isExpanded: sub => expanded.has(sub.id),
              render: sub => {
                const perms = editedPerms[sub.id] || sub.permissions;
                return (
                  <div className="space-y-4 py-2">
                    {Object.entries(PermissionGroups).map(([group, groupPerms]) => {
                      const allSelected = groupPerms.every(p => perms.includes(p));
                      const someSelected = groupPerms.some(p => perms.includes(p));
                      return (
                        <div key={group} className="space-y-2">
                          <div className="flex items-center gap-2">
                            <Checkbox checked={allSelected} indeterminate={someSelected && !allSelected} onChange={() => toggleGroup(sub.id, groupPerms)} />
                            <span className="text-xs font-medium text-neutral-300 uppercase tracking-wider">{group}</span>
                          </div>
                          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 pl-6">
                            {groupPerms.map(perm => (
                              <div
                                key={perm}
                                onClick={() => togglePerm(sub.id, perm)}
                                className="flex items-center gap-2 cursor-pointer group"
                              >
                                <div onClick={e => e.stopPropagation()}>
                                  <Checkbox checked={perms.includes(perm)} onChange={() => togglePerm(sub.id, perm)} />
                                </div>
                                <span className="text-xs text-neutral-400 group-hover:text-neutral-200 transition">{PermissionLabels[perm] || perm}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                );
              }
            }}
          />
        </div>
      </div>

      <AddSubuserModal
        open={showAdd}
        serverId={id || ''}
        onClose={() => setShowAdd(false)}
        onAdded={subuser => setSubusers(s => [...s, subuser])}
      />

      <RemoveSubuserModal
        open={!!removeSubuser}
        serverId={id || ''}
        subuserId={removeSubuser?.id || ''}
        username={removeSubuser?.user?.username || 'this user'}
        onClose={() => setRemoveSubuser(null)}
        onRemoved={() => removeSubuser && handleRemoved(removeSubuser.id)}
      />

      {hasChanges && createPortal(
        <div className="fixed inset-x-0 bottom-0 z-[95] transition-all duration-200 ease-out">
          <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
            <div className="rounded-t-lg border border-neutral-800 bg-neutral-900/95 px-3 py-2 shadow-2xl backdrop-blur">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm text-neutral-300">You have unsaved changes.</div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" onClick={handleReset}>Reset</Button>
                  <Button onClick={handleSave} loading={saving}>Save changes</Button>
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
