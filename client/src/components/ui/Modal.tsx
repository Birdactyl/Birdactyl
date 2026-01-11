import { useEffect, useState, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
}

export default function Modal({ open, onClose, title, description, children, className }: ModalProps) {
  const [visible, setVisible] = useState(false);
  const [animate, setAnimate] = useState(false);
  const [closing, setClosing] = useState(false);
  const cachedTitle = useRef(title);
  const cachedDesc = useRef(description);

  const handleClose = useCallback(() => {
    if (closing) return;
    setClosing(true);
    setAnimate(false);
  }, [closing]);

  useEffect(() => {
    if (open) {
      cachedTitle.current = title;
      cachedDesc.current = description;
      setClosing(false);
      setVisible(true);
      requestAnimationFrame(() => requestAnimationFrame(() => setAnimate(true)));
      document.body.style.overflow = 'hidden';
    }
  }, [open, title, description]);

  useEffect(() => {
    if (!open && visible && !closing) {
      handleClose();
    }
  }, [open, visible, closing, handleClose]);

  useEffect(() => {
    if (closing) {
      const timer = setTimeout(() => {
        setVisible(false);
        setClosing(false);
        document.body.style.overflow = '';
        onClose();
      }, 200);
      return () => clearTimeout(timer);
    }
  }, [closing, onClose]);

  useEffect(() => {
    if (!visible) return;
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') handleClose(); };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [visible, handleClose]);

  if (!visible) return null;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className={`absolute inset-0 bg-black/60 transition-opacity duration-200 ${animate && !closing ? 'opacity-100' : 'opacity-0'}`}
        onClick={handleClose}
      />
      <div className={`relative bg-neutral-900 rounded-xl border border-neutral-800 w-full p-6 shadow-2xl transition-all duration-200 ${animate && !closing ? 'opacity-100 scale-100' : 'opacity-0 scale-95'} ${className || 'max-w-md'}`}>
        <h2 className="text-lg font-semibold text-neutral-100 mb-1">{cachedTitle.current}</h2>
        {cachedDesc.current && <p className="text-sm text-neutral-400 mb-6">{cachedDesc.current}</p>}
        {children}
      </div>
    </div>,
    document.body
  );
}
