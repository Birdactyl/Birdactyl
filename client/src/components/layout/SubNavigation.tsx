import { createContext, useContext, useState, useEffect, useCallback, useId, useMemo, useRef } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ContextMenuZone } from '../ui/ContextMenu';

export interface SubNavTab {
    name: string;
    path: string;
    icon: string;
}

interface SubNavState {
    id: string;
    tabs: SubNavTab[];
    basePath: string;
    leading?: React.ReactNode;
}

interface SubNavContextType {
    subNav: SubNavState | null;
    register: (state: SubNavState) => void;
    unregister: (id: string) => void;
}

const SubNavContext = createContext<SubNavContextType | null>(null);

export function SubNavProvider({ children }: { children: React.ReactNode }) {
    const [subNav, setSubNav] = useState<SubNavState | null>(null);

    const register = useCallback((state: SubNavState) => {
        setSubNav(state);
    }, []);

    const unregister = useCallback((id: string) => {
        setSubNav(current => (current?.id === id ? null : current));
    }, []);

    return (
        <SubNavContext.Provider value={{ subNav, register, unregister }}>
            {children}
        </SubNavContext.Provider>
    );
}

export function useSubNav() {
    const ctx = useContext(SubNavContext);
    if (!ctx) throw new Error('useSubNav must be used within SubNavProvider');
    return ctx;
}

export function SubNavigation({ basePath, tabs, leading }: {
    basePath: string;
    tabs: SubNavTab[];
    leading?: React.ReactNode;
}) {
    const id = useId();
    const { register, unregister } = useSubNav();

    useEffect(() => {
        register({ id, tabs, basePath, leading });
        return () => unregister(id);
    }, [id, register, unregister, basePath, tabs, leading]);

    return null;
}
import { useSlidingIndicator } from './Topbar';

export function SubNavBar({ subNav }: { subNav: SubNavState }) {
    const location = useLocation();
    const currentPath = location.pathname.replace(subNav.basePath, '') || '';
    const containerRef = useRef<HTMLDivElement>(null);

    const tabItems = useMemo(() =>
        subNav.tabs.map(tab => ({
            tab,
            href: `${subNav.basePath}${tab.path}`,
            contextItems: [
                { label: 'Open in new tab', onClick: () => window.open(`${subNav.basePath}${tab.path}`, '_blank') },
            ] as const,
        })),
        [subNav.basePath, subNav.tabs]
    );

    const activeKey = (() => {
        const found = [...subNav.tabs].reverse().find(tab =>
            tab.path === ''
                ? (currentPath === '' || currentPath === '/')
                : currentPath.startsWith(tab.path)
        );
        return found?.path ?? '';
    })();

    const { setTabRef, style: indicatorStyle } = useSlidingIndicator(activeKey, containerRef);

    return (
        <div ref={containerRef} className="relative flex items-center h-14 gap-1 overflow-x-auto scrollbar-hide">
            <div
                className="absolute top-1/2 -translate-y-1/2 h-9 rounded-lg bg-neutral-800 pointer-events-none"
                style={indicatorStyle}
            />
            {subNav.leading}
            {tabItems.map(({ tab, href, contextItems }) => {
                const active = tab.path === ''
                    ? (currentPath === '' || currentPath === '/')
                    : currentPath.startsWith(tab.path);
                return (
                    <ContextMenuZone key={tab.path} items={[...contextItems]} className="inline-flex">
                        <Link
                            ref={setTabRef(tab.path)}
                            to={href}
                            className={`
                                relative inline-flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium
                                transition-colors whitespace-nowrap shrink-0 z-[1]
                                ${active
                                    ? 'text-neutral-100'
                                    : 'text-neutral-500 hover:text-neutral-300'
                                }
                            `}
                        >
                            {tab.name}
                        </Link>
                    </ContextMenuZone>
                );
            })}
        </div>
    );
}
