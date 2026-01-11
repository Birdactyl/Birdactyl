import { useState, useEffect } from 'react';
import { adminGetUsers, adminGetRegistrationStatus, adminSetRegistrationStatus } from '../../../lib/api';
import { useTable } from '../../../hooks/useTable';
import { notify, Button, Input, Pagination, Checkbox, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, BulkActionBar, Table } from '../../../components';
import { CreateUserModal, EditUserModal, UserActionModal, UserAPIKeysModal } from '../../../components/modals';

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

  useEffect(() => { adminGetRegistrationStatus().then(res => { if (res.success && res.data) setRegistrationEnabled(res.data.enabled); }); }, []);

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

  if (!table.ready) return null;

  const selectedUsers = table.items.filter(u => table.selected.has(u.id));
  const hasSelectedBanned = selectedUsers.some(u => u.is_banned);
  const hasSelectedUnbanned = selectedUsers.some(u => !u.is_banned);
  const hasSelectedNonAdmin = selectedUsers.some(u => !u.is_admin);
  const hasSelectedRevokableAdmin = selectedUsers.some(u => u.is_admin && !u.is_root_admin);
  const filterLabels = { all: 'All Users', admin: 'Admins', banned: 'Banned' };

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
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onSelect={() => setEditUser(user)}>Edit</DropdownMenuItem>
            {!user.is_root_admin && <DropdownMenuItem onSelect={() => setApiKeysUser(user)}>API Keys</DropdownMenuItem>}
            <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'forceReset', ids: [user.id] })}>Force Password Reset</DropdownMenuItem>
            {user.is_banned ? (
              <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'unban', ids: [user.id] })}>Unban</DropdownMenuItem>
            ) : (
              <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'ban', ids: [user.id] })} className="text-red-400">Ban</DropdownMenuItem>
            )}
            <DropdownMenuItem onSelect={() => setConfirmAction({ type: 'delete', ids: [user.id] })} className="text-red-400">Delete</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    <Icons.filter className="w-4 h-4 text-neutral-400" />
                    {filterLabels[table.filter]}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('all')}>All Users</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('admin')}>Admins</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => table.handleFilterChange('banned')}>Banned</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button type="button" className="rounded-lg border border-neutral-800 px-3 py-2 text-xs text-neutral-100 transition hover:border-neutral-500 focus:outline-none bg-neutral-800/80 flex items-center gap-2">
                    Registration: {registrationEnabled ? 'On' : 'Off'}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onSelect={() => !registrationEnabled && toggleRegistration()}>On</DropdownMenuItem>
                  <DropdownMenuItem onSelect={() => registrationEnabled && toggleRegistration()}>Off</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
          <Pagination page={table.page} totalPages={table.totalPages} total={table.total} perPage={table.perPage} onPageChange={table.setPage} onPerPageChange={table.handlePerPageChange} loading={table.loading} />
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
            <Table columns={columns} data={table.items} keyField="id" loading={table.loading} emptyText="No users found" rowClassName={user => table.selected.has(user.id) ? 'bg-neutral-800/20' : ''} />
          </div>
        </div>
      </div>

      <CreateUserModal open={showCreate} onClose={() => setShowCreate(false)} onCreated={table.reload} />
      <EditUserModal open={!!editUser} user={editUser} onClose={() => setEditUser(null)} onSaved={table.reload} />
      <UserActionModal open={!!confirmAction} type={confirmAction?.type || 'ban'} ids={confirmAction?.ids || []} onClose={() => setConfirmAction(null)} onComplete={() => { table.reload(); table.clearSelection(); }} />
      <UserAPIKeysModal open={!!apiKeysUser} user={apiKeysUser} onClose={() => setApiKeysUser(null)} />
    </>
  );
}
