import { useEffect, useState } from 'react';
import { getUser, setUser } from '../../lib/auth';
import { updateProfile, updatePassword, getSessions, revokeSession, revokeAllSessions, getAPIKeys, createAPIKey, deleteAPIKey, type APIKey, type APIKeyCreated } from '../../lib/api';
import { formatDate, parseUserAgent } from '../../lib/utils';
import { notify, Input, Button, Icons, Modal, Table } from '../../components';
import { getPluginTabs, evaluatePluginGuard } from '../../lib/pluginLoader';
import { PluginRenderer } from '../../components/plugins';

type Tab = 'account' | 'sessions' | 'api-keys' | string;

interface Session { id: string; ip: string; user_agent: string; created_at: string; expires_at: string; is_current: boolean; }

interface PluginTabInfo {
  pluginId: string;
  id: string;
  label: string;
  component: string;
}

export default function SettingsPage() {
  const [tab, setTab] = useState<Tab>('account');
  const user = getUser();

  const pluginTabs = getPluginTabs('user-settings')
    .filter(({ pluginId, tab }) => evaluatePluginGuard(pluginId, tab.guard))
    .map(({ pluginId, tab }): PluginTabInfo => ({
      pluginId,
      id: `plugin-${pluginId}-${tab.id}`,
      label: tab.label,
      component: tab.component,
    }));

  const [profile, setProfile] = useState({ username: user?.username || '', email: user?.email || '', loading: false });
  const [password, setPassword] = useState({ current: '', new: '', confirm: '', loading: false });
  const [sessionsState, setSessionsState] = useState<{ list: Session[]; loading: boolean }>({ list: [], loading: false });

  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [apiKeysLoading, setApiKeysLoading] = useState(false);
  const [createModal, setCreateModal] = useState({ open: false, loading: false, name: '', expiresIn: '' });
  const [newKey, setNewKey] = useState<APIKeyCreated | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ key: APIKey; loading: boolean } | null>(null);

  const loadSessions = async () => {
    setSessionsState(s => ({ ...s, loading: true }));
    const res = await getSessions();
    setSessionsState({ list: res.success && res.data ? res.data : [], loading: false });
  };

  const loadApiKeys = async () => {
    setApiKeysLoading(true);
    const res = await getAPIKeys();
    if (res.success && res.data) setApiKeys(res.data);
    setApiKeysLoading(false);
  };

  useEffect(() => {
    if (tab === 'sessions') loadSessions();
    if (tab === 'api-keys') loadApiKeys();
  }, [tab]);

  const handleProfileSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfile(p => ({ ...p, loading: true }));
    const res = await updateProfile(profile.username, profile.email);
    if (res.success && res.data) {
      setUser({ ...user!, username: res.data.username, email: res.data.email });
      notify('Profile updated', 'Your profile has been saved', 'success');
    } else notify('Error', res.error || 'Failed to update profile', 'error');
    setProfile(p => ({ ...p, loading: false }));
  };

  const handlePasswordSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password.new !== password.confirm) { notify('Error', 'Passwords do not match', 'error'); return; }
    if (password.new.length < 8) { notify('Error', 'Password must be at least 8 characters', 'error'); return; }
    setPassword(p => ({ ...p, loading: true }));
    const res = await updatePassword(password.current, password.new);
    if (res.success) { setPassword({ current: '', new: '', confirm: '', loading: false }); notify('Password updated', 'Your password has been changed', 'success'); }
    else { notify('Error', res.error || 'Failed to update password', 'error'); setPassword(p => ({ ...p, loading: false })); }
  };

  const handleRevokeSession = async (sessionId: string) => {
    const res = await revokeSession(sessionId);
    if (res.success) { setSessionsState(s => ({ ...s, list: s.list.filter(x => x.id !== sessionId) })); notify('Session revoked', 'The session has been terminated', 'success'); }
    else notify('Error', res.error || 'Failed to revoke session', 'error');
  };

  const handleRevokeAll = async () => {
    const res = await revokeAllSessions();
    if (res.success) { loadSessions(); notify('Sessions revoked', 'All other sessions have been terminated', 'success'); }
    else notify('Error', res.error || 'Failed to revoke sessions', 'error');
  };
  void handleRevokeAll;

  const handleCreateApiKey = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateModal(m => ({ ...m, loading: true }));
    const expiresIn = createModal.expiresIn ? parseInt(createModal.expiresIn) : undefined;
    const res = await createAPIKey(createModal.name || 'API Key', expiresIn);
    if (res.success && res.data) {
      setNewKey(res.data);
      setCreateModal({ open: false, loading: false, name: '', expiresIn: '' });
      loadApiKeys();
    } else {
      notify('Error', res.error || 'Failed to create API key', 'error');
      setCreateModal(m => ({ ...m, loading: false }));
    }
  };

  const handleDeleteApiKey = async () => {
    if (!deleteModal) return;
    setDeleteModal(m => m && { ...m, loading: true });
    const res = await deleteAPIKey(deleteModal.key.id);
    if (res.success) {
      notify('Deleted', 'API key deleted', 'success');
      setDeleteModal(null);
      loadApiKeys();
    } else {
      notify('Error', res.error || 'Failed to delete', 'error');
      setDeleteModal(m => m && { ...m, loading: false });
    }
  };

  const copyKey = () => {
    if (newKey) {
      navigator.clipboard.writeText(newKey.key);
      notify('Copied', 'API key copied to clipboard', 'success');
    }
  };

  const apiKeyColumns = [
    { key: 'name', header: 'Name', render: (k: APIKey) => <span className="text-sm font-medium text-neutral-100">{k.name}</span> },
    { key: 'prefix', header: 'Key', render: (k: APIKey) => <span className="text-sm font-mono text-neutral-400">{k.key_prefix}...</span> },
    { key: 'expires', header: 'Expires', render: (k: APIKey) => <span className="text-sm text-neutral-400">{k.expires_at ? new Date(k.expires_at).toLocaleDateString() : 'Never'}</span> },
    { key: 'used', header: 'Last Used', render: (k: APIKey) => <span className="text-sm text-neutral-400">{k.last_used_at ? new Date(k.last_used_at).toLocaleString() : 'Never'}</span> },
    { key: 'created', header: 'Created', render: (k: APIKey) => <span className="text-sm text-neutral-400">{new Date(k.created_at).toLocaleDateString()}</span> },
    {
      key: 'actions', header: '', align: 'right' as const, render: (k: APIKey) => (
        <button onClick={() => setDeleteModal({ key: k, loading: false })} className="text-xs text-red-400 hover:text-red-300">Delete</button>
      )
    },
  ];

  const tabs = [
    { id: 'account', label: 'Account' },
    { id: 'sessions', label: 'Sessions' },
    { id: 'api-keys', label: 'API Keys' },
    ...pluginTabs.map(t => ({ id: t.id, label: t.label })),
  ];

  const activePluginTab = pluginTabs.find(t => t.id === tab);

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="flex items-center gap-4">
        <div className="flex items-center justify-center rounded-2xl overflow-hidden border border-neutral-700/50" style={{ width: 64, height: 64 }}>
          <div className="w-full h-full bg-neutral-700 flex items-center justify-center text-2xl font-semibold text-neutral-200">
            {user?.username?.[0]?.toUpperCase() || 'U'}
          </div>
        </div>
        <div>
          <h1 className="text-2xl font-semibold text-neutral-100">{user?.username || 'User'}</h1>
          <p className="text-sm text-neutral-400 mt-0.5">{user?.email || 'user@example.com'}</p>
        </div>
      </div>

      <div className="border-b border-neutral-800">
        <div className="flex gap-1">
          {tabs.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setTab(t.id as Tab)}
              className={`px-4 py-2.5 text-sm font-medium transition-colors relative ${tab === t.id ? 'text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'
                }`}
            >
              {t.label}
              {tab === t.id && (
                <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-neutral-100 rounded-full" />
              )}
            </button>
          ))}
        </div>
      </div>

      {tab === 'account' && (
        <div className="space-y-6">
          <div className="rounded-xl bg-neutral-800/30">
            <div className="px-6 pt-6 pb-3">
              <h3 className="text-lg font-semibold text-neutral-100">Profile</h3>
              <p className="mt-1 text-sm text-neutral-400">Update your username and email address.</p>
            </div>
            <div className="px-6 pb-6 pt-2">
              <form onSubmit={handleProfileSave} className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Input label="Username" value={profile.username} onChange={e => setProfile(p => ({ ...p, username: e.target.value }))} />
                  <Input label="Email address" type="email" value={profile.email} onChange={e => setProfile(p => ({ ...p, email: e.target.value }))} />
                </div>
                <div className="flex justify-end pt-2">
                  <Button type="submit" loading={profile.loading}>Save Changes</Button>
                </div>
              </form>
            </div>
          </div>

          <div className="rounded-xl bg-neutral-800/30">
            <div className="px-6 pt-6 pb-3">
              <h3 className="text-lg font-semibold text-neutral-100">Password</h3>
              <p className="mt-1 text-sm text-neutral-400">Change your account password.</p>
            </div>
            <div className="px-6 pb-6 pt-2">
              <form onSubmit={handlePasswordSave} className="space-y-4">
                <Input label="Current password" value={password.current} onChange={e => setPassword(p => ({ ...p, current: e.target.value }))} hideable />
                <Input label="New password" value={password.new} onChange={e => setPassword(p => ({ ...p, new: e.target.value }))} hideable />
                <Input label="Confirm new password" value={password.confirm} onChange={e => setPassword(p => ({ ...p, confirm: e.target.value }))} hideable />
                <div className="flex justify-end pt-2">
                  <Button type="submit" loading={password.loading}>Update Password</Button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {tab === 'sessions' && (
        <div className="rounded-xl bg-neutral-800/30">
          <div className="px-6 pt-6 pb-3">
            <h3 className="text-lg font-semibold text-neutral-100 tracking-tight">Active sessions</h3>
            <p className="text-sm text-neutral-400">Manage and monitor your active sessions across all devices.</p>
          </div>
          <div className="px-6 pb-6 pt-2">
            {sessionsState.loading ? (
              <p className="text-sm text-neutral-500">Loading sessions...</p>
            ) : sessionsState.list.length === 0 ? (
              <p className="text-sm text-neutral-500">No active sessions.</p>
            ) : (
              <div className="space-y-4">
                {sessionsState.list.map((session) => (
                  <div key={session.id} className="relative bg-neutral-800 border border-transparent rounded-xl pt-5 p-4 hover:border-neutral-600 transition-colors overflow-hidden">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-start space-x-4 min-w-0 flex-1">
                        <div className="flex-shrink-0 hidden sm:block">
                          <div className="w-10 h-10 bg-neutral-700 rounded-lg flex items-center justify-center">
                            <Icons.monitor className="h-4 w-4 text-neutral-100" />
                          </div>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center flex-wrap gap-2">
                            <h3 className="text-sm font-semibold text-neutral-100 truncate">{parseUserAgent(session.user_agent)}</h3>
                            {session.is_current && (
                              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-900/50 text-emerald-300 flex-shrink-0">
                                <div className="w-1.5 h-1.5 bg-emerald-400 rounded-full mr-1.5"></div>
                                This session
                              </span>
                            )}
                          </div>
                          <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-neutral-400">
                            <div className="flex items-center space-x-1">
                              <Icons.globe className="w-4 h-4 flex-shrink-0" />
                              <span className="truncate">{session.ip || 'Unknown IP'}</span>
                            </div>
                            <div className="flex items-center space-x-1">
                              <Icons.clock className="w-4 h-4 flex-shrink-0" />
                              <span className="truncate">Started {formatDate(session.created_at)}</span>
                            </div>
                          </div>
                          <p className="mt-2 text-xs text-neutral-500 truncate">{session.user_agent}</p>
                        </div>
                      </div>
                      <div className="flex-shrink-0">
                        {!session.is_current && (
                          <button
                            onClick={() => handleRevokeSession(session.id)}
                            className="rounded-lg px-2.5 py-1.5 text-xs font-semibold text-red-400 bg-red-500/10 hover:bg-red-500/20 transition-all"
                          >
                            Revoke
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {tab === 'api-keys' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-neutral-100">API Keys</h3>
              <p className="text-sm text-neutral-400">Manage your API keys for external access.</p>
            </div>
            <Button onClick={() => setCreateModal(m => ({ ...m, open: true }))}><Icons.plus className="w-4 h-4" />Create Key</Button>
          </div>

          <div className="rounded-xl bg-neutral-800/30">
            <div className="px-4 py-2 text-xs text-neutral-400">{apiKeys.length} API key{apiKeys.length !== 1 ? 's' : ''}</div>
            <div className="bg-neutral-900/40 rounded-lg p-1">
              <Table columns={apiKeyColumns} data={apiKeys} keyField="id" loading={apiKeysLoading} emptyText="No API keys yet" />
            </div>
          </div>
        </div>
      )}

      {activePluginTab && (
        <PluginRenderer
          pluginId={activePluginTab.pluginId}
          component={activePluginTab.component}
          props={{ user }}
        />
      )}

      <Modal open={createModal.open} onClose={() => !createModal.loading && setCreateModal(m => ({ ...m, open: false }))} title="Create API Key" description="Generate a new API key for external access.">
        <form onSubmit={handleCreateApiKey} className="space-y-4">
          <Input label="Name" placeholder="My CLI Key" value={createModal.name} onChange={e => setCreateModal(m => ({ ...m, name: e.target.value }))} />
          <Input label="Expires In (days)" placeholder="Leave blank for never" type="number" min={1} value={createModal.expiresIn} onChange={e => setCreateModal(m => ({ ...m, expiresIn: e.target.value }))} />
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setCreateModal(m => ({ ...m, open: false }))} disabled={createModal.loading}>Cancel</Button>
            <Button type="submit" loading={createModal.loading}>Create</Button>
          </div>
        </form>
      </Modal>

      <Modal open={!!newKey} onClose={() => setNewKey(null)} title="API Key Created" description="Copy your API key now. You won't be able to see it again!">
        <div className="space-y-4">
          <div className="p-3 bg-neutral-900 rounded-lg font-mono text-sm text-emerald-400 break-all">{newKey?.key}</div>
          <div className="flex justify-end gap-3">
            <Button onClick={copyKey}><Icons.clipboard className="w-4 h-4" />Copy</Button>
            <Button variant="ghost" onClick={() => setNewKey(null)}>Done</Button>
          </div>
        </div>
      </Modal>

      <Modal open={!!deleteModal} onClose={() => !deleteModal?.loading && setDeleteModal(null)} title="Delete API Key" description={`Delete "${deleteModal?.key.name}"? This cannot be undone.`}>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={() => setDeleteModal(null)} disabled={deleteModal?.loading}>Cancel</Button>
          <Button onClick={handleDeleteApiKey} loading={deleteModal?.loading} variant="danger">Delete</Button>
        </div>
      </Modal>

    </div>
  );
}
