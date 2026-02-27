import React, { createContext, useContext, useState, useRef, useEffect, useLayoutEffect, useCallback, useId } from 'react';
import { createPortal } from 'react-dom';

interface ContextMenuItem {
    label: React.ReactNode;
    icon?: React.ReactNode;
    onClick?: () => void;
    variant?: 'default' | 'danger';
    disabled?: boolean;
}

type ContextMenuItems = (ContextMenuItem | 'separator')[];

const HIDDEN_STYLE: React.CSSProperties = {
    position: 'fixed',
    top: -9999,
    left: -9999,
    visibility: 'hidden' as const,
    zIndex: 9999,
};

interface ContextMenuContextType {
    register: (id: string, items: ContextMenuItems) => void;
    unregister: (id: string) => void;
}

const ContextMenuContext = createContext<ContextMenuContextType | null>(null);

const MenuPanel = ({
    items,
    menuRef,
    menuStyle,
    positioned,
    isClosing,
    focusedIndex,
    setFocusedIndex,
    onItemClick,
    onKeyDown,
}: {
    items: ContextMenuItems;
    menuRef: React.Ref<HTMLDivElement>;
    menuStyle: React.CSSProperties;
    positioned: boolean;
    isClosing: boolean;
    focusedIndex: number;
    setFocusedIndex: (i: number) => void;
    onItemClick: (item: ContextMenuItem) => void;
    onKeyDown: (e: React.KeyboardEvent) => void;
}) => {
    return createPortal(
        <div
            ref={menuRef}
            style={menuStyle}
            role="menu"
            tabIndex={-1}
            onKeyDown={onKeyDown}
            className={`
        min-w-[180px] p-1
        bg-neutral-100 dark:bg-neutral-900
        rounded-lg
        ring-1 ring-neutral-200 dark:ring-neutral-800
        shadow-xl shadow-black/30
        ${positioned ? (isClosing ? 'animate-menu-out' : 'animate-menu-in') : ''}
      `}
        >
            {items.map((item, index) => {
                if (item === 'separator') {
                    return (
                        <div
                            key={index}
                            className="my-1 mx-1 border-t border-neutral-200 dark:border-neutral-800"
                            role="separator"
                        />
                    );
                }

                const isDanger = item.variant === 'danger';
                const isFocused = focusedIndex === index;

                return (
                    <button
                        key={index}
                        type="button"
                        role="menuitem"
                        disabled={item.disabled}
                        onClick={() => onItemClick(item)}
                        onMouseEnter={() => setFocusedIndex(index)}
                        onMouseLeave={() => setFocusedIndex(-1)}
                        className={`
              w-full flex items-center gap-2 px-2.5 py-1.5 text-sm
              rounded-md transition-colors duration-100 ease-out
              disabled:opacity-40 disabled:cursor-not-allowed
              cursor-pointer select-none tracking-tight

              ${isDanger
                                ? `text-red-500 dark:text-red-400 ${isFocused ? 'bg-red-100 dark:bg-red-500/10' : ''}`
                                : `text-neutral-600 dark:text-neutral-400 ${isFocused ? 'bg-white dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100' : ''}`
                            }
            `}
                    >
                        {item.icon && <span className="h-4 w-4 shrink-0 flex items-center">{item.icon}</span>}
                        <span className="font-semibold">{item.label}</span>
                    </button>
                );
            })}
        </div>,
        document.body
    );
};

function useMenuKeyboard(items: ContextMenuItems, focusedIndex: number, setFocusedIndex: (i: number) => void, onSelect: (item: ContextMenuItem) => void) {
    const actionableIndices = items
        .map((item, i) => (item !== 'separator' && !item.disabled ? i : -1))
        .filter(i => i !== -1);

    const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            const pos = actionableIndices.indexOf(focusedIndex);
            setFocusedIndex(actionableIndices[pos < actionableIndices.length - 1 ? pos + 1 : 0]);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            const pos = actionableIndices.indexOf(focusedIndex);
            setFocusedIndex(actionableIndices[pos > 0 ? pos - 1 : actionableIndices.length - 1]);
        } else if (e.key === 'Enter' && focusedIndex >= 0) {
            e.preventDefault();
            const item = items[focusedIndex];
            if (item !== 'separator' && !item.disabled) {
                onSelect(item);
            }
        }
    }, [items, focusedIndex, actionableIndices, setFocusedIndex, onSelect]);

    return handleKeyDown;
}

