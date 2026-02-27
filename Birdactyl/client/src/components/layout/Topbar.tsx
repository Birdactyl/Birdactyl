import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useRef, useState, useLayoutEffect, useCallback, useEffect } from 'react';
import { getUser, isAdmin } from '../../lib/auth';
import { logout } from '../../lib/api';
import { registry, NavItem } from '../../registry';
import { ContextMenu } from '../ui/ContextMenu';
import { Icons } from '../Icons';
import { useSubNav } from './SubNavigation';
import { SubNavBar } from './SubNavigation';

type IconKey = keyof typeof Icons;

function isActive(itemPath: string, currentPath: string) {
    if (itemPath === '/console') return currentPath === '/console';
    return currentPath.startsWith(itemPath);
}

function useSlidingIndicator(activeKey: string, containerRef: React.RefObject<HTMLElement | null>, modeKey: string = 'default') {
    const tabRefs = useRef<Map<string, HTMLElement>>(new Map());
    const [style, setStyle] = useState<React.CSSProperties>({ opacity: 0 });
    const hasAnimated = useRef<string | null>(null);

    const setTabRef = useCallback((key: string) => (el: HTMLElement | null) => {
        if (el) tabRefs.current.set(key, el);
        else tabRefs.current.delete(key);
    }, []);

    useLayoutEffect(() => {
        const el = tabRefs.current.get(activeKey);
        const container = containerRef.current;
        if (!el || !container) {
            setStyle(s => ({ ...s, opacity: 0 }));
            return;
        }

        const update = () => {
            const containerRect = container.getBoundingClientRect();
            const elRect = el.getBoundingClientRect();
            const skipTransition = hasAnimated.current === modeKey;
            hasAnimated.current = modeKey;
            setStyle({
                left: elRect.left - containerRect.left + container.scrollLeft,
                width: elRect.width,
                opacity: 1,
                transition: skipTransition ? 'left 250ms cubic-bezier(0.4, 0, 0.2, 1), width 250ms cubic-bezier(0.4, 0, 0.2, 1), opacity 150ms' : 'none',
            });
        };

        update();

        const observer = new ResizeObserver(update);
        observer.observe(container);
        return () => observer.disconnect();
    }, [activeKey, containerRef, modeKey]);

    return { setTabRef, style };
}

function UserDropdown({ user }: { user: { username: string; email: string } | null }) {
    const navigate = useNavigate();
    const handleLogout = async () => {
        await logout();
        navigate('/auth', { replace: true });
    };

    const userInitial = user?.username?.charAt(0).toUpperCase() ?? '?';

    return (
        <ContextMenu
            align="end"
            trigger={
                <button
                    type="button"
                    className="flex items-center justify-center w-9 h-9 rounded-lg overflow-hidden transition-all cursor-pointer hover:bg-white/[0.06]"
                >
                    <div className="flex items-center justify-center w-7 h-7 rounded-full bg-neutral-800 ring-1 ring-neutral-700">
                        <span className="text-xs font-bold text-neutral-300 uppercase">{userInitial}</span>
                    </div>
                </button>
            }
            items={[
                { label: `${user?.username || 'User'}`, disabled: true },
                'separator',
                { label: 'Account Settings', icon: <Icons.cogFilled className="h-4 w-4" />, onClick: () => navigate('/console/settings') },
                { label: 'Logout', icon: <Icons.logout className="h-4 w-4" />, onClick: handleLogout, variant: 'danger' },
            ]}
        />
    );
}

