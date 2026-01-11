import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Icons, Table, PermissionDenied, CreateDatabaseModal, DeleteDatabaseModal, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../../../components';
import { notify } from '../../../components/feedback/Notification';
import { getServerDatabases, rotateDatabasePassword, ServerDatabase } from '../../../lib/api';

export default function DatabasesPage() {
  const { id } = useParams<{ id: string }>();
  const { can, loading: permsLoading } = useServerPermissions(id);
  const [databases, setDatabases] = useState<ServerDatabase[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModal, setCreateModal] = useState(false);
  const [deleteModal, setDeleteModal] = useState<ServerDatabase | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [visiblePasswords, setVisiblePasswords] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (!id) return;
    getServerDatabases(id).then(res => {
      if (res.success && res.data) setDatabases(res.data);
      setLoading(false);
    });
  }, [id]);

  const handleRotate = async (db: ServerDatabase) => {
    if (!id) return;
    const res = await rotateDatabasePassword(id, db.id);
    if (res.success && res.data) {
      setDatabases(prev => prev.map(d => d.id === db.id ? res.data! : d));
      notify('Success', 'Password rotated', 'success');
    } else {
      notify('Error', res.error || 'Failed to rotate password', 'error');
    }
  };

  const toggleExpand = (dbId: string) => setExpanded(prev => { const next = new Set(prev); next.has(dbId) ? next.delete(dbId) : next.add(dbId); return next; });
  const togglePassword = (dbId: string) => setVisiblePasswords(prev => { const next = new Set(prev); next.has(dbId) ? next.delete(dbId) : next.add(dbId); return next; });
  const copy = (text: string, label: string) => { navigator.clipboard.writeText(text); notify('Copied', `${label} copied to clipboard`, 'success'); };

  if (permsLoading || loading) return null;
  if (!can('database.view')) return <PermissionDenied message="You don't have permission to view databases" />;

  const columns = [
    { key: 'expand', header: '', className: 'w-8', render: (db: ServerDatabase) => (
      <button onClick={() => toggleExpand(db.id)} className="text-neutral-400 hover:text-neutral-200 transition">
        <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(db.id) ? 'rotate-90' : ''}`} />
      </button>
    )},
    { key: 'database', header: 'Database', render: (db: ServerDatabase) => (
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-violet-500/20">
          <Icons.database className="w-4 h-4 text-violet-400" />
        </div>
        <div>
          <div className="text-sm font-medium text-neutral-100">{db.database_name}</div>
          <div className="text-xs text-neutral-500">{db.host}:{db.port}</div>
        </div>
      </div>
    )},
    { key: 'username', header: 'Username', render: (db: ServerDatabase) => (
      <div className="flex items-center gap-2">
        <code className="text-sm text-neutral-300 font-mono">{db.username}</code>
        <button onClick={() => copy(db.username, 'Username')} className="text-neutral-500 hover:text-neutral-300">
          <Icons.copy className="w-3.5 h-3.5" />
        </button>
      </div>
    )},
    { key: 'password', header: 'Password', render: (db: ServerDatabase) => (
      <div className="flex items-center gap-2">
        <code className="text-sm text-neutral-300 font-mono">{visiblePasswords.has(db.id) ? db.password : '••••••••••••'}</code>
        <button onClick={() => togglePassword(db.id)} className="text-neutral-500 hover:text-neutral-300">
          {visiblePasswords.has(db.id) ? <Icons.eyeOff className="w-3.5 h-3.5" /> : <Icons.eye className="w-3.5 h-3.5" />}
        </button>
        <button onClick={() => copy(db.password, 'Password')} className="text-neutral-500 hover:text-neutral-300">
          <Icons.copy className="w-3.5 h-3.5" />
        </button>
      </div>
    )},
    { key: 'actions', header: '', align: 'right' as const, render: (db: ServerDatabase) => (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {can('database.update') && <DropdownMenuItem onSelect={() => handleRotate(db)}>Rotate Password</DropdownMenuItem>}
          {can('database.delete') && <DropdownMenuItem onSelect={() => setDeleteModal(db)} className="text-red-400">Delete</DropdownMenuItem>}
        </DropdownMenuContent>
      </DropdownMenu>
    )},
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">Server</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">Databases</span>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-100">Databases</h1>
          <p className="text-sm text-neutral-400">Manage MySQL databases for this server.</p>
        </div>
        {can('database.create') && <Button onClick={() => setCreateModal(true)}><Icons.plus className="w-4 h-4" />New Database</Button>}
      </div>

      <div className="rounded-xl bg-neutral-800/30">
        <div className="px-4 py-2 text-xs text-neutral-400">{databases.length} database{databases.length !== 1 ? 's' : ''}</div>
        <div className="bg-neutral-900/40 rounded-lg p-1">
          <Table
            columns={columns}
            data={databases}
            keyField="id"
            emptyText="No databases created yet"
            expandable={{
              isExpanded: db => expanded.has(db.id),
              render: db => (
                <div className="grid grid-cols-2 gap-6 text-xs">
                  <div>
                    <div className="text-neutral-500 mb-1">Connection String</div>
                    <div className="flex items-center gap-2">
                      <code className="text-neutral-200 font-mono break-all">mysql://{db.username}:{visiblePasswords.has(db.id) ? db.password : '********'}@{db.host}:{db.port}/{db.database_name}</code>
                      <button onClick={() => copy(`mysql://${db.username}:${db.password}@${db.host}:${db.port}/${db.database_name}`, 'Connection string')} className="text-neutral-500 hover:text-neutral-300 flex-shrink-0">
                        <Icons.copy className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                  <div>
                    <div className="text-neutral-500 mb-1">JDBC URL</div>
                    <div className="flex items-center gap-2">
                      <code className="text-neutral-200 font-mono break-all">jdbc:mysql://{db.host}:{db.port}/{db.database_name}</code>
                      <button onClick={() => copy(`jdbc:mysql://${db.host}:${db.port}/${db.database_name}`, 'JDBC URL')} className="text-neutral-500 hover:text-neutral-300 flex-shrink-0">
                        <Icons.copy className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                </div>
              )
            }}
          />
        </div>
      </div>

      <CreateDatabaseModal serverId={id!} open={createModal} onClose={() => setCreateModal(false)} onCreated={db => setDatabases(prev => [...prev, db])} />
      <DeleteDatabaseModal serverId={id!} database={deleteModal} open={!!deleteModal} onClose={() => setDeleteModal(null)} onDeleted={dbId => setDatabases(prev => prev.filter(d => d.id !== dbId))} />
    </div>
  );
}
