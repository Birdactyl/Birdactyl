import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, getSFTPDetails, resetSFTPPassword, Server, SFTPDetails } from '../../../lib/api';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { notify, Button, Icons, PermissionDenied } from '../../../components';

export default function SFTPPage() {
  const { id } = useParams<{ id: string }>();
  const [server, setServer] = useState<Server | null>(null);
  const [sftp, setSftp] = useState<SFTPDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [resetting, setResetting] = useState(false);
  const [newPassword, setNewPassword] = useState<string | null>(null);
  const { can, loading: permsLoading } = useServerPermissions(id);

  useEffect(() => {
    if (!id) return;
    Promise.all([
      getServer(id).then(res => res.success && res.data && setServer(res.data)),
      getSFTPDetails(id).then(res => res.success && res.data && setSftp(res.data))
    ]).finally(() => setLoading(false));
  }, [id]);

  const handleResetPassword = async () => {
    if (!id) return;
    setResetting(true);
    const res = await resetSFTPPassword(id);
    if (res.success && res.data) {
      setNewPassword(res.data.password);
      notify('Password Reset', 'Your new SFTP password has been generated', 'success');
    } else {
      notify('Error', res.error || 'Failed to reset password', 'error');
    }
    setResetting(false);
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    notify('Copied', `${label} copied to clipboard`, 'success');
  };

  if (loading || permsLoading) return null;
  if (!can('sftp.view')) return <PermissionDenied message="You don't have permission to view SFTP details" />;

  const host = server?.node?.fqdn || 'unknown';
  const port = sftp?.port || 2022;
  const username = sftp?.username || `${server?.user_id}.${server?.id}`;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-1 text-sm text-neutral-400">
        <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
        <span className="text-neutral-400">/</span>
        <span className="font-semibold text-neutral-100">SFTP</span>
      </div>

      <div>
        <h1 className="text-xl font-semibold text-neutral-100">SFTP Access</h1>
        <p className="text-sm text-neutral-400">Connect to your server files using an SFTP client like FileZilla or WinSCP.</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl bg-neutral-800/30 overflow-hidden">
          <div className="px-6 pt-6 pb-3">
            <h3 className="text-lg font-semibold text-neutral-100">Connection Details</h3>
            <p className="mt-1 text-sm text-neutral-400">Use these credentials in your SFTP client.</p>
          </div>
          <div className="px-6 pb-6 pt-2 space-y-3">
            <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50 group">
              <div>
                <span className="text-xs text-neutral-400 block">Host</span>
                <span className="text-sm text-neutral-100 font-mono">{host}</span>
              </div>
              <button
                onClick={() => copyToClipboard(host, 'Host')}
                className="p-1.5 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-700 transition-colors opacity-0 group-hover:opacity-100"
              >
                <Icons.clipboard className="w-4 h-4" />
              </button>
            </div>

            <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50 group">
              <div>
                <span className="text-xs text-neutral-400 block">Port</span>
                <span className="text-sm text-neutral-100 font-mono">{port}</span>
              </div>
              <button
                onClick={() => copyToClipboard(String(port), 'Port')}
                className="p-1.5 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-700 transition-colors opacity-0 group-hover:opacity-100"
              >
                <Icons.clipboard className="w-4 h-4" />
              </button>
            </div>

            <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50 group">
              <div>
                <span className="text-xs text-neutral-400 block">Username</span>
                <span className="text-sm text-neutral-100 font-mono">{username}</span>
              </div>
              <button
                onClick={() => copyToClipboard(username, 'Username')}
                className="p-1.5 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-700 transition-colors opacity-0 group-hover:opacity-100"
              >
                <Icons.clipboard className="w-4 h-4" />
              </button>
            </div>

            <div className="flex items-center justify-between py-2.5 px-3 rounded-lg bg-neutral-900/50 group">
              <div>
                <span className="text-xs text-neutral-400 block">Password</span>
                <span className="text-sm text-neutral-400">Use your account password or set a dedicated SFTP password</span>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-xl bg-neutral-800/30 overflow-hidden">
            <div className="px-6 pt-6 pb-3">
              <h3 className="text-lg font-semibold text-neutral-100">SFTP Password</h3>
              <p className="mt-1 text-sm text-neutral-400">Generate a dedicated password for SFTP access.</p>
            </div>
            <div className="px-6 pb-6 pt-2 space-y-4">
              {newPassword ? (
                <div className="space-y-3">
                  <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
                    <p className="text-xs text-emerald-400 mb-2">Your new SFTP password (save it now, it won't be shown again):</p>
                    <div className="flex items-center gap-2">
                      <code className="flex-1 text-sm font-mono text-emerald-300 bg-neutral-900/50 px-3 py-2 rounded break-all">{newPassword}</code>
                      <button
                        onClick={() => copyToClipboard(newPassword, 'Password')}
                        className="p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/20 transition-colors"
                      >
                        <Icons.clipboard className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                  <Button variant="ghost" onClick={() => setNewPassword(null)} className="w-full">
                    Done
                  </Button>
                </div>
              ) : (
                <>
                  <p className="text-sm text-neutral-300">
                    Click below to generate a new SFTP password. This will replace any existing SFTP password for this server.
                  </p>
                  {can('sftp.reset_password') && (
                    <Button onClick={handleResetPassword} loading={resetting} className="w-full">
                      Generate New Password
                    </Button>
                  )}
                </>
              )}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
