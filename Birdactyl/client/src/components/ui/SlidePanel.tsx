import { useEffect, useState, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Icons } from '../Icons';

interface SlidePanelProps {
    open: boolean;
    onClose: () => void;
    title: string;
    description?: string;
    children: ReactNode;
    footer?: ReactNode;
    width?: string;
}

export default function SlidePanel({ open, onClose, title, description, children, footer, width = 'max-w-lg' }: SlidePanelProps) {
    const [visible, setVisible] = useState(false);
    const [animate, setAnimate] = useState(false);
    const [closing, setClosing] = useState(false);
    const prevOpen = useRef(open);
    const onCloseRef = useRef(onClose);
    onCloseRef.current = onClose;

    const handleClose = useCallback(() => {
        if (closing) return;
        setClosing(true);
        setAnimate(false);
    }, [closing]);

    useEffect(() => {
        if (open) {
            if (!prevOpen.current) {
                setVisible(true);
                requestAnimationFrame(() => {
                    requestAnimationFrame(() => setAnimate(true));
                });
                document.body.style.overflow = 'hidden';
            }
        } else if (prevOpen.current && visible && !closing) {
            handleClose();
        }
        prevOpen.current = open;
    }, [open, visible, closing, handleClose]);

    useEffect(() => {
        if (closing) {
            const timer = setTimeout(() => {
                setVisible(false);
                setClosing(false);
                document.body.style.overflow = '';
                onCloseRef.current();
            }, 300);
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
        <div className="fixed inset-0 z-50 flex justify-end">
            <div
                className={`absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity duration-300 ${isActive ? 'opacity-100' : 'opacity-0'}`}
                onClick={handleClose}
            />

            <div className={`
        relative w-full h-full flex flex-col
        bg-neutral-900/90 backdrop-blur-2xl
        border-l border-neutral-700/40
        shadow-2xl shadow-black/50
        transition-transform duration-300 ease-out
        ${isActive ? 'translate-x-0' : 'translate-x-full'}
        ${width}
      `}>
                <div className="absolute inset-y-0 -left-px w-px bg-gradient-to-b from-transparent via-neutral-500/40 to-transparent" />

                <div className="relative shrink-0 flex items-start justify-between gap-4 px-6 pt-5 pb-4 border-b border-neutral-700/40">
                    <div className="min-w-0">
                        <h2 className="text-base font-semibold text-neutral-100 tracking-tight">{title}</h2>
                        {description && <p className="mt-1 text-sm text-neutral-400 leading-relaxed">{description}</p>}
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

                <div className="flex-1 overflow-y-auto px-6 py-5">
                    {children}
                </div>

                {footer && (
                    <div className="shrink-0 px-6 py-4 border-t border-neutral-700/40 bg-neutral-900/50">
                        {footer}
                    </div>
                )}
            </div>
        </div>,
        document.body
    );
}
