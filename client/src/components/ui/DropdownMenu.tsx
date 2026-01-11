import { useState, useRef, useEffect, createContext, useContext, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Icons } from '../Icons';

interface DropdownContextType {
  open: boolean;
  setOpen: (open: boolean) => void;
  triggerRef: React.RefObject<HTMLElement | null>;
}

const DropdownContext = createContext<DropdownContextType | null>(null);

export function DropdownMenu({ children, className }: { children: ReactNode; className?: string }) {
  const [open, setOpen] = useState(false);
  const triggerRef = useRef<HTMLElement>(null);

  return (
    <DropdownContext.Provider value={{ open, setOpen, triggerRef }}>
      <div className={`relative ${className || 'inline-block'}`}>{children}</div>
    </DropdownContext.Provider>
  );
}

export function DropdownMenuTrigger({ children, asChild, className }: { children: ReactNode; asChild?: boolean; className?: string }) {
  const ctx = useContext(DropdownContext)!;
  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    ctx.setOpen(!ctx.open);
  };

  if (asChild) {
    return (
      <div onClick={handleClick} ref={ctx.triggerRef as React.RefObject<HTMLDivElement>} className={className || 'w-full'}>
        {children}
      </div>
    );
  }
  return <button ref={ctx.triggerRef as React.RefObject<HTMLButtonElement>} onClick={handleClick} type="button">{children}</button>;
}

interface DropdownMenuContentProps {
  children: ReactNode;
  align?: 'start' | 'center' | 'end';
  side?: 'top' | 'bottom';
  sideOffset?: number;
  className?: string;
}

export function DropdownMenuContent({ children, side, sideOffset = 8, className = '' }: DropdownMenuContentProps) {
  const ctx = useContext(DropdownContext)!;
  const contentRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null);
  const [animating, setAnimating] = useState(false);
  const [show, setShow] = useState(false);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!ctx.open) {
      setReady(false);
      return;
    }
    
    setShow(false);
    setReady(true);
  }, [ctx.open]);

  useEffect(() => {
    if (!ready) return;

    const trigger = ctx.triggerRef.current;
    const content = contentRef.current;
    if (!trigger || !content) return;

    const triggerRect = trigger.getBoundingClientRect();
    const contentRect = content.getBoundingClientRect();
    const vw = document.documentElement.clientWidth;
    const vh = document.documentElement.clientHeight;

    const spaceBelow = vh - triggerRect.bottom;
    const spaceAbove = triggerRect.top;
    let top: number;
    
    const preferTop = side === 'top';
    const preferBottom = side === 'bottom';
    
    if (preferTop && spaceAbove >= contentRect.height + sideOffset) {
      top = triggerRect.top - contentRect.height - sideOffset;
    } else if (preferBottom && spaceBelow >= contentRect.height + sideOffset) {
      top = triggerRect.bottom + sideOffset;
    } else if (spaceBelow >= contentRect.height + sideOffset || spaceBelow >= spaceAbove) {
      top = triggerRect.bottom + sideOffset;
    } else {
      top = triggerRect.top - contentRect.height - sideOffset;
    }

    let left: number;
    if (triggerRect.left + contentRect.width <= vw) {
      left = triggerRect.left;
    } else {
      left = triggerRect.right - contentRect.width;
    }
    left = Math.max(0, Math.min(left, vw - contentRect.width));

    setPos({ top, left });
    requestAnimationFrame(() => setShow(true));
  }, [ready, sideOffset]);

  useEffect(() => {
    if (!ctx.open && show) {
      setShow(false);
      setAnimating(true);
      const timer = setTimeout(() => {
        setAnimating(false);
        setPos(null);
      }, 150);
      return () => clearTimeout(timer);
    }
  }, [ctx.open, show]);

  useEffect(() => {
    if (!ctx.open) return;
    const handleClick = (e: MouseEvent) => {
      if (!contentRef.current?.contains(e.target as Node) && 
          !ctx.triggerRef.current?.contains(e.target as Node)) {
        ctx.setOpen(false);
      }
    };
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') ctx.setOpen(false);
    };
    document.addEventListener('mousedown', handleClick);
    document.addEventListener('keydown', handleKey);
    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('keydown', handleKey);
    };
  }, [ctx.open]);

  if (!ctx.open && !animating) return null;

  const triggerWidth = ctx.triggerRef.current?.offsetWidth;

  return createPortal(
    <div
      ref={contentRef}
      role="menu"
      className={`fixed z-[10000] min-w-[8rem] overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900 p-1.5 shadow-2xl transition-[opacity,transform] duration-150 ease-out
        ${show ? 'opacity-100 scale-100' : 'opacity-0 scale-95 pointer-events-none'}
        ${className}`}
      style={{
        top: pos?.top ?? -9999,
        left: pos?.left ?? -9999,
        '--trigger-width': triggerWidth ? `${triggerWidth}px` : 'auto',
      } as React.CSSProperties}
    >
      {children}
    </div>,
    document.body
  );
}

