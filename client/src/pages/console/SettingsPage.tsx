import { useEffect, useState } from 'react';
import { Routes, Route } from 'react-router-dom';
import { getUser, setUser } from '../../lib/auth';
import { updateProfile, updatePassword, getSessions, revokeSession, revokeAllSessions, getAPIKeys, createAPIKey, deleteAPIKey, setup2FA, enable2FA, disable2FA, regenerateBackupCodes, type APIKey, type APIKeyCreated } from '../../lib/api';
import { formatDate, parseUserAgent } from '../../lib/utils';
import { notify, Input, Button, Icons, Modal, SlidePanel } from '../../components';
import { SubNavigation } from '../../components/layout/SubNavigation';
import { getPluginTabs, evaluatePluginGuard } from '../../lib/pluginLoader';
import { PluginRenderer } from '../../components/plugins';

interface Session { id: string; ip: string; user_agent: string; created_at: string; expires_at: string; is_current: boolean; }

interface PluginTabInfo {
  pluginId: string;
  id: string;
  label: string;
  icon: string;
  component: string;
}

const settingsTabs = [
  { name: 'Account', path: '', icon: 'users' },
  { name: 'Security', path: '/security', icon: 'shield' },
  { name: 'Sessions', path: '/sessions', icon: 'monitor' },
  { name: 'API Keys', path: '/api-keys', icon: 'key' },
];

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

function AccountTab() {
  const user = getUser();
  const [profile, setProfile] = useState({ username: user?.username || '', email: user?.email || '', loading: false });

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

  return (
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Profile"
        description="Your public display name and contact email."
        footer={<Button type="submit" form="profileForm" loading={profile.loading}>Save Changes</Button>}
      >
        <form id="profileForm" onSubmit={handleProfileSave} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input label="Username" value={profile.username} onChange={e => setProfile(p => ({ ...p, username: e.target.value }))} />
          <Input label="Email address" type="email" value={profile.email} onChange={e => setProfile(p => ({ ...p, email: e.target.value }))} />
        </form>
      </SectionCard>
    </div>
  );
}