function InlineNav({ items, adminItem }: { items: NavItem[]; adminItem: NavItem | undefined }) {
    const location = useLocation();
    const navigate = useNavigate();
    const adminActive = adminItem && location.pathname.startsWith(adminItem.href);
    const AdminIcon = adminItem?.icon && adminItem.icon in Icons ? Icons[adminItem.icon as IconKey] : null;

    return (
        <div className="flex items-center gap-1 ml-4">
            <span className="w-px h-5 bg-neutral-800 mr-1" />
            {items.map((item) => {
                const active = isActive(item.href, location.pathname);
                const IconComponent = item.icon && item.icon in Icons ? Icons[item.icon as IconKey] : null;
                return (
                    <Link
                        key={item.id}
                        to={item.href}
                        className={`
                            inline-flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium
                            transition-colors whitespace-nowrap shrink-0
                            ${active
                                ? 'text-neutral-100 bg-neutral-800'
                                : 'text-neutral-500 hover:text-neutral-300 hover:bg-neutral-800/50'
                            }
                        `}
                        title={item.label}
                    >
                        {IconComponent && <IconComponent className={`h-4 w-4 ${active ? 'text-neutral-300' : 'text-neutral-600'}`} />}
                        <span className="hidden sm:inline">{item.label}</span>
                    </Link>
                );
            })}
            {adminItem && (
                <ContextMenu
                    align="start"
                    rightClickOnly
                    trigger={
                        <Link
                            to={adminItem.href}
                            className={`
                                inline-flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium
                                transition-colors whitespace-nowrap shrink-0
                                ${adminActive
                                    ? 'text-neutral-100 bg-neutral-800'
                                    : 'text-neutral-500 hover:text-neutral-300 hover:bg-neutral-800/50'
                                }
                            `}
                        >
                            {AdminIcon && <AdminIcon className={`h-4 w-4 ${adminActive ? 'text-neutral-300' : 'text-neutral-600'}`} />}
                            <span className="hidden sm:inline">{adminItem.label}</span>
                        </Link>
                    }
                    items={(adminItem.children || []).map(child => ({
                        label: child.label,
                        onClick: () => navigate(child.href),
                    }))}
                />
            )}
        </div>
    );
}

function NavTabs({ items, adminItem }: { items: NavItem[]; adminItem: NavItem | undefined }) {
    const location = useLocation();
    const navigate = useNavigate();
    const containerRef = useRef<HTMLDivElement>(null);
    const adminActive = adminItem && location.pathname.startsWith(adminItem.href);
    const AdminIcon = adminItem?.icon && adminItem.icon in Icons ? Icons[adminItem.icon as IconKey] : null;

    const activeKey = (() => {
        if (adminItem && location.pathname.startsWith(adminItem.href)) return adminItem.id;
        const found = [...items].reverse().find(item => isActive(item.href, location.pathname));
        return found?.id || '';
    })();

    const { setTabRef, style: indicatorStyle } = useSlidingIndicator(activeKey, containerRef, 'tabs');

    return (
        <div ref={containerRef} className="relative flex items-center h-14 gap-1 overflow-x-auto scrollbar-hide">
            <div
                className="absolute top-1/2 -translate-y-1/2 h-9 rounded-lg bg-neutral-800 pointer-events-none"
                style={indicatorStyle}
            />
            {items.map((item) => {
                const active = isActive(item.href, location.pathname);
                const IconComponent = item.icon && item.icon in Icons ? Icons[item.icon as IconKey] : null;
                return (
                    <Link
                        key={item.id}
                        ref={setTabRef(item.id)}
                        to={item.href}
                        className={`
                            relative inline-flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium
                            transition-colors whitespace-nowrap shrink-0 z-[1]
                            ${active
                                ? 'text-neutral-100'
                                : 'text-neutral-500 hover:text-neutral-300'
                            }
                        `}
                    >
                        {IconComponent && (
                            <IconComponent className={`h-4 w-4 transition-colors ${active ? 'text-neutral-300' : 'text-neutral-600'}`} />
                        )}
                        {item.label}
                    </Link>
                );
            })}

            {adminItem && (
                <ContextMenu
                    align="start"
                    rightClickOnly
                    trigger={
                        <Link
                            ref={setTabRef(adminItem.id)}
                            to={adminItem.href}
                            className={`
                                relative inline-flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium
                                transition-colors whitespace-nowrap shrink-0 z-[1]
                                ${adminActive
                                    ? 'text-neutral-100'
                                    : 'text-neutral-500 hover:text-neutral-300'
                                }
                            `}
                        >
                            {AdminIcon && (
                                <AdminIcon className={`h-4 w-4 transition-colors ${adminActive ? 'text-neutral-300' : 'text-neutral-600'}`} />
                            )}
                            {adminItem.label}
                        </Link>
                    }
                    items={(adminItem.children || []).map(child => ({
                        label: child.label,
                        onClick: () => navigate(child.href),
                    }))}
                />
            )}
        </div>
    );
}