export function DropdownMenuLabel({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`px-2 py-2 text-xs font-medium text-neutral-300 ${className}`}>{children}</div>;
}

export function DropdownMenuSeparator() {
  return <div className="-mx-1.5 my-1.5 h-px bg-neutral-800" />;
}

export function DropdownMenuGroup({ children }: { children: ReactNode }) {
  return <div role="group">{children}</div>;
}

interface DropdownMenuItemProps {
  children: ReactNode;
  disabled?: boolean;
  destructive?: boolean;
  onSelect?: () => void;
  shortcut?: string;
  className?: string;
}

export function DropdownMenuItem({ children, disabled, destructive, onSelect, shortcut, className = '' }: DropdownMenuItemProps) {
  const ctx = useContext(DropdownContext)!;

  const handleClick = () => {
    if (disabled) return;
    onSelect?.();
    ctx.setOpen(false);
  };

  return (
    <div
      role="menuitem"
      onClick={handleClick}
      className={`relative flex cursor-default select-none items-center gap-2 rounded-lg px-2.5 py-2 text-xs outline-none transition-colors
        ${disabled ? 'pointer-events-none opacity-50' : 'cursor-pointer hover:bg-neutral-800'}
        ${destructive ? 'text-red-400 hover:bg-red-500/10' : 'text-neutral-300 hover:text-neutral-100'}
        ${className}`}
    >
      {children}
      {shortcut && <span className="ml-auto text-[10px] tracking-widest text-neutral-500">{shortcut}</span>}
    </div>
  );
}

const SubContext = createContext<{ open: boolean; setOpen: (o: boolean) => void } | null>(null);

export function DropdownMenuSub({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  return <SubContext.Provider value={{ open, setOpen }}><div className="relative">{children}</div></SubContext.Provider>;
}

export function DropdownMenuSubTrigger({ children }: { children: ReactNode }) {
  const sub = useContext(SubContext)!;
  return (
    <div
      onMouseEnter={() => sub.setOpen(true)}
      onMouseLeave={() => sub.setOpen(false)}
      className="relative flex cursor-default select-none items-center gap-2 rounded-md px-2 py-1.5 text-sm text-neutral-200 outline-none hover:bg-neutral-700/70"
    >
      {children}
      <Icons.chevronRight className="ml-auto h-4 w-4" />
    </div>
  );
}

export function DropdownMenuSubContent({ children }: { children: ReactNode }) {
  const sub = useContext(SubContext)!;
  if (!sub.open) return null;
  return (
    <div
      onMouseEnter={() => sub.setOpen(true)}
      onMouseLeave={() => sub.setOpen(false)}
      className="absolute left-full top-0 z-[10000] min-w-[8rem] overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900 p-1.5 shadow-2xl"
    >
      {children}
    </div>
  );
}
