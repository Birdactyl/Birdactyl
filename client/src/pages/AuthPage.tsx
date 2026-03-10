import { useState, useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Input, Button, notify } from '../components';
import { login, register, updatePassword, verify2FA, requestPasswordReset, resetPassword, verifyEmail } from '../lib/api';
import { isAuthenticated, setUser, clearPasswordResetFlag, setAccessToken, setRefreshToken } from '../lib/auth';

const EmailIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="16" x="2" y="4" rx="2" />
    <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
  </svg>
);

const LockIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="18" height="11" x="3" y="11" rx="2" ry="2" />
    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
  </svg>
);

const UserIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
    <circle cx="12" cy="7" r="4" />
  </svg>
);

function AuthCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    let tick = 0;

    function resize() {
      const w = window.innerWidth;
      const h = window.innerHeight;
      canvas!.width = Math.floor(w * dpr);
      canvas!.height = Math.floor(h * dpr);
      canvas!.style.width = w + 'px';
      canvas!.style.height = h + 'px';
      ctx!.setTransform(1, 0, 0, 1, 0, 0);
      ctx!.scale(dpr, dpr);
    }

    function draw() {
      const w = canvas!.width / dpr;
      const h = canvas!.height / dpr;
      if (!w || !h) return;
      ctx!.clearRect(0, 0, w, h);

      const cx = w / 2;
      const cy = h / 2;
      const maxR = Math.sqrt(cx * cx + cy * cy);
      const dot = 2;

      const angleA = tick * 0.0004;
      const angleB = -tick * 0.0003;
      const cosA = Math.cos(angleA);
      const sinA = Math.sin(angleA);
      const cosB = Math.cos(angleB);
      const sinB = Math.sin(angleB);
      const sA = 8;
      const sB = 9;

      for (let y = 0; y < h; y += dot) {
        for (let x = 0; x < w; x += dot) {
          const dx = x - cx;
          const dy = y - cy;
          const dist = Math.sqrt(dx * dx + dy * dy) / maxR;
          const edgeStart = 0.42;
          const edgeProgress = Math.max(0, (dist - edgeStart) / (1 - edgeStart));
          if (edgeProgress <= 0) continue;
          const es = edgeProgress * edgeProgress;

          const ax = dx * cosA - dy * sinA;
          const ay = dx * sinA + dy * cosA;
          const gA = (Math.cos(ax * (Math.PI * 2 / sA)) + 1) * 0.5 * (Math.cos(ay * (Math.PI * 2 / sA)) + 1) * 0.5;

          const bx = dx * cosB - dy * sinB;
          const by = dx * sinB + dy * cosB;
          const gB = (Math.cos(bx * (Math.PI * 2 / sB)) + 1) * 0.5 * (Math.cos(by * (Math.PI * 2 / sB)) + 1) * 0.5;

          const mv = gA * gB;
          if (mv > 0.93 - es * 0.55) {
            ctx!.fillStyle = `rgba(255,255,255,${0.03 + es * 0.1})`;
            ctx!.fillRect(x, y, dot, dot);
          }
        }
      }
      tick++;
    }

    resize();
    draw();
    const interval = setInterval(draw, 60);
    window.addEventListener('resize', resize);

    return () => {
      clearInterval(interval);
      window.removeEventListener('resize', resize);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 pointer-events-none"
    />
  );
}

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

  const [twoFactorChallenge, setTwoFactorChallenge] = useState<string | null>(null);
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [useBackupCode, setUseBackupCode] = useState(false);

  const [searchParams] = useSearchParams();
  const [forgotPassword, setForgotPassword] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');
  const [forgotSent, setForgotSent] = useState(false);
  const [resetToken, setResetToken] = useState<string | null>(null);
  const [resetNewPassword, setResetNewPassword] = useState('');
  const [resetConfirmPassword, setResetConfirmPassword] = useState('');
  const [resetDone, setResetDone] = useState(false);
  const [verifyToken, setVerifyToken] = useState<string | null>(null);
  const [verifyStatus, setVerifyStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [verifyError, setVerifyError] = useState('');

  useEffect(() => {
    if (isAuthenticated()) {
      navigate('/console', { replace: true });
    }
    const rToken = searchParams.get('reset');
    if (rToken) {
      setResetToken(rToken);
    }
    const vToken = searchParams.get('verify');
    if (vToken) {
      setVerifyToken(vToken);
      verifyEmail(vToken).then(res => {
        if (res.success) {
          setVerifyStatus('success');
        } else {
          setVerifyStatus('error');
          setVerifyError(res.error || 'Verification failed');
        }
      });
    }
  }, [navigate, searchParams]);

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

    const data = result.data as { user?: { id: string; username: string; email: string; is_admin: boolean; force_password_reset: boolean; totp_enabled: boolean; email_verified: boolean }; tokens?: { access_token: string; refresh_token: string }; '2fa_required'?: boolean; challenge_token?: string };

    if (data?.['2fa_required'] && data?.challenge_token) {
      setTwoFactorChallenge(data.challenge_token);
      setLoading(false);
      return;
    }

    if (data?.tokens) {
      setAccessToken(data.tokens.access_token);
      setRefreshToken(data.tokens.refresh_token);
    }

    if (data?.user) {
      setUser(data.user);
      if (data.user.force_password_reset) {
        setForceReset(true);
        setLoading(false);
        return;
      }
    }

    notify(isLogin ? 'Welcome back!' : 'Account created', isLogin ? 'Successfully logged in' : 'You can now use your account', 'success');
    navigate('/console', { replace: true });
    setLoading(false);
  };

  const handleTwoFactorVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!twoFactorChallenge || !twoFactorCode) return;
    setLoading(true);

    const result = await verify2FA(twoFactorChallenge, twoFactorCode);

    if (!result.success) {
      notify('Verification failed', result.error || 'Invalid code', 'error');
      setLoading(false);
      return;
    }

    const data = result.data as { user?: { id: string; username: string; email: string; is_admin: boolean; force_password_reset: boolean; totp_enabled: boolean; email_verified: boolean }; tokens?: { access_token: string; refresh_token: string } };

    if (data?.tokens) {
      setAccessToken(data.tokens.access_token);
      setRefreshToken(data.tokens.refresh_token);
    }

    if (data?.user) {
      setUser(data.user);
      if (data.user.force_password_reset) {
        setForceReset(true);
        setTwoFactorChallenge(null);
        setLoading(false);
        return;
      }
    }

    notify('Welcome back!', 'Successfully logged in', 'success');
    navigate('/console', { replace: true });
    setLoading(false);
  };

  const handleForgotPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!forgotEmail) return;
    setLoading(true);
    const res = await requestPasswordReset(forgotEmail);
    if (res.success) {
      setForgotSent(true);
    } else {
      notify('Error', res.error || 'Something went wrong', 'error');
    }
    setLoading(false);
  };

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!resetToken) return;
    if (resetNewPassword !== resetConfirmPassword) {
      notify('Error', 'Passwords do not match', 'error');
      return;
    }
    if (resetNewPassword.length < 8) {
      notify('Error', 'Password must be at least 8 characters', 'error');
      return;
    }
    setLoading(true);
    const res = await resetPassword(resetToken, resetNewPassword);
    if (res.success) {
      setResetDone(true);
      notify('Password reset', 'Your password has been updated. You can now log in.', 'success');
    } else {
      notify('Error', res.error || 'Failed to reset password', 'error');
    }
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

  if (verifyToken) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
        <AuthCanvas />
        <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
          <div className="mx-auto w-full max-w-sm">
            <div className="mb-8 space-y-4">
              <h1 className="text-xl font-medium tracking-tight text-neutral-500">
                Email Verification<br />
                <span className="text-white">
                  {verifyStatus === 'loading' ? 'Verifying...' : verifyStatus === 'success' ? 'Verified' : 'Verification failed'}
                </span>
              </h1>
            </div>
            <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
              {verifyStatus === 'loading' && (
                <p className="text-sm text-neutral-400">Verifying your email address...</p>
              )}
              {verifyStatus === 'success' && (
                <div className="space-y-5">
                  <p className="text-sm text-neutral-400">Your email has been verified. You can now use all features of your account.</p>
                  <Button className="w-full" onClick={() => navigate('/auth', { replace: true })}>Continue to login</Button>
                </div>
              )}
              {verifyStatus === 'error' && (
                <div className="space-y-5">
                  <p className="text-sm text-red-400">{verifyError}</p>
                  <Button className="w-full" onClick={() => navigate('/auth', { replace: true })}>Back to login</Button>
                </div>
              )}
            </div>
          </div>
        </main>
        <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
          <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
            Powered by Birdactyl
          </a>
        </p>
      </div>
    );
  }

  if (resetToken) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
        <AuthCanvas />
        <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
          <div className="mx-auto w-full max-w-sm">
            <div className="mb-8 space-y-4">
              <h1 className="text-xl font-medium tracking-tight text-neutral-500">
                Reset your password<br />
                <span className="text-white">{resetDone ? 'All done' : 'Choose a new password'}</span>
              </h1>
            </div>
            <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
              {resetDone ? (
                <div className="space-y-5">
                  <p className="text-sm text-neutral-400">Your password has been updated. You can now log in with your new password.</p>
                  <Button className="w-full" onClick={() => { setResetToken(null); setResetDone(false); navigate('/auth', { replace: true }); }}>Back to login</Button>
                </div>
              ) : (
                <form className="space-y-5" onSubmit={handleResetPassword}>
                  <Input label="New Password" icon={LockIcon} placeholder="New password" value={resetNewPassword} onChange={e => setResetNewPassword(e.target.value)} hideable required />
                  <Input label="Confirm Password" icon={LockIcon} placeholder="Confirm password" value={resetConfirmPassword} onChange={e => setResetConfirmPassword(e.target.value)} hideable required />
                  <Button className="w-full" loading={loading}>Reset Password</Button>
                </form>
              )}
            </div>
          </div>
        </main>
        <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
          <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
            Powered by Birdactyl
          </a>
        </p>
      </div>
    );
  }

  if (forgotPassword) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
        <AuthCanvas />
        <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
          <div className="mx-auto w-full max-w-sm">
            <div className="mb-8 space-y-4">
              <h1 className="text-xl font-medium tracking-tight text-neutral-500">
                Forgot your password?<br />
                <span className="text-white">{forgotSent ? 'Check your email' : 'Enter your email to reset it'}</span>
              </h1>
            </div>
            <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
              {forgotSent ? (
                <div className="space-y-5">
                  <p className="text-sm text-neutral-400">If an account exists with that email, we've sent a password reset link. Check your inbox and spam folder.</p>
                  <Button className="w-full" onClick={() => { setForgotPassword(false); setForgotSent(false); setForgotEmail(''); }}>Back to login</Button>
                </div>
              ) : (
                <form className="space-y-5" onSubmit={handleForgotPassword}>
                  <Input label="Email" icon={EmailIcon} type="email" placeholder="you@example.com" value={forgotEmail} onChange={e => setForgotEmail(e.target.value)} required />
                  <Button className="w-full" loading={loading}>Send Reset Link</Button>
                </form>
              )}
            </div>
            {!forgotSent && (
              <p className="mt-16 text-center text-sm text-neutral-500">
                Remember your password?
                <button type="button" onClick={() => setForgotPassword(false)} className="ml-1 font-semibold text-white hover:underline cursor-pointer">Sign in</button>
              </p>
            )}
          </div>
        </main>
        <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
          <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
            Powered by Birdactyl
          </a>
        </p>
      </div>
    );
  }

  if (twoFactorChallenge) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
        <AuthCanvas />
        <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
          <div className="mx-auto w-full max-w-sm">
            <div className="mb-8 space-y-4">
              <h1 className="text-xl font-medium tracking-tight text-neutral-500">
                Two-factor authentication<br />
                <span className="text-white">{useBackupCode ? 'Enter a backup code' : 'Enter your authentication code'}</span>
              </h1>
            </div>
            <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
              <form className="space-y-5" onSubmit={handleTwoFactorVerify}>
                <Input
                  label={useBackupCode ? 'Backup Code' : 'Authentication Code'}
                  icon={LockIcon}
                  placeholder={useBackupCode ? '00000000' : '000000'}
                  value={twoFactorCode}
                  onChange={e => setTwoFactorCode(e.target.value)}
                  autoComplete="one-time-code"
                  required
                />
                <Button className="w-full" loading={loading}>Verify</Button>
              </form>
              <div className="mt-4 flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => { setUseBackupCode(!useBackupCode); setTwoFactorCode(''); }}
                  className="text-xs text-neutral-500 hover:text-white transition-colors"
                >
                  {useBackupCode ? 'Use authenticator app' : 'Use a backup code'}
                </button>
                <button
                  type="button"
                  onClick={() => { setTwoFactorChallenge(null); setTwoFactorCode(''); setUseBackupCode(false); }}
                  className="text-xs text-neutral-500 hover:text-white transition-colors"
                >
                  Back to login
                </button>
              </div>
            </div>
          </div>
        </main>
        <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
          <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
            Powered by Birdactyl
          </a>
        </p>
      </div>
    );
  }

  if (forceReset) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
        <AuthCanvas />
        <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
          <div className="mx-auto w-full max-w-sm">
            <div className="mb-8 space-y-4">
              <h1 className="text-xl font-medium tracking-tight text-neutral-500">
                Password Reset Required<br />
                <span className="text-white">Choose a new password</span>
              </h1>
            </div>
            <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
              <form className="space-y-5" onSubmit={handlePasswordReset}>
                <Input label="New Password" icon={LockIcon} placeholder="New password" value={newPassword} onChange={e => setNewPassword(e.target.value)} hideable required />
                <Input label="Confirm Password" icon={LockIcon} placeholder="Confirm password" value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)} hideable required />
                <Button className="w-full" loading={loading}>Update Password</Button>
              </form>
            </div>
          </div>
        </main>
        <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
          <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
            Powered by Birdactyl
          </a>
        </p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0a0a0a] relative overflow-hidden">
      <AuthCanvas />
      <main className="relative z-10 min-h-dvh flex flex-col items-center justify-center p-6">
        <div className="mx-auto w-full max-w-sm">
          <div className="mb-8 space-y-4">
            <h1 className="text-xl font-medium tracking-tight text-neutral-500">
              {isLogin ? (
                <>Welcome back to Birdactyl<br /><span className="text-white">Log in to continue</span></>
              ) : (
                <>Create your Birdactyl account<br /><span className="text-white">Get started in seconds</span></>
              )}
            </h1>
          </div>

          <div className="rounded-lg bg-neutral-950 p-8 ring-1 ring-neutral-800">
            <form className="space-y-5" onSubmit={handleSubmit} noValidate>
              <Input
                label="Email"
                icon={EmailIcon}
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                disableAutofill
                required
              />
              {!isLogin && (
                <Input
                  label="Username"
                  icon={UserIcon}
                  type="text"
                  placeholder="coolhandle"
                  autoComplete="username"
                  value={username}
                  onChange={e => setUsername(e.target.value)}
                  disableAutofill
                  required
                />
              )}
              <Input
                label="Password"
                icon={LockIcon}
                placeholder="Password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                disableAutofill
                hideable
                required
              />
              <Button className="w-full" loading={loading}>
                {isLogin ? 'Sign in' : 'Create account'}
              </Button>
              {isLogin && (
                <div className="text-center">
                  <button type="button" onClick={() => setForgotPassword(true)} className="text-xs text-neutral-500 hover:text-white transition-colors">
                    Forgot your password?
                  </button>
                </div>
              )}
            </form>
          </div>

          <p className="mt-16 text-center text-sm text-neutral-500">
            {isLogin ? "Don't have an account?" : 'Already have an account?'}
            <button
              type="button"
              onClick={() => setIsLogin(!isLogin)}
              className="ml-1 font-semibold text-white hover:underline cursor-pointer"
            >
              {isLogin ? 'Create one' : 'Sign in'}
            </button>
          </p>
        </div>
      </main>
      <p className="fixed bottom-6 left-0 right-0 z-10 text-center text-sm font-semibold text-neutral-500 tracking-tight">
        <a href="https://github.com/Birdactyl/Birdactyl" target="_blank" rel="noopener noreferrer" className="hover:text-neutral-300 transition-colors">
          Powered by Birdactyl
        </a>
      </p>
    </div>
  );
}
