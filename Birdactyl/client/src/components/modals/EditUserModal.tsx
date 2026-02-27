import { useState, useEffect, useRef } from 'react';
import { adminUpdateUser } from '../../lib/api';
import { notify, SlidePanel, Input, Button } from '../';

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
  const formRef = useRef<HTMLFormElement>(null);
  const cachedUser = useRef<User | null>(null);

  useEffect(() => {
    if (user && open) {
      cachedUser.current = user;
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

  const activeUser = user ?? cachedUser.current;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingRef.current || !activeUser) return;
    submittingRef.current = true;
    setLoading(true);

    const data: { email?: string; username?: string; password?: string; ram_limit?: number | null; cpu_limit?: number | null; disk_limit?: number | null; server_limit?: number | null } = {};
    if (username !== activeUser.username) data.username = username;
    if (email !== activeUser.email) data.email = email;
    if (password) data.password = password;

    const ramVal = ramLimit === '' ? null : parseInt(ramLimit);
    const cpuVal = cpuLimit === '' ? null : parseInt(cpuLimit);
    const diskVal = diskLimit === '' ? null : parseInt(diskLimit);
    const serverVal = serverLimit === '' ? null : parseInt(serverLimit);

    if (ramVal !== activeUser.ram_limit) data.ram_limit = ramVal === null ? 0 : ramVal;
    if (cpuVal !== activeUser.cpu_limit) data.cpu_limit = cpuVal === null ? 0 : cpuVal;
    if (diskVal !== activeUser.disk_limit) data.disk_limit = diskVal === null ? 0 : diskVal;
    if (serverVal !== activeUser.server_limit) data.server_limit = serverVal === null ? 0 : serverVal;

    if (Object.keys(data).length === 0) {
      notify('Info', 'No changes to save', 'info');
      setLoading(false);
      submittingRef.current = false;
      return;
    }

    const res = await adminUpdateUser(activeUser.id, data);
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
    <SlidePanel
      open={open}
      onClose={onClose}
      title="Edit User"
      description={activeUser?.username ?? ''}
      footer={
        <div className="flex justify-end gap-3">
          <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button onClick={() => formRef.current?.requestSubmit()} loading={loading}>Save Changes</Button>
        </div>
      }
    >
      {activeUser && (
        <div className="space-y-4">
          <div className="flex gap-1 p-1 bg-neutral-800/50 rounded-lg">
            <button type="button" onClick={() => setTab('account')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'account' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Account</button>
            <button type="button" onClick={() => setTab('resources')} className={`flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${tab === 'resources' ? 'bg-neutral-700 text-neutral-100' : 'text-neutral-400 hover:text-neutral-200'}`}>Resources</button>
          </div>
          <form ref={formRef} onSubmit={handleSubmit} className="space-y-4">
            {tab === 'account' && (
              <>
                <Input label="Username" value={username} onChange={e => setUsername(e.target.value)} required />
                <Input label="Email" type="email" value={email} onChange={e => setEmail(e.target.value)} required />
                <Input label="New Password" value={password} onChange={e => setPassword(e.target.value)} hideable placeholder="Leave blank to keep current" />
                <div className="pt-2 space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-neutral-400">Status</span>
                    <div className="flex gap-2">
                      {activeUser.is_banned && <span className="px-2 py-0.5 rounded bg-red-500/10 text-red-400 ring-1 ring-inset ring-red-500/20">Banned</span>}
                      {activeUser.is_admin && <span className="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 ring-1 ring-inset ring-amber-500/20">Admin</span>}
                      {activeUser.force_password_reset && <span className="px-2 py-0.5 rounded bg-orange-500/10 text-orange-400 ring-1 ring-inset ring-orange-500/20">Reset Required</span>}
                      {!activeUser.is_banned && !activeUser.is_admin && <span className="px-2 py-0.5 rounded bg-neutral-500/10 text-neutral-400 ring-1 ring-inset ring-neutral-500/20">User</span>}
                    </div>
                  </div>
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-neutral-400">Joined</span>
                    <span className="text-neutral-300">{new Date(activeUser.created_at).toLocaleDateString()}</span>
                  </div>
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-neutral-400">User ID</span>
                    <span className="text-neutral-500 font-mono">{activeUser.id.slice(0, 8)}</span>
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
                {(activeUser.ram_limit || activeUser.cpu_limit || activeUser.disk_limit || activeUser.server_limit) && (
                  <button type="button" onClick={resetLimits} className="text-xs text-neutral-400 hover:text-neutral-200 transition-colors">Reset all to defaults</button>
                )}
              </>
            )}
          </form>
        </div>
      )}
    </SlidePanel>
  );
}