export const ContextMenuProvider = ({ children }: { children: React.ReactNode }) => {
    const [isOpen, setIsOpen] = useState(false);
    const [isClosing, setIsClosing] = useState(false);
    const [focusedIndex, setFocusedIndex] = useState(-1);
    const [positioned, setPositioned] = useState(false);
    const [menuStyle, setMenuStyle] = useState<React.CSSProperties>(HIDDEN_STYLE);
    const [activeItems, setActiveItems] = useState<ContextMenuItems>([]);
    const menuRef = useRef<HTMLDivElement>(null);
    const clickPosRef = useRef({ x: 0, y: 0 });
    const registryRef = useRef<Map<string, ContextMenuItems>>(new Map());

    const register = useCallback((id: string, items: ContextMenuItems) => {
        registryRef.current.set(id, items);
    }, []);

    const unregister = useCallback((id: string) => {
        registryRef.current.delete(id);
    }, []);

    const close = useCallback(() => {
        setIsClosing(true);
        setTimeout(() => {
            setIsOpen(false);
            setIsClosing(false);
            setFocusedIndex(-1);
            setPositioned(false);
            setMenuStyle(HIDDEN_STYLE);
            setActiveItems([]);
        }, 80);
    }, []);

    useEffect(() => {
        const handler = (e: MouseEvent) => {
            e.preventDefault();

            if (isOpen) {
                close();
                return;
            }

            const collected: ContextMenuItems = [];
            let el = e.target as HTMLElement | null;
            const foundZones: string[] = [];

            while (el) {
                const zoneId = el.getAttribute('data-context-menu-zone');
                if (zoneId && registryRef.current.has(zoneId)) {
                    foundZones.push(zoneId);
                }
                el = el.parentElement;
            }

            foundZones.forEach((zoneId, i) => {
                if (i > 0) collected.push('separator');
                collected.push(...registryRef.current.get(zoneId)!);
            });

            if (collected.length === 0) return;

            clickPosRef.current = { x: e.clientX, y: e.clientY };
            setActiveItems(collected);
            setPositioned(false);
            setMenuStyle(HIDDEN_STYLE);
            setIsOpen(true);
            setFocusedIndex(-1);
        };

        document.addEventListener('contextmenu', handler);
        return () => document.removeEventListener('contextmenu', handler);
    }, [isOpen, close]);

    useLayoutEffect(() => {
        if (!isOpen || isClosing || positioned || !menuRef.current) return;

        const menu = menuRef.current;
        const { x, y } = clickPosRef.current;
        const menuW = menu.offsetWidth;
        const menuH = menu.offsetHeight;
        const vw = window.innerWidth;
        const vh = window.innerHeight;

        let top = y;
        let left = x;

        if (left + menuW > vw - 8) left = vw - menuW - 8;
        if (left < 8) left = 8;
        if (top + menuH > vh - 8) top = vh - menuH - 8;
        if (top < 8) top = 8;

        setMenuStyle({
            position: 'fixed',
            top,
            left,
            visibility: 'visible',
            zIndex: 9999,
        });
        setPositioned(true);
    }, [isOpen, isClosing, positioned]);

    useEffect(() => {
        if (!isOpen) return;
        const handler = (e: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
                close();
            }
        };
        document.addEventListener('mousedown', handler);
        return () => document.removeEventListener('mousedown', handler);
    }, [isOpen, close]);

    useEffect(() => {
        if (!isOpen) return;
        const handler = (e: KeyboardEvent) => {
            if (e.key === 'Escape') close();
        };
        document.addEventListener('keydown', handler);
        return () => document.removeEventListener('keydown', handler);
    }, [isOpen, close]);

    const handleItemClick = useCallback((item: ContextMenuItem) => {
        item.onClick?.();
        close();
    }, [close]);

    const handleKeyDown = useMenuKeyboard(activeItems, focusedIndex, setFocusedIndex, handleItemClick);

    return (
        <ContextMenuContext.Provider value={{ register, unregister }}>
            {children}
            {isOpen && (
                <MenuPanel
                    items={activeItems}
                    menuRef={menuRef}
                    menuStyle={menuStyle}
                    positioned={positioned}
                    isClosing={isClosing}
                    focusedIndex={focusedIndex}
                    setFocusedIndex={setFocusedIndex}
                    onItemClick={handleItemClick}
                    onKeyDown={handleKeyDown}
                />
            )}
        </ContextMenuContext.Provider>
    );
};

interface ContextMenuZoneProps extends Record<string, unknown> {
    children: React.ReactNode;
    items: ContextMenuItems;
    className?: string;
    as?: React.ElementType;
}

export const ContextMenuZone = ({ children, items, className, as: Tag = 'div', ...rest }: ContextMenuZoneProps) => {
    const id = useId();
    const ctx = useContext(ContextMenuContext);

    useEffect(() => {
        ctx?.register(id, items);
    });

    useEffect(() => {
        return () => ctx?.unregister(id);
    }, [id, ctx]);

    return (
        <Tag data-context-menu-zone={id} className={className} {...rest}>
            {children}
        </Tag>
    );
};

