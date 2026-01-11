import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Input, Button, notify } from '../components';
import { login, register, updatePassword } from '../lib/api';
import { isAuthenticated, setUser, clearPasswordResetFlag } from '../lib/auth';

export default function AuthPage() {
  const navigate = useNavigate();
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [forceReset, setForceReset] = useState(false);

  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  useEffect(() => {
    if (isAuthenticated()) {
      navigate('/console', { replace: true });
    }
  }, [navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    const result = isLogin
      ? await login(email, password)
      : await register(email, username, password);

    if (!result.success) {
      if (result.rateLimited) {
        notify('Slow down!', `Too many attempts. Try again in ${result.retryAfter} seconds`, 'error');
      } else if (!result.hasNotifications) {
        notify(isLogin ? 'Login failed' : 'Registration failed', result.error || 'Something went wrong', 'error');
      }
      setLoading(false);
      return;
    }

    const userData = (result.data as { user?: { id: string; username: string; email: string; is_admin: boolean; force_password_reset: boolean } })?.user;
    if (userData) {
      setUser(userData);
      if (userData.force_password_reset) {
        setForceReset(true);
        setLoading(false);
        return;
      }
    }

    notify(isLogin ? 'Welcome back!' : 'Account created', isLogin ? 'Successfully logged in' : 'You can now use your account', 'success');
    navigate('/console', { replace: true });
    setLoading(false);
  };

  const handlePasswordReset = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      notify('Error', 'Passwords do not match', 'error');
      return;
    }
    if (newPassword.length < 8) {
      notify('Error', 'Password must be at least 8 characters', 'error');
      return;
    }
    setLoading(true);
    const res = await updatePassword(password, newPassword);
    if (res.success) {
      clearPasswordResetFlag();
      notify('Success', 'Password updated', 'success');
      navigate('/console', { replace: true });
    } else {
      notify('Error', res.error || 'Failed to update password', 'error');
    }
    setLoading(false);
  };

  if (forceReset) {
    return (
      <div className="w-screen h-screen bg-[#0a0a0a] flex items-center justify-center">
        <div className="w-full max-w-md space-y-8 px-6">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-neutral-100">Password Reset Required</h1>
            <p className="text-sm text-neutral-400">An administrator has required you to change your password.</p>
          </div>
          <form className="space-y-4" onSubmit={handlePasswordReset}>
            <Input label="New Password" placeholder="New password" value={newPassword} onChange={e => setNewPassword(e.target.value)} hideable required />
            <Input label="Confirm Password" placeholder="Confirm password" value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)} hideable required />
            <Button className="w-full" loading={loading}>Update Password</Button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="w-screen h-screen bg-[#0a0a0a] flex">
      <div className="w-1/2 h-full relative bg-gradient-to-br from-[#0a0a0a] via-[#050505] to-black">
        <div className="absolute inset-0 opacity-[0.15] bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PGZlQ29sb3JNYXRyaXggdHlwZT0ic2F0dXJhdGUiIHZhbHVlcz0iMCIvPjwvZmlsdGVyPjxwYXRoIGQ9Ik0wIDBoMzAwdjMwMEgweiIgZmlsdGVyPSJ1cmwoI2EpIiBvcGFjaXR5PSIuMDUiLz48L3N2Zz4=')]" />
      </div>
      <div className="w-1/2 h-full flex items-center justify-center px-6 py-10">
        <div className="w-full max-w-md space-y-8">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-neutral-100">{isLogin ? 'Welcome home' : 'Create an account'}</h1>
            <p className="text-sm text-neutral-400">{isLogin ? 'Sign into your account' : 'Join Birdactyl today'}</p>
          </div>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <Input label="Email" type="email" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} disableAutofill required />
            {!isLogin && <Input label="Username" placeholder="Username" value={username} onChange={e => setUsername(e.target.value)} disableAutofill required />}
            <Input label="Password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} disableAutofill hideable required />
            <Button className="w-full" loading={loading}>{isLogin ? 'Login' : 'Register'}</Button>
          </form>
          <p className="text-xs text-neutral-300">
            {isLogin ? 'New to Birdactyl? ' : 'Have an account? '}
            <button type="button" onClick={() => setIsLogin(!isLogin)} className="text-sky-400 underline underline-offset-2">{isLogin ? 'Register' : 'Login'}</button>
          </p>
          <div className="mt-6 text-xs text-neutral-400">© 2025 Birdactyl</div>
        </div>
      </div>
    </div>
  );
}
