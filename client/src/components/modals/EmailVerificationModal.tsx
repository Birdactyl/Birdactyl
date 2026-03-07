import { useState, useEffect } from 'react';
import { Modal, Button, notify } from '..';
import { sendVerificationEmail } from '../../lib/api';
import { eventBus } from '../../lib/eventBus';

export function EmailVerificationModal() {
  const [open, setOpen] = useState(false);
  const [sending, setSending] = useState(false);
  const [email, setEmail] = useState<string | undefined>(undefined);

  useEffect(() => {
    const unsub = eventBus.on('email:verification_required', (data) => {
      setEmail(data?.email);
      setOpen(true);
    });
    return unsub;
  }, []);

  const handleResend = async () => {
    setSending(true);
    const res = await sendVerificationEmail(email);
    if (res.success) {
      notify('Success', 'Verification email sent. Please check your inbox.', 'success');
      setOpen(false);
    } else {
      notify('Error', res.error || 'Failed to send verification email', 'error');
    }
    setSending(false);
  };

  return (
    <Modal
      open={open}
      onClose={() => setOpen(false)}
      title="Verification Required"
      description="You need to verify your email address to perform this action."
      className="max-w-md"
    >
      <div className="space-y-6">
        <p className="text-sm text-neutral-400">
          Your account is limited because your email address has not been verified yet. We sent you an email when you registered. If you need it again, you can request a new one below.
        </p>
        
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button variant="ghost" onClick={() => setOpen(false)}>
            Close
          </Button>
          <Button loading={sending} onClick={handleResend}>
            Resend Email
          </Button>
        </div>
      </div>
    </Modal>
  );
}