export function useContextMenuZone(items: ContextMenuItems) {
    const id = useId();
    const ctx = useContext(ContextMenuContext);

    useEffect(() => {
        ctx?.register(id, items);
    });

    useEffect(() => {
        return () => ctx?.unregister(id);
    }, [id, ctx]);

    return { 'data-context-menu-zone': id };
}

interface ContextMenuProps {
    trigger: React.ReactNode;
    items: ContextMenuItems;
    align?: 'start' | 'end';
    className?: string;
    rightClick?: boolean;
    rightClickOnly?: boolean;
}

export const ContextMenu = ({
    trigger,
    items,
    align = 'end',
    className = "",
    rightClick = false,
    rightClickOnly = false,
}: ContextMenuProps) => {
    const [isOpen, setIsOpen] = useState(false);
    const [isClosing, setIsClosing] = useState(false);
    const [focusedIndex, setFocusedIndex] = useState(-1);
    const [menuStyle, setMenuStyle] = useState<React.CSSProperties>(HIDDEN_STYLE);
    const [positioned, setPositioned] = useState(false);
    const triggerRef = useRef<HTMLDivElement>(null);
    const menuRef = useRef<HTMLDivElement>(null);
    const triggerRectRef = useRef<DOMRect | null>(null);

    const close = useCallback(() => {
        setIsClosing(true);
        setTimeout(() => {
            setIsOpen(false);
            setIsClosing(false);
            setFocusedIndex(-1);
            setPositioned(false);
            setMenuStyle(HIDDEN_STYLE);
        }, 80);
    }, []);

    const handleTriggerClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (isOpen) {
            close();
            return;
        }
        triggerRectRef.current = triggerRef.current?.getBoundingClientRect() ?? null;
        setPositioned(false);
        setMenuStyle(HIDDEN_STYLE);
        setIsOpen(true);
        setFocusedIndex(-1);
    };

    useLayoutEffect(() => {
        if (!isOpen || isClosing || positioned || !menuRef.current || !triggerRectRef.current) return;

        const menu = menuRef.current;
        const rect = triggerRectRef.current;
        const menuW = menu.offsetWidth;
        const menuH = menu.offsetHeight;
        const vw = window.innerWidth;
        const vh = window.innerHeight;

        let top = rect.bottom + 8;
        let left = align === 'end' ? rect.right - menuW : rect.left;

        if (left + menuW > vw - 8) left = vw - menuW - 8;
        if (left < 8) left = 8;
        if (top + menuH > vh - 8) top = rect.top - menuH - 8;
        if (top < 8) top = 8;

        setMenuStyle({
            position: 'fixed',
            top,
            left,
            visibility: 'visible',
            zIndex: 9999,
        });
        setPositioned(true);
    }, [isOpen, isClosing, positioned, align]);

    useEffect(() => {
        if (!isOpen) return;
        const handler = (e: MouseEvent) => {
            const target = e.target as Node;
            if (
                triggerRef.current && !triggerRef.current.contains(target) &&
                (!menuRef.current || !menuRef.current.contains(target))
            ) {
                close();
            }
        };
        document.addEventListener('mousedown', handler);
        return () => document.removeEventListener('mousedown', handler);
    }, [isOpen, close]);

    useEffect(() => {
        if (!isOpen) return;
        const handler = (e: KeyboardEvent) => {
            if (e.key === 'Escape') close();
        };
        document.addEventListener('keydown', handler);
        return () => document.removeEventListener('keydown', handler);
    }, [isOpen, close]);

    const handleItemClick = useCallback((item: ContextMenuItem) => {
        item.onClick?.();
        close();
    }, [close]);

    const handleKeyDown = useMenuKeyboard(items, focusedIndex, setFocusedIndex, handleItemClick);

    return (
        <div className={`relative inline-flex ${className}`}>
            <div
                ref={triggerRef}
                onClick={rightClickOnly ? undefined : handleTriggerClick}
                onContextMenu={(rightClick || rightClickOnly) ? (e: React.MouseEvent) => { e.preventDefault(); handleTriggerClick(e); } : undefined}
                className="inline-flex"
            >
                {trigger}
            </div>

            {isOpen && (
                <MenuPanel
                    items={items}
                    menuRef={menuRef}
                    menuStyle={menuStyle}
                    positioned={positioned}
                    isClosing={isClosing}
                    focusedIndex={focusedIndex}
                    setFocusedIndex={setFocusedIndex}
                    onItemClick={handleItemClick}
                    onKeyDown={handleKeyDown}
                />
            )}
        </div>
    );
};
