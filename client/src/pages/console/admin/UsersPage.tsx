import { useState, useEffect } from 'react';
import { adminGetUsers, adminGetRegistrationStatus, adminSetRegistrationStatus, adminGetEmailVerificationSettings, adminSetEmailVerificationSettings } from '../../../lib/api';
import { useTable } from '../../../hooks/useTable';
import { notify, Button, Input, Pagination, Checkbox, Icons, ContextMenu, BulkActionBar, Table, SlidePanel } from '../../../components';
import { CreateUserModal, EditUserModal, UserActionModal, UserAPIKeysModal } from '../../../components/modals';

const VerificationActionGroups: Record<string, { label: string; actions: { key: string; label: string }[] }> = {
  auth: {
    label: 'Authentication',
    actions: [
      { key: 'auth.login', label: 'Login' },
    ],
  },
  profile: {
    label: 'Profile',
    actions: [
      { key: 'profile.update', label: 'Update Profile' },
      { key: 'profile.password_change', label: 'Change Password' },
      { key: 'profile.2fa_setup', label: '2FA Setup' },
      { key: 'profile.2fa_enable', label: '2FA Enable' },
      { key: 'profile.2fa_disable', label: '2FA Disable' },
    ],
  },
  server: {
    label: 'Server',
    actions: [
      { key: 'server.create', label: 'Create Server' },
      { key: 'server.delete', label: 'Delete Server' },
      { key: 'server.start', label: 'Start Server' },
      { key: 'server.stop', label: 'Stop Server' },
      { key: 'server.kill', label: 'Kill Server' },
      { key: 'server.restart', label: 'Restart Server' },
      { key: 'server.reinstall', label: 'Reinstall Server' },
      { key: 'server.command', label: 'Send Command' },
    ],
  },
  serverSettings: {
    label: 'Server Settings',
    actions: [
      { key: 'server.name.update', label: 'Rename Server' },
      { key: 'server.resources.update', label: 'Update Resources' },
      { key: 'server.variables.update', label: 'Update Variables' },
    ],
  },
  allocations: {
    label: 'Allocations',
    actions: [
      { key: 'server.allocation.add', label: 'Add Allocation' },
      { key: 'server.allocation.set_primary', label: 'Set Primary' },
      { key: 'server.allocation.delete', label: 'Delete Allocation' },
    ],
  },
  files: {
    label: 'Files',
    actions: [
      { key: 'server.file.create_folder', label: 'Create Folder' },
      { key: 'server.file.write', label: 'Write File' },
      { key: 'server.file.upload', label: 'Upload File' },
      { key: 'server.file.delete', label: 'Delete File' },
      { key: 'server.file.move', label: 'Move File' },
      { key: 'server.file.copy', label: 'Copy File' },
      { key: 'server.file.compress', label: 'Compress' },
      { key: 'server.file.decompress', label: 'Decompress' },
    ],
  },
  backups: {
    label: 'Backups',
    actions: [
      { key: 'server.backup.create', label: 'Create Backup' },
      { key: 'server.backup.delete', label: 'Delete Backup' },
      { key: 'server.backup.restore', label: 'Restore Backup' },
    ],
  },
  subusers: {
    label: 'Subusers',
    actions: [
      { key: 'server.subuser.add', label: 'Add Subuser' },
      { key: 'server.subuser.update', label: 'Update Subuser' },
      { key: 'server.subuser.remove', label: 'Remove Subuser' },
    ],
  },
  databases: {
    label: 'Databases',
    actions: [
      { key: 'server.database.create', label: 'Create Database' },
      { key: 'server.database.delete', label: 'Delete Database' },
      { key: 'server.database.rotate_password', label: 'Rotate Password' },
    ],
  },
  sftp: {
    label: 'SFTP',
    actions: [
      { key: 'server.sftp.password_reset', label: 'Reset SFTP Password' },
    ],
  },
};

interface User { id: string; username: string; email: string; is_admin: boolean; is_banned: boolean; is_root_admin: boolean; force_password_reset: boolean; ram_limit: number | null; cpu_limit: number | null; disk_limit: number | null; server_limit: number | null; created_at: string; }
type Filter = 'all' | 'admin' | 'banned';
type ActionType = 'ban' | 'unban' | 'delete' | 'setAdmin' | 'revokeAdmin' | 'forceReset';