function generateQRCodeSVG(url: string): string {
  const data = encodeURIComponent(url);
  return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${data}&bgcolor=0a0a0a&color=ffffff&format=svg`;
}

function SecurityTab() {
  const user = getUser();
  const [password, setPassword] = useState({ current: '', new: '', confirm: '', loading: false });
  const [totpEnabled, setTotpEnabled] = useState(user?.totp_enabled || false);
  const [setupData, setSetupData] = useState<{ secret: string; url: string } | null>(null);
  const [setupCode, setSetupCode] = useState('');
  const [setupLoading, setSetupLoading] = useState(false);
  const [backupCodes, setBackupCodes] = useState<string[] | null>(null);
  const [disablePassword, setDisablePassword] = useState('');
  const [disableLoading, setDisableLoading] = useState(false);
  const [regenPassword, setRegenPassword] = useState('');
  const [regenLoading, setRegenLoading] = useState(false);
  const [showDisable, setShowDisable] = useState(false);
  const [showRegen, setShowRegen] = useState(false);

  const handlePasswordSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password.new !== password.confirm) { notify('Error', 'Passwords do not match', 'error'); return; }
    if (password.new.length < 8) { notify('Error', 'Password must be at least 8 characters', 'error'); return; }
    setPassword(p => ({ ...p, loading: true }));
    const res = await updatePassword(password.current, password.new);
    if (res.success) { setPassword({ current: '', new: '', confirm: '', loading: false }); notify('Password updated', 'Your password has been changed', 'success'); }
    else { notify('Error', res.error || 'Failed to update password', 'error'); setPassword(p => ({ ...p, loading: false })); }
  };

  const handleSetup = async () => {
    setSetupLoading(true);
    const res = await setup2FA();
    if (res.success && res.data) {
      setSetupData(res.data);
    } else {
      notify('Error', res.error || 'Failed to set up 2FA', 'error');
    }
    setSetupLoading(false);
  };

  const handleEnable = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!setupCode) return;
    setSetupLoading(true);
    const res = await enable2FA(setupCode);
    if (res.success && res.data) {
      setTotpEnabled(true);
      setSetupData(null);
      setSetupCode('');
      setBackupCodes(res.data.backup_codes);
      const currentUser = getUser();
      if (currentUser) setUser({ ...currentUser, totp_enabled: true });
      notify('2FA enabled', 'Two-factor authentication has been enabled', 'success');
    } else {
      notify('Error', res.error || 'Invalid code', 'error');
    }
    setSetupLoading(false);
  };

  const handleDisable = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!disablePassword) return;
    setDisableLoading(true);
    const res = await disable2FA(disablePassword);
    if (res.success) {
      setTotpEnabled(false);
      setShowDisable(false);
      setDisablePassword('');
      const currentUser = getUser();
      if (currentUser) setUser({ ...currentUser, totp_enabled: false });
      notify('2FA disabled', 'Two-factor authentication has been disabled', 'success');
    } else {
      notify('Error', res.error || 'Failed to disable 2FA', 'error');
    }
    setDisableLoading(false);
  };

  const handleRegenerate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!regenPassword) return;
    setRegenLoading(true);
    const res = await regenerateBackupCodes(regenPassword);
    if (res.success && res.data) {
      setBackupCodes(res.data.backup_codes);
      setShowRegen(false);
      setRegenPassword('');
      notify('Backup codes regenerated', 'Your old backup codes are no longer valid', 'success');
    } else {
      notify('Error', res.error || 'Failed to regenerate codes', 'error');
    }
    setRegenLoading(false);
  };

  const copyBackupCodes = () => {
    if (backupCodes) {
      navigator.clipboard.writeText(backupCodes.join('\n'));
      notify('Copied', 'Backup codes copied to clipboard', 'success');
    }
  };

  return (
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Change Password"
        description="Use a strong password that you don't use on any other sites."
        footer={<Button type="submit" form="passwordForm" loading={password.loading}>Update Password</Button>}
      >
        <form id="passwordForm" onSubmit={handlePasswordSave} className="space-y-4 max-w-md">
          <Input label="Current password" value={password.current} onChange={e => setPassword(p => ({ ...p, current: e.target.value }))} hideable />
          <Input label="New password" value={password.new} onChange={e => setPassword(p => ({ ...p, new: e.target.value }))} hideable />
          <Input label="Confirm new password" value={password.confirm} onChange={e => setPassword(p => ({ ...p, confirm: e.target.value }))} hideable />
        </form>
      </SectionCard>

      <SectionCard
        title="Two-Factor Authentication"
        description="Add an extra layer of security to your account."
      >
        {totpEnabled ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 rounded-lg border border-emerald-500/20 bg-emerald-500/[0.03]">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-emerald-500/10">
                  <Icons.shield className="w-4 h-4 text-emerald-400" />
                </div>
                <div>
                  <div className="text-sm font-medium text-neutral-200">Authenticator App</div>
                  <div className="text-xs text-emerald-400 mt-0.5">Enabled</div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="ghost" onClick={() => setShowRegen(true)}>Backup Codes</Button>
                <Button variant="danger" onClick={() => setShowDisable(true)}>Disable</Button>
              </div>
            </div>
          </div>
        ) : setupData ? (
          <div className="space-y-5">
            <div className="flex flex-col items-center gap-4 p-5 rounded-lg border border-neutral-800 bg-neutral-900/30">
              <p className="text-sm text-neutral-400 text-center">Scan this QR code with your authenticator app</p>
              <img src={generateQRCodeSVG(setupData.url)} alt="QR Code" className="w-[200px] h-[200px] rounded-lg" />
              <div className="w-full">
                <p className="text-xs text-neutral-500 mb-1">Or enter this secret manually:</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 p-2 bg-neutral-900 rounded text-xs text-neutral-300 font-mono break-all border border-neutral-800">{setupData.secret}</code>
                  <button
                    onClick={() => { navigator.clipboard.writeText(setupData.secret); notify('Copied', 'Secret copied to clipboard', 'success'); }}
                    className="shrink-0 p-2 rounded-lg hover:bg-neutral-800 transition-colors"
                  >
                    <Icons.clipboard className="w-4 h-4 text-neutral-400" />
                  </button>
                </div>
              </div>
            </div>
            <form onSubmit={handleEnable} className="space-y-4">
              <Input
                label="Verification code"
                placeholder="Enter the 6-digit code from your app"
                value={setupCode}
                onChange={e => setSetupCode(e.target.value)}
                autoComplete="one-time-code"
              />
              <div className="flex items-center gap-3">
                <Button loading={setupLoading}>Enable 2FA</Button>
                <Button variant="ghost" onClick={() => { setSetupData(null); setSetupCode(''); }}>Cancel</Button>
              </div>
            </form>
          </div>
        ) : (
          <div className="flex items-center justify-between p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-neutral-800">
                <Icons.shield className="w-4 h-4 text-neutral-400" />
              </div>
              <div>
                <div className="text-sm font-medium text-neutral-200">Authenticator App</div>
                <div className="text-xs text-neutral-500 mt-0.5">Use an authenticator app to generate one-time codes.</div>
              </div>
            </div>
            <Button onClick={handleSetup} loading={setupLoading}>Enable</Button>
          </div>
        )}
      </SectionCard>

      <Modal open={!!backupCodes} onClose={() => setBackupCodes(null)} title="Backup Codes" description="Save these codes in a safe place. Each code can only be used once.">
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-2">
            {backupCodes?.map((code, i) => (
              <div key={i} className="p-2 bg-neutral-900 rounded-lg font-mono text-sm text-neutral-300 text-center border border-neutral-800">{code}</div>
            ))}
          </div>
          <div className="rounded-lg border border-neutral-800/50 bg-amber-500/5 p-3 flex items-start gap-3">
            <Icons.errorCircle className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
            <div className="text-xs text-neutral-400 leading-relaxed">
              These codes will <span className="text-neutral-200 font-medium">not</span> be shown again. Store them securely.
            </div>
          </div>
          <div className="flex justify-end gap-3">
            <Button onClick={copyBackupCodes}><Icons.clipboard className="w-4 h-4 mr-1.5" />Copy All</Button>
            <Button variant="ghost" onClick={() => setBackupCodes(null)}>Done</Button>
          </div>
        </div>
      </Modal>

      <Modal open={showDisable} onClose={() => { setShowDisable(false); setDisablePassword(''); }} title="Disable Two-Factor Authentication" description="Enter your password to confirm disabling 2FA.">
        <form onSubmit={handleDisable} className="space-y-4 pt-2">
          <Input label="Password" value={disablePassword} onChange={e => setDisablePassword(e.target.value)} hideable />
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => { setShowDisable(false); setDisablePassword(''); }} disabled={disableLoading}>Cancel</Button>
            <Button variant="danger" loading={disableLoading}>Disable 2FA</Button>
          </div>
        </form>
      </Modal>

      <Modal open={showRegen} onClose={() => { setShowRegen(false); setRegenPassword(''); }} title="Regenerate Backup Codes" description="Enter your password to generate new backup codes. Your old codes will be invalidated.">
        <form onSubmit={handleRegenerate} className="space-y-4 pt-2">
          <Input label="Password" value={regenPassword} onChange={e => setRegenPassword(e.target.value)} hideable />
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => { setShowRegen(false); setRegenPassword(''); }} disabled={regenLoading}>Cancel</Button>
            <Button loading={regenLoading}>Regenerate</Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}

function SessionsTab() {
  const [sessionsState, setSessionsState] = useState<{ list: Session[]; loading: boolean }>({ list: [], loading: true });

  const loadSessions = async () => {
    setSessionsState(s => ({ ...s, loading: true }));
    const res = await getSessions();
    setSessionsState({ list: res.success && res.data ? res.data : [], loading: false });
  };

  useEffect(() => { loadSessions(); }, []);

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

  return (
    <div className="max-w-3xl">
      <SectionCard
        title="Active Sessions"
        description="Devices and browsers that are currently signed in to your account."
        footer={
          sessionsState.list.filter(s => !s.is_current).length > 0
            ? <Button variant="danger" onClick={handleRevokeAll}>Revoke All Other Sessions</Button>
            : undefined
        }
      >
        {sessionsState.loading ? (
          <div className="py-8 text-center text-sm text-neutral-500">Loading sessions...</div>
        ) : sessionsState.list.length === 0 ? (
          <div className="py-8 text-center text-sm text-neutral-500">No active sessions found.</div>
        ) : (
          <div className="space-y-3">
            {sessionsState.list.map((session) => (
              <div key={session.id} className={`flex items-center gap-4 p-4 rounded-lg border transition-colors ${session.is_current
                ? 'border-emerald-500/20 bg-emerald-500/[0.03]'
                : 'border-neutral-800 bg-neutral-900/30 hover:border-neutral-700'
                }`}>
                <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-800 shrink-0">
                  <Icons.monitor className="h-4 w-4 text-neutral-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-neutral-100 truncate">{parseUserAgent(session.user_agent)}</span>
                    {session.is_current && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 shrink-0">
                        <span className="w-1.5 h-1.5 bg-emerald-400 rounded-full" />
                        Current
                      </span>
                    )}
                  </div>
                  <div className="mt-1 flex items-center gap-3 text-xs text-neutral-500">
                    <span className="flex items-center gap-1">
                      <Icons.globe className="w-3 h-3" />
                      {session.ip || 'Unknown IP'}
                    </span>
                    <span className="flex items-center gap-1">
                      <Icons.clock className="w-3 h-3" />
                      {formatDate(session.created_at)}
                    </span>
                  </div>
                </div>
                {!session.is_current && (
                  <button
                    onClick={() => handleRevokeSession(session.id)}
                    className="shrink-0 rounded-lg px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-500/10 transition-colors"
                  >
                    Revoke
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </SectionCard>
    </div>
  );
}

function APIKeysTab() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [apiKeysLoading, setApiKeysLoading] = useState(true);
  const [createPanel, setCreatePanel] = useState({ open: false, loading: false, name: '', expiresIn: '' });
  const [newKey, setNewKey] = useState<APIKeyCreated | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ key: APIKey; loading: boolean } | null>(null);

  const loadApiKeys = async () => {
    setApiKeysLoading(true);
    const res = await getAPIKeys();
    if (res.success && res.data) setApiKeys(res.data);
    setApiKeysLoading(false);
  };

  useEffect(() => { loadApiKeys(); }, []);

  const handleCreateApiKey = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreatePanel(m => ({ ...m, loading: true }));
    const expiresIn = createPanel.expiresIn ? parseInt(createPanel.expiresIn) : undefined;
    const res = await createAPIKey(createPanel.name || 'API Key', expiresIn);
    if (res.success && res.data) {
      setNewKey(res.data);
      setCreatePanel({ open: false, loading: false, name: '', expiresIn: '' });
      loadApiKeys();
    } else {
      notify('Error', res.error || 'Failed to create API key', 'error');
      setCreatePanel(m => ({ ...m, loading: false }));
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

  return (
    <div className="max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-neutral-100">API Keys</h3>
          <p className="text-sm text-neutral-400 mt-0.5">Manage keys for programmatic access to the API.</p>
        </div>
        <Button onClick={() => setCreatePanel(m => ({ ...m, open: true }))}>
          <Icons.plus className="w-4 h-4 mr-1.5" />Create Key
        </Button>
      </div>

      {apiKeysLoading ? (
        <div className="py-12 text-center text-sm text-neutral-500">Loading...</div>
      ) : apiKeys.length === 0 ? (
        <div className="rounded-xl border border-dashed border-neutral-800 py-12 text-center">
          <div className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-neutral-800/60 mb-3">
            <Icons.key className="w-5 h-5 text-neutral-600" />
          </div>
          <p className="text-sm text-neutral-400">No API keys yet.</p>
          <button
            onClick={() => setCreatePanel(m => ({ ...m, open: true }))}
            className="mt-2 inline-flex items-center gap-1 text-xs font-medium border border-neutral-800 rounded-lg px-2 py-1.5 text-neutral-400 hover:text-neutral-100 transition-colors"
          >
            <Icons.plus className="h-3.5 w-3.5" />
            <span className="text-sm font-medium">Create your first key</span>
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {apiKeys.map(k => (
            <div key={k.id} className="flex items-center gap-4 p-4 rounded-xl border border-neutral-800 bg-neutral-900/30 hover:border-neutral-700 transition-colors">
              <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-neutral-800 shrink-0">
                <Icons.key className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-neutral-100">{k.name}</div>
                <div className="mt-0.5 flex items-center gap-3 text-xs text-neutral-500">
                  <span className="font-mono">{k.key_prefix}...</span>
                  <span className="w-1 h-1 rounded-full bg-neutral-700" />
                  <span>Created {new Date(k.created_at).toLocaleDateString()}</span>
                  {k.expires_at && (
                    <>
                      <span className="w-1 h-1 rounded-full bg-neutral-700" />
                      <span>Expires {new Date(k.expires_at).toLocaleDateString()}</span>
                    </>
                  )}
                  {k.last_used_at && (
                    <>
                      <span className="w-1 h-1 rounded-full bg-neutral-700" />
                      <span>Last used {new Date(k.last_used_at).toLocaleString()}</span>
                    </>
                  )}
                </div>
              </div>
              <button
                onClick={() => setDeleteModal({ key: k, loading: false })}
                className="shrink-0 rounded-lg px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-500/10 transition-colors"
              >
                Delete
              </button>
            </div>
          ))}
        </div>
      )}

      <SlidePanel
        open={createPanel.open}
        onClose={() => !createPanel.loading && setCreatePanel(m => ({ ...m, open: false }))}
        title="Create API Key"
        description="Generate a new key for programmatic access to the API."
        footer={
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => setCreatePanel(m => ({ ...m, open: false }))} disabled={createPanel.loading}>Cancel</Button>
            <Button type="submit" form="apiKeyForm" loading={createPanel.loading}>
              <Icons.key className="w-4 h-4 mr-1.5" />Create Key
            </Button>
          </div>
        }
      >
        <form id="apiKeyForm" onSubmit={handleCreateApiKey} className="space-y-6">
          <div className="space-y-2">
            <label className="text-sm font-medium text-neutral-200">Key Name</label>
            <Input
              placeholder="e.g. CLI Access, CI/CD Pipeline..."
              value={createPanel.name}
              onChange={e => setCreatePanel(m => ({ ...m, name: e.target.value }))}
            />
            <p className="text-xs text-neutral-500">A descriptive name to identify this key.</p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-neutral-200">Expiration (days)</label>
            <Input
              placeholder="Leave blank for no expiration"
              type="number"
              min={1}
              value={createPanel.expiresIn}
              onChange={e => setCreatePanel(m => ({ ...m, expiresIn: e.target.value }))}
            />
            <p className="text-xs text-neutral-500">Optionally set the key to auto-expire after a number of days.</p>
          </div>

          <div className="rounded-lg border border-neutral-800/50 bg-amber-500/5 p-4 flex items-start gap-3">
            <Icons.errorCircle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="text-xs text-neutral-400 leading-relaxed">
              The full API key will only be shown <span className="text-neutral-200 font-medium">once</span> after creation. Make sure to copy and store it securely.
            </div>
          </div>
        </form>
      </SlidePanel>

      <Modal open={!!newKey} onClose={() => setNewKey(null)} title="API Key Created" description="Copy your API key now. You won't be able to see it again.">
        <div className="space-y-4">
          <div className="p-3 bg-neutral-900 rounded-lg font-mono text-sm text-emerald-400 break-all border border-neutral-800">{newKey?.key}</div>
          <div className="flex justify-end gap-3">
            <Button onClick={copyKey}><Icons.clipboard className="w-4 h-4 mr-1.5" />Copy</Button>
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

export default function SettingsPage() {
  const basePath = '/console/settings';

  const pluginTabs = getPluginTabs('user-settings')
    .filter(({ pluginId, tab }) => evaluatePluginGuard(pluginId, tab.guard))
    .map(({ pluginId, tab }): PluginTabInfo => ({
      pluginId,
      id: `plugin-${pluginId}-${tab.id}`,
      label: tab.label,
      icon: tab.icon || 'cube',
      component: tab.component,
    }));

  const allTabs = [
    ...settingsTabs,
    ...pluginTabs.map(t => ({ name: t.label, path: `/${t.id}`, icon: t.icon })),
  ];

  return (
    <>
      <SubNavigation basePath={basePath} tabs={allTabs} />

      <Routes>
        <Route path="security" element={<SecurityTab />} />
        <Route path="sessions" element={<SessionsTab />} />
        <Route path="api-keys" element={<APIKeysTab />} />
        {pluginTabs.map(tab => (
          <Route
            key={tab.id}
            path={tab.id}
            element={<PluginRenderer pluginId={tab.pluginId} component={tab.component} props={{}} />}
          />
        ))}
        <Route path="*" element={<AccountTab />} />
      </Routes>
    </>
  );
}
