import { useState, useEffect, useRef } from 'react';
import { adminUpdateUser } from '../../lib/api';
import { notify, Modal, Input, Button } from '../';

interface User {
  id: string;
  username: string;
  email: string;
  is_admin: boolean;
  is_banned: boolean;
  force_password_reset: boolean;
  ram_limit: number | null;
  cpu_limit: number | null;
  disk_limit: number | null;
  server_limit: number | null;
  created_at: string;
}

interface Props {
  user: User | null;
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
}

type Tab = 'account' | 'resources';

export default function EditUserModal({ user, open, onClose, onSaved }: Props) {
  const [tab, setTab] = useState<Tab>('account');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [ramLimit, setRamLimit] = useState('');
  const [cpuLimit, setCpuLimit] = useState('');
  const [diskLimit, setDiskLimit] = useState('');
  const [serverLimit, setServerLimit] = useState('');
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (user && open) {
      submittingRef.current = false;
      setLoading(false);
      setUsername(user.username);
      setEmail(user.email);
      setPassword('');
      setRamLimit(user.ram_limit?.toString() ?? '');
      setCpuLimit(user.cpu_limit?.toString() ?? '');
      setDiskLimit(user.disk_limit?.toString() ?? '');
      setServerLimit(user.server_limit?.toString() ?? '');
      setTab('account');
    }
  }, [user, open]);

  if (!user) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);

    const data: { email?: string; username?: string; password?: string; ram_limit?: number | null; cpu_limit?: number | null; disk_limit?: number | null; server_limit?: number | null } = {};
    if (username !== user.username) data.username = username;
    if (email !== user.email) data.email = email;
    if (password) data.password = password;

    const ramVal = ramLimit === '' ? null : parseInt(ramLimit);
    const cpuVal = cpuLimit === '' ? null : parseInt(cpuLimit);
    const diskVal = diskLimit === '' ? null : parseInt(diskLimit);
    const serverVal = serverLimit === '' ? null : parseInt(serverLimit);

    if (ramVal !== user.ram_limit) data.ram_limit = ramVal === null ? 0 : ramVal;
    if (cpuVal !== user.cpu_limit) data.cpu_limit = cpuVal === null ? 0 : cpuVal;
    if (diskVal !== user.disk_limit) data.disk_limit = diskVal === null ? 0 : diskVal;
    if (serverVal !== user.server_limit) data.server_limit = serverVal === null ? 0 : serverVal;

    if (Object.keys(data).length === 0) {
      notify('Info', 'No changes to save', 'info');
      setLoading(false);
      submittingRef.current = false;
      return;
    }

    const res = await adminUpdateUser(user.id, data);
    if (res.success) {
      notify('Success', 'User updated', 'success');
      onSaved();
      onClose();
    } else {
      notify('Error', res.error || 'Could not update user', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  const resetLimits = () => {
    setRamLimit('');
    setCpuLimit('');
    setDiskLimit('');
    setServerLimit('');
  };

  return (
    <Modal open={open} onClose={onClose} title="Edit User" description={user.username}>
      <div className="space-y-4">
        <div className="flex gap-1 p-1 bg-neutral-800/50 rounded-lg">
          <button type="button" onClick={() => setTab('account')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'account' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Account</button>
          <button type="button" onClick={() => setTab('resources')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'resources' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Resources</button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          {tab === 'account' && (
            <>
              <Input label="Username" value={username} onChange={e => setUsername(e.target.value)} required />
              <Input label="Email" type="email" value={email} onChange={e => setEmail(e.target.value)} required />
              <Input label="New Password" value={password} onChange={e => setPassword(e.target.value)} hideable placeholder="Leave blank to keep current" />
              <div className="pt-2 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Status</span>
                  <div className="flex gap-2">
                    {user.is_banned && <span className="px-2 py-0.5 rounded bg-red-500/10 text-red-400 ring-1 ring-inset ring-red-500/20">Banned</span>}
                    {user.is_admin && <span className="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 ring-1 ring-inset ring-amber-500/20">Admin</span>}
                    {user.force_password_reset && <span className="px-2 py-0.5 rounded bg-orange-500/10 text-orange-400 ring-1 ring-inset ring-orange-500/20">Reset Required</span>}
                    {!user.is_banned && !user.is_admin && <span className="px-2 py-0.5 rounded bg-neutral-500/10 text-neutral-400 ring-1 ring-inset ring-neutral-500/20">User</span>}
                  </div>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">Joined</span>
                  <span className="text-neutral-300">{new Date(user.created_at).toLocaleDateString()}</span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-neutral-400">User ID</span>
                  <span className="text-neutral-500 font-mono">{user.id.slice(0, 8)}</span>
                </div>
              </div>
            </>
          )}
          {tab === 'resources' && (
            <>
              <p className="text-xs text-neutral-500">Leave blank to use system defaults. Set a value to override for this user.</p>
              <div className="grid grid-cols-2 gap-3">
                <Input label="RAM Limit (MB)" type="number" value={ramLimit} onChange={e => setRamLimit(e.target.value)} placeholder="Default" />
                <Input label="CPU Limit (%)" type="number" value={cpuLimit} onChange={e => setCpuLimit(e.target.value)} placeholder="Default" />
                <Input label="Disk Limit (MB)" type="number" value={diskLimit} onChange={e => setDiskLimit(e.target.value)} placeholder="Default" />
                <Input label="Max Servers" type="number" value={serverLimit} onChange={e => setServerLimit(e.target.value)} placeholder="Default" />
              </div>
              {(user.ram_limit || user.cpu_limit || user.disk_limit || user.server_limit) && (
                <button type="button" onClick={resetLimits} className="text-xs text-neutral-400 hover:text-neutral-200 transition-colors">Reset all to defaults</button>
              )}
            </>
          )}
          <div className="flex justify-end gap-3 pt-4 border-t border-neutral-800">
            <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
            <Button type="submit" loading={loading}>Save Changes</Button>
          </div>
        </form>
      </div>
    </Modal>
  );
}
