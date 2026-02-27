import { ReactNode } from 'react';
import { createPortal } from 'react-dom';

interface FloatingBarProps {
    show: boolean;
    children: ReactNode;
}

export function FloatingBar({ show, children }: FloatingBarProps) {
    if (!show) return null;

    return createPortal(
        <div className="fixed inset-x-0 bottom-0 z-[95] transition-all duration-200 ease-out">
            <div className="mx-auto max-w-2xl px-3 pb-[env(safe-area-inset-bottom)]">
                <div className="rounded-t-lg border border-neutral-800 bg-neutral-900/95 shadow-2xl backdrop-blur px-3 py-2">
                    <div className="flex items-center justify-between gap-3">
                        {children}
                    </div>
                </div>
            </div>
        </div>,
        document.body
    );
}