export default function UsersPage() {
  const table = useTable<User, Filter>({
    mode: 'server',
    fetchFn: async (page, perPage, search, filter) => {
      const res = await adminGetUsers(page, perPage, search, filter);
      return { ...res, data: res.data ? { ...res.data, items: res.data.users } : undefined };
    },
    defaultFilter: 'all',
    itemsKey: 'users',
  });

  const [registrationEnabled, setRegistrationEnabled] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [confirmAction, setConfirmAction] = useState<{ type: ActionType; ids: string[] } | null>(null);
  const [apiKeysUser, setApiKeysUser] = useState<User | null>(null);
  const [showVerification, setShowVerification] = useState(false);
  const [verificationEnabled, setVerificationEnabled] = useState(false);
  const [restrictions, setRestrictions] = useState<string[]>([]);
  const [savingVerification, setSavingVerification] = useState(false);

  useEffect(() => { adminGetRegistrationStatus().then(res => { if (res.success && res.data) setRegistrationEnabled(res.data.enabled); }); }, []);

  useEffect(() => {
    if (showVerification) {
      adminGetEmailVerificationSettings().then(res => {
        if (res.success && res.data) {
          setVerificationEnabled(res.data.enabled);
          setRestrictions(res.data.restrictions || []);
        }
      });
    }
  }, [showVerification]);

  const toggleRegistration = async () => {
    const newVal = !registrationEnabled;
    const res = await adminSetRegistrationStatus(newVal);
    if (res.success) {
      setRegistrationEnabled(newVal);
      notify('Success', newVal ? 'Registration enabled' : 'Registration disabled', 'success');
    } else {
      notify('Error', res.error || 'Failed to update setting', 'error');
    }
  };

  const toggleRestriction = (key: string) => {
    setRestrictions(prev => prev.includes(key) ? prev.filter(r => r !== key) : [...prev, key]);
  };

  const toggleGroup = (actions: { key: string }[]) => {
    const keys = actions.map(a => a.key);
    const allSelected = keys.every(k => restrictions.includes(k));
    if (allSelected) {
      setRestrictions(prev => prev.filter(r => !keys.includes(r)));
    } else {
      setRestrictions(prev => [...new Set([...prev, ...keys])]);
    }
  };

  const saveVerificationSettings = async () => {
    setSavingVerification(true);
    const res = await adminSetEmailVerificationSettings({ enabled: verificationEnabled, restrictions });
    if (res.success) {
      notify('Saved', 'Email verification settings updated', 'success');
    } else {
      notify('Error', res.error || 'Failed to save', 'error');
    }
    setSavingVerification(false);
  };

  if (!table.ready) return null;

  const selectedUsers = table.items.filter(u => table.selected.has(u.id));
  const hasSelectedBanned = selectedUsers.some(u => u.is_banned);
  const hasSelectedUnbanned = selectedUsers.some(u => !u.is_banned);
  const hasSelectedNonAdmin = selectedUsers.some(u => !u.is_admin);
  const hasSelectedRevokableAdmin = selectedUsers.some(u => u.is_admin && !u.is_root_admin);
  const filterLabels = { all: 'All Users', admin: 'Admins', banned: 'Banned' };

  const getUserActions = (user: User) => [
    { label: 'Edit', onClick: () => setEditUser(user) },
    ...(!user.is_root_admin ? [{ label: 'API Keys', onClick: () => setApiKeysUser(user) }] : []),
    ...(!user.is_admin && !user.is_root_admin ? [{ label: 'Set Admin', onClick: () => setConfirmAction({ type: 'setAdmin', ids: [user.id] }) }] : []),
    ...(user.is_admin && !user.is_root_admin ? [{ label: 'Revoke Admin', onClick: () => setConfirmAction({ type: 'revokeAdmin', ids: [user.id] }) }] : []),
    { label: 'Force Password Reset', onClick: () => setConfirmAction({ type: 'forceReset', ids: [user.id] }) },
    'separator' as const,
    ...(user.is_banned
      ? [{ label: 'Unban', onClick: () => setConfirmAction({ type: 'unban', ids: [user.id] }) }]
      : [{ label: 'Ban', onClick: () => setConfirmAction({ type: 'ban', ids: [user.id] }), variant: 'danger' as const }]
    ),
    { label: 'Delete', onClick: () => setConfirmAction({ type: 'delete', ids: [user.id] }), variant: 'danger' as const },
  ];

  const columns = [
    { key: 'select', header: <Checkbox checked={table.allSelected} indeterminate={table.someSelected} onChange={table.toggleSelectAll} />, className: 'w-12', render: (user: User) => <Checkbox checked={table.selected.has(user.id)} onChange={() => table.toggleSelect(user.id)} /> },
    {
      key: 'user', header: 'User', render: (user: User) => (
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center justify-center h-8 w-8 rounded-lg bg-neutral-700 text-sm font-medium text-neutral-200">{user.username?.[0]?.toUpperCase() || '?'}</span>
          <div>
            <div className="text-sm font-medium text-neutral-100">{user.username}</div>
            <div className="text-xs text-neutral-500 font-mono">{user.id}</div>
          </div>
        </div>
      )
    },
    { key: 'email', header: 'Email', render: (user: User) => <span className="text-sm text-neutral-300">{user.email}</span> },
    {
      key: 'status', header: 'Status', render: (user: User) => (
        <div className="flex items-center gap-2">
          {user.is_banned ? (
            <span className="inline-flex items-center rounded-md bg-red-500/10 px-2 py-1 text-xs font-medium text-red-400 ring-1 ring-inset ring-red-500/20">Banned</span>
          ) : user.is_root_admin ? (
            <span className="inline-flex items-center rounded-md bg-rose-500/10 px-2 py-1 text-xs font-medium text-rose-400 ring-1 ring-inset ring-rose-500/20">Root</span>
          ) : user.is_admin ? (
            <span className="inline-flex items-center rounded-md bg-amber-500/10 px-2 py-1 text-xs font-medium text-amber-400 ring-1 ring-inset ring-amber-500/20">Admin</span>
          ) : (
            <span className="inline-flex items-center rounded-md bg-neutral-500/10 px-2 py-1 text-xs font-medium text-neutral-400 ring-1 ring-inset ring-neutral-500/20">User</span>
          )}
          {user.force_password_reset && <span className="inline-flex items-center rounded-md bg-orange-500/10 px-2 py-1 text-xs font-medium text-orange-400 ring-1 ring-inset ring-orange-500/20">Reset</span>}
        </div>
      )
    },
    { key: 'joined', header: 'Joined', render: (user: User) => <span className="text-sm text-neutral-400">{new Date(user.created_at).toLocaleDateString()}</span> },
    {
      key: 'actions', header: '', align: 'right' as const, render: (user: User) => (
        <ContextMenu
          align="end"
          trigger={<Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>}
          items={getUserActions(user)}
        />
      )
    },
  ];

  return (
    <>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold text-neutral-100">Users</h1>
            <p className="text-sm text-neutral-400">Manage all users across the platform.</p>
          </div>
          <Button onClick={() => setShowCreate(true)} className="w-full sm:w-auto"><Icons.plus className="w-4 h-4" />Create User</Button>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
            <form onSubmit={table.handleSearch} className="flex-1 sm:max-w-sm">
              <Input placeholder="Search by name or email..." value={table.searchInput} onChange={e => table.setSearchInput(e.target.value)} />
            </form>
            <div className="flex items-center gap-3">
              <ContextMenu
                align="start"
                trigger={
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    <Icons.filter className="w-4 h-4 text-neutral-400" />
                    {filterLabels[table.filter]}
                  </button>
                }
                items={[
                  { label: 'All Users', onClick: () => table.handleFilterChange('all') },
                  { label: 'Admins', onClick: () => table.handleFilterChange('admin') },
                  { label: 'Banned', onClick: () => table.handleFilterChange('banned') },
                ]}
              />
              <ContextMenu
                align="start"
                trigger={
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    Registration: {registrationEnabled ? 'On' : 'Off'}
                  </button>
                }
                items={[
                  { label: 'On', onClick: () => !registrationEnabled && toggleRegistration() },
                  { label: 'Off', onClick: () => registrationEnabled && toggleRegistration() },
                ]}
              />
              <button
                type="button"
                onClick={() => setShowVerification(true)}
                className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2"
              >
                <Icons.mail className="w-4 h-4 text-neutral-400" />
                Email Verification
              </button>
            </div>
          </div>
          <Pagination page={table.page} totalPages={table.totalPages} total={table.total} perPage={table.perPage} perPageOptions={(() => {
            const options = [10, 20, 50, 100];
            const steps = [250, 500, 1000, 2500, 5000, 10000];
            for (const step of steps) {
              if (table.total > step) options.push(step);
            }
            if (table.total > 100 && !options.includes(table.total)) {
              options.push(table.total);
            }
            return options;
          })()} onPageChange={table.setPage} onPerPageChange={table.handlePerPageChange} loading={table.loading} />
        </div>

        <BulkActionBar count={table.selected.size} onClear={table.clearSelection}>
          {hasSelectedUnbanned && <button onClick={() => setConfirmAction({ type: 'ban', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors">Ban</button>}
          {hasSelectedBanned && <button onClick={() => setConfirmAction({ type: 'unban', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-green-400 hover:bg-green-500/10 transition-colors">Unban</button>}
          {hasSelectedNonAdmin && <button onClick={() => setConfirmAction({ type: 'setAdmin', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-amber-400 hover:bg-amber-500/10 transition-colors">Set Admin</button>}
          {hasSelectedRevokableAdmin && <button onClick={() => setConfirmAction({ type: 'revokeAdmin', ids: selectedUsers.filter(u => u.is_admin && !u.is_root_admin).map(u => u.id) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-amber-400 hover:bg-amber-500/10 transition-colors">Revoke Admin</button>}
          <button onClick={() => setConfirmAction({ type: 'forceReset', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-orange-400 hover:bg-orange-500/10 transition-colors">Force Reset</button>
          <button onClick={() => setConfirmAction({ type: 'delete', ids: Array.from(table.selected) })} className="text-xs font-medium px-3 py-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors">Delete</button>
        </BulkActionBar>

        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-4 py-2 text-xs text-neutral-400">{table.total} users</div>
          <div className="bg-neutral-900/40 rounded-lg p-1">
            <Table columns={columns} data={table.items} keyField="id" loading={table.loading} emptyText="No users found" rowClassName={user => table.selected.has(user.id) ? 'bg-neutral-800/20' : ''} contextMenu={getUserActions} />
          </div>
        </div>
      </div>

      <CreateUserModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={table.reload} />
      <EditUserModal open={!!editUser} user={editUser} onClose={() => setEditUser(null)} onSaved={table.reload} />
      <UserActionModal open={!!confirmAction} type={confirmAction?.type || 'ban'} ids={confirmAction?.ids || []} onClose={() => setConfirmAction(null)} onComplete={() => { table.reload(); table.clearSelection(); }} />
      <UserAPIKeysModal open={!!apiKeysUser} user={apiKeysUser} onClose={() => setApiKeysUser(null)} />

      <SlidePanel
        open={showVerification}
        onClose={() => setShowVerification(false)}
        title="Email Verification"
        description="Require users to verify their email before performing certain actions. Admins are always exempt."
        width="max-w-xl"
        footer={
          <div className="flex items-center justify-end gap-3">
            <Button variant="ghost" onClick={() => setShowVerification(false)}>Cancel</Button>
            <Button onClick={saveVerificationSettings} loading={savingVerification}>Save</Button>
          </div>
        }
      >
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-neutral-100">Enable Email Verification</div>
              <div className="text-xs text-neutral-500 mt-0.5">When enabled, verification emails are sent on registration</div>
            </div>
            <button
              type="button"
              onClick={() => setVerificationEnabled(!verificationEnabled)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors cursor-pointer ${verificationEnabled ? 'bg-white' : 'bg-neutral-700'}`}
            >
              <span className={`inline-block h-4 w-4 rounded-full transition-transform ${verificationEnabled ? 'translate-x-6 bg-neutral-900' : 'translate-x-1 bg-neutral-400'}`} />
            </button>
          </div>

          {verificationEnabled && (
            <>
              <div className="border-t border-neutral-800 pt-4">
                <div className="text-xs font-medium text-neutral-400 uppercase tracking-wider mb-4">Restricted Actions</div>
                <p className="text-xs text-neutral-500 mb-4">Unverified users will be blocked from performing the checked actions below.</p>
              </div>

              <div className="space-y-5">
                {Object.entries(VerificationActionGroups).map(([groupKey, group]) => {
                  const allSelected = group.actions.every(a => restrictions.includes(a.key));
                  const someSelected = group.actions.some(a => restrictions.includes(a.key));
                  return (
                    <div key={groupKey} className="space-y-2">
                      <div className="flex items-center gap-2">
                        <Checkbox checked={allSelected} indeterminate={someSelected && !allSelected} onChange={() => toggleGroup(group.actions)} />
                        <span className="text-xs font-medium text-neutral-300 uppercase tracking-wider">{group.label}</span>
                      </div>
                      <div className="grid grid-cols-2 gap-2 pl-6">
                        {group.actions.map(action => (
                          <div
                            key={action.key}
                            onClick={() => toggleRestriction(action.key)}
                            className="flex items-center gap-2 cursor-pointer group"
                          >
                            <div onClick={e => e.stopPropagation()}>
                              <Checkbox checked={restrictions.includes(action.key)} onChange={() => toggleRestriction(action.key)} />
                            </div>
                            <span className="text-xs text-neutral-400 group-hover:text-neutral-200 transition">{action.label}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </div>
      </SlidePanel>
    </>
  );
}
