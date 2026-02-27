import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, getSFTPDetails, resetSFTPPassword, Server, SFTPDetails } from '../../../lib/api';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { notify, Button, Icons, PermissionDenied } from '../../../components';

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

function CopyRow({ label, value }: { label: string; value: string }) {
  const copy = () => {
    navigator.clipboard.writeText(value);
    notify('Copied', `${label} copied to clipboard`, 'success');
  };

  return (
    <div className="flex items-center justify-between py-3 px-4 rounded-lg bg-neutral-900/30 border border-neutral-800/50 group">
      <div className="min-w-0">
        <div className="text-[10px] font-medium text-neutral-500 uppercase tracking-wide mb-0.5">{label}</div>
        <div className="text-sm text-neutral-100 font-mono truncate">{value}</div>
      </div>
      <button
        onClick={copy}
        className="shrink-0 p-1.5 rounded-lg text-neutral-500 hover:text-neutral-200 hover:bg-neutral-800 transition-all sm:opacity-0 sm:group-hover:opacity-100"
        title={`Copy ${label}`}
      >
        <Icons.clipboard className="w-4 h-4" />
      </button>
    </div>
  );
}

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

  if (loading || permsLoading) return null;
  if (!can('sftp.view')) return <PermissionDenied message="You don't have permission to view SFTP details" />;

  const host = server?.node?.fqdn || 'unknown';
  const port = String(sftp?.port || 2022);
  const username = sftp?.username || `${server?.user_id}.${server?.id}`;

  return (
    <div className="max-w-3xl space-y-6">
      <SectionCard
        title="Connection Details"
        description="Use these credentials in your SFTP client to connect to your server files."
      >
        <div className="space-y-2">
          <CopyRow label="Host" value={host} />
          <CopyRow label="Port" value={port} />
          <CopyRow label="Username" value={username} />
        </div>
      </SectionCard>

      <SectionCard
        title="Quick Connect"
        description="Copy a ready-to-use connection string for your terminal."
      >
        <CopyRow label="Connection String" value={`sftp -P ${port} ${username}@${host}`} />
        <div className="mt-3 rounded-lg border border-neutral-800/50 bg-blue-500/5 p-3.5 flex items-start gap-3">
          <Icons.errorCircle className="w-4 h-4 text-blue-400 shrink-0 mt-0.5" />
          <p className="text-xs text-neutral-400 leading-relaxed">
            You can also use GUI clients like <span className="text-neutral-200 font-medium">FileZilla</span>, <span className="text-neutral-200 font-medium">WinSCP</span>, or <span className="text-neutral-200 font-medium">Cyberduck</span>. Enter the host, port, and username above with your SFTP password.
          </p>
        </div>
      </SectionCard>

      <SectionCard
        title="SFTP Password"
        description="Generate a dedicated password for SFTP access, separate from your account password."
        footer={
          !newPassword && can('sftp.reset_password')
            ? <Button onClick={handleResetPassword} loading={resetting}><Icons.key className="w-4 h-4 mr-1.5" />Generate New Password</Button>
            : undefined
        }
      >
        {newPassword ? (
          <div className="space-y-4">
            <div className="p-4 rounded-lg border border-emerald-500/20 bg-emerald-500/[0.03]">
              <div className="flex items-start gap-3">
                <Icons.shield className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-emerald-400 font-medium mb-2">Your new SFTP password has been generated. Save it now -- it won't be shown again.</p>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 text-sm font-mono text-emerald-300 bg-neutral-900/60 px-3 py-2 rounded-lg break-all border border-neutral-800">{newPassword}</code>
                    <button
                      onClick={() => { navigator.clipboard.writeText(newPassword); notify('Copied', 'Password copied', 'success'); }}
                      className="shrink-0 p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/10 transition-colors"
                    >
                      <Icons.clipboard className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <Button variant="ghost" onClick={() => setNewPassword(null)} className="w-full">Done</Button>
          </div>
        ) : (
          <div className="flex items-center justify-between p-4 rounded-lg border border-neutral-800 bg-neutral-900/30">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-neutral-800">
                <Icons.key className="w-4 h-4 text-neutral-400" />
              </div>
              <div>
                <div className="text-sm font-medium text-neutral-200">Dedicated SFTP Password</div>
                <div className="text-xs text-neutral-500 mt-0.5">Generate a password specifically for SFTP access to this server.</div>
              </div>
            </div>
          </div>
        )}
      </SectionCard>
    </div>
  );
}
