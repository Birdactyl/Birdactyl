import { useState, useRef, useEffect } from 'react';
import { adminCreateUser } from '../../lib/api';
import { notify, Modal, Input, Button } from '../';

interface Props {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}

const generateRandom = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  const rand = (len: number) => Array.from({ length: len }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
  return { email: `${rand(8)}@test.local`, username: `user_${rand(6)}`, password: rand(12) };
};

export default function CreateUserModal({ open, onClose, onCreated }: Props) {
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
      setEmail('');
      setUsername('');
      setPassword('');
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await adminCreateUser(email, username, password);
    if (res.success) {
      notify('User created', `${username} has been created`, 'success');
      onCreated();
      onClose();
    } else {
      notify('Failed', res.error || 'Could not create user', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  const handleRandom = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    const data = generateRandom();
    setLoading(true);
    const res = await adminCreateUser(data.email, data.username, data.password);
    if (res.success) {
      notify('Random user created', `${data.username} / ${data.password}`, 'success');
      onCreated();
      onClose();
    } else {
      notify('Failed', res.error || 'Could not create user', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Create User" description="Add a new user to the system.">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input label="Email" type="email" value={email} onChange={e => setEmail(e.target.value)} disableAutofill required />
        <Input label="Username" value={username} onChange={e => setUsername(e.target.value)} disableAutofill required />
        <Input label="Password" value={password} onChange={e => setPassword(e.target.value)} disableAutofill hideable required />
        <div className="flex items-center justify-between pt-4">
          <button type="button" onClick={handleRandom} disabled={loading} className="text-xs text-neutral-400 hover:text-neutral-200 transition-colors">Generate random user</button>
          <div className="flex gap-3">
            <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
            <Button type="submit" loading={loading}>Create</Button>
          </div>
        </div>
      </form>
    </Modal>
  );
}