export default function Topbar() {
    const location = useLocation();
    const user = getUser();
    const { subNav: explicitSubNav } = useSubNav();

    const navItems = registry.getNavItems('nav');
    const platformItems = registry.getNavItems('platform');
    const adminItems = registry.getNavItems('admin').filter(i => !i.guard || (i.guard === 'admin' && isAdmin()));

    const tabItems: NavItem[] = [...navItems, ...platformItems];
    const adminItem = adminItems[0];

    const autoSubNav = (() => {
        if (explicitSubNav) return null;
        const activeItem = tabItems.find(item => (item.children?.length ?? 0) >= 2 && location.pathname.startsWith(item.href));
        if (!activeItem?.children?.length) return null;
        return {
            id: `auto-${activeItem.id}`,
            basePath: '',
            tabs: activeItem.children.map(child => ({
                name: child.label,
                path: child.href,
                icon: '',
            })),
        };
    })();

    const subNav = explicitSubNav || autoSubNav;
    const hasSubNav = subNav !== null;

    const [showSubNav, setShowSubNav] = useState(hasSubNav);
    const [transitioning, setTransitioning] = useState(false);
    const prevHasSubNav = useRef(hasSubNav);

    useEffect(() => {
        if (hasSubNav !== prevHasSubNav.current) {
            prevHasSubNav.current = hasSubNav;
            setTransitioning(true);
            requestAnimationFrame(() => {
                requestAnimationFrame(() => {
                    setShowSubNav(hasSubNav);
                    setTimeout(() => setTransitioning(false), 200);
                });
            });
            return () => {};
        } else {
            setShowSubNav(hasSubNav);
        }
    }, [hasSubNav]);

    const bottomNavVisible = !showSubNav && !transitioning;
    const subNavVisible = showSubNav && !transitioning;

    return (
        <header className="shrink-0 w-full bg-[#0a0a0a] border-b border-neutral-800/50 z-50">
            <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
                <div className="flex items-center h-14 gap-4">
                    <Link to="/console" className="shrink-0 group flex items-center gap-2.5">
                        <span className="text-sm font-bold tracking-tight text-white group-hover:text-neutral-300 transition-colors">
                            Birdactyl
                        </span>
                    </Link>

                    <div
                        className="flex transition-all duration-200 ease-out overflow-hidden"
                        style={{
                            opacity: hasSubNav ? 1 : 0,
                            maxWidth: hasSubNav ? 600 : 0,
                            transform: hasSubNav ? 'translateX(0)' : 'translateX(-8px)',
                        }}
                    >
                        {hasSubNav && <InlineNav items={tabItems} adminItem={adminItem} />}
                    </div>

                    <div className="ml-auto flex items-center gap-2">
                        <UserDropdown user={user} />
                    </div>
                </div>
            </div>

            <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 relative overflow-hidden">
                <div
                    className="transition-all duration-200 ease-out"
                    style={{
                        opacity: bottomNavVisible ? 1 : 0,
                        transform: bottomNavVisible ? 'translateY(0)' : 'translateY(8px)',
                        height: hasSubNav ? 0 : 'auto',
                        overflow: hasSubNav ? 'hidden' : 'visible',
                        pointerEvents: hasSubNav ? 'none' : 'auto',
                    }}
                >
                    <NavTabs items={tabItems} adminItem={adminItem} />
                </div>

                <div
                    className="transition-all duration-200 ease-out"
                    style={{
                        opacity: subNavVisible ? 1 : 0,
                        transform: subNavVisible ? 'translateY(0)' : 'translateY(-8px)',
                        height: hasSubNav ? 'auto' : 0,
                        overflow: hasSubNav ? 'visible' : 'hidden',
                        pointerEvents: hasSubNav ? 'auto' : 'none',
                    }}
                >
                    {subNav && <SubNavBar subNav={subNav} />}
                </div>
            </div>
        </header>
    );
}

export { useSlidingIndicator };
