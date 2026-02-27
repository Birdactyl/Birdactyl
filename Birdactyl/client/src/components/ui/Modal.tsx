import { useEffect, useState, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Icons } from '../Icons';

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
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

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
      setAnimate(false);
      setVisible(true);
      setTimeout(() => setAnimate(true), 20);
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
        onCloseRef.current();
      }, 250);
      return () => clearTimeout(timer);
    }
  }, [closing]);

  useEffect(() => {
    if (!visible) return;
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') handleClose(); };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [visible, handleClose]);

  if (!visible) return null;

  const isActive = animate && !closing;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className={`absolute inset-0 bg-black/70 backdrop-blur-sm transition-all duration-300 ${isActive ? 'opacity-100' : 'opacity-0'}`}
        onClick={handleClose}
      />

      <div className={`
        relative w-full overflow-hidden
        bg-neutral-900/80 backdrop-blur-2xl
        rounded-2xl
        border border-neutral-700/40
        shadow-xl shadow-black/40
        transition-all duration-300 ease-out
        ${isActive ? 'opacity-100 scale-100 translate-y-0' : 'opacity-0 scale-[0.97] translate-y-2'}
        ${className || 'max-w-md'}
      `}>
        <div className="absolute inset-x-0 -top-px h-px bg-gradient-to-r from-transparent via-neutral-500/60 to-transparent" />
        <div className="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-neutral-500/[0.04] to-transparent pointer-events-none" />

        <div className="relative flex items-start justify-between gap-4 px-6 pt-5 pb-4">
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-neutral-100 tracking-tight">{cachedTitle.current}</h2>
            {cachedDesc.current && <p className="mt-1 text-sm text-neutral-400 leading-relaxed">{cachedDesc.current}</p>}
          </div>
          <button
            type="button"
            onClick={handleClose}
            className="shrink-0 -mr-1 -mt-0.5 inline-flex h-7 w-7 items-center justify-center rounded-lg text-neutral-500 hover:text-neutral-200 hover:bg-neutral-700/50 transition-all duration-150 cursor-pointer"
            aria-label="Close"
          >
            <Icons.xFilled className="w-4 h-4" />
          </button>
        </div>

        <div className="mx-6 border-t border-neutral-700/40" />

        <div className="px-6 pt-4 pb-5">
          {children}
        </div>
      </div>
    </div>,
    document.body
  );
}
