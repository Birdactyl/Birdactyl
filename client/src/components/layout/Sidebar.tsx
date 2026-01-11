import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { getUser, isAdmin } from '../../lib/auth';
import { logout } from '../../lib/api';
import { registry, SidebarItem } from '../../registry';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuItem,
} from '../ui/DropdownMenu';
import { Icons } from '../Icons';

type IconKey = keyof typeof Icons;

function UserDropdown({ user }: { user: { username: string; email: string } | null }) {
  const navigate = useNavigate();
  const handleLogout = async () => {
    await logout();
    navigate('/auth', { replace: true });
  };

  return (
    <DropdownMenu className="relative w-full">
      <DropdownMenuTrigger asChild className="w-full">
        <button className="flex w-full items-center rounded-md px-2 py-2 text-left text-xs text-neutral-200 hover:bg-neutral-800 transition-colors cursor-pointer">
          <div className="inline-flex items-center gap-2 min-w-0">
            <div className="flex items-center justify-center rounded-lg overflow-hidden" style={{ width: 20, height: 20 }}>
              <div className="w-full h-full bg-neutral-600 flex items-center justify-center text-[10px] font-bold text-white">
                {user?.username?.[0]?.toUpperCase() || 'U'}
              </div>
            </div>
            <div className="flex flex-col min-w-0">
              <span className="truncate text-xs font-medium">{user?.username || 'User'}</span>
              <span className="truncate text-[9px] text-neutral-400 tracking-widest uppercase">
                {user?.email ? `${user.email.slice(0, 2)}${'*'.repeat(Math.max(0, user.email.split('@')[0].length - 2))}@${user.email.split('@')[1]}`.toUpperCase() : 'US**@EXAMPLE.COM'}
              </span>
            </div>
          </div>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" side="top" sideOffset={16}>
        <DropdownMenuLabel>
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center rounded-lg overflow-hidden" style={{ width: 28, height: 28 }}>
              <div className="w-full h-full bg-neutral-600 flex items-center justify-center text-xs font-bold text-white">
                {user?.username?.[0]?.toUpperCase() || 'U'}
              </div>
            </div>
            <div className="flex flex-col min-w-0">
              <span className="truncate text-xs font-medium text-neutral-100">{user?.username || 'User'}</span>
              <span className="truncate text-[11px] text-neutral-400">{user?.email || 'user@example.com'}</span>
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => navigate('/console/settings')}>
          <Icons.cogFilled className="h-4 w-4" />
          Account Settings
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={handleLogout} destructive>
          <Icons.logout className="h-4 w-4" />
          Logout
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function NavLink({ href, icon, children, active }: { href: string; icon?: string; children: React.ReactNode; active?: boolean }) {
  const IconComponent = icon && icon in Icons ? Icons[icon as IconKey] : null;
  return (
    <Link
      to={href}
      className={`group inline-flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors cursor-pointer border ${
        active
          ? 'font-semibold text-neutral-100 bg-neutral-700/90 border-transparent shadow-xs'
          : 'font-medium text-neutral-400 border-transparent hover:text-neutral-200 hover:border-neutral-800'
      }`}
    >
      {IconComponent && <IconComponent className="h-4 w-4" />}
      <span>{children}</span>
    </Link>
  );
}

function Collapsible({ icon, label, children, defaultOpen = false }: { icon: string; label: string; children: React.ReactNode; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  const IconComponent = icon in Icons ? Icons[icon as IconKey] : null;

  return (
    <div className="w-full mt-1">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="group inline-flex w-full items-center justify-between gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors cursor-pointer border font-medium text-neutral-400 border-transparent hover:text-neutral-200 hover:border-neutral-800"
      >
        <span className="inline-flex items-center gap-2">
          {IconComponent && <IconComponent className="h-4 w-4" />}
          <span className="text-xs font-medium">{label}</span>
        </span>
        <Icons.chevronRightSmall className={`h-3 w-3 transition-transform ${open ? 'rotate-90' : ''}`} />
      </button>
      {open && <div className="mt-1 space-y-1">{children}</div>}
    </div>
  );
}

function SidebarSection({ items }: { items: SidebarItem[] }) {
  const location = useLocation();
  const isActive = (href: string) => location.pathname === href;

  return (
    <>
      {items.map((item) =>
        item.children ? (
          <Collapsible key={item.id} icon={item.icon} label={item.label} defaultOpen>
            {item.children.map((child) => (
              <Link
                key={child.href}
                to={child.href}
                className={`group inline-flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-xs transition-colors cursor-pointer border ${
                  isActive(child.href)
                    ? 'font-semibold text-neutral-100 bg-neutral-700/90 border-transparent shadow-xs'
                    : 'font-medium text-neutral-400 border-transparent hover:text-neutral-200 hover:border-neutral-800'
                }`}
              >
                <span className="inline-block h-4 w-4" />
                <span className="truncate">{child.label}</span>
              </Link>
            ))}
          </Collapsible>
        ) : (
          <div key={item.id} className="w-full mt-1">
            <NavLink href={item.href} icon={item.icon} active={isActive(item.href)}>{item.label}</NavLink>
          </div>
        )
      )}
    </>
  );
}

export default function Sidebar({ onClose }: { onClose?: () => void }) {
  const location = useLocation();
  const isActive = (href: string) => location.pathname === href;
  const user = getUser();

  const navItems = registry.getSidebarItems('nav');
  const platformItems = registry.getSidebarItems('platform');
  const adminItems = registry.getSidebarItems('admin').filter(i => !i.guard || (i.guard === 'admin' && isAdmin()));

  return (
    <aside className="h-full w-64 overflow-y-auto">
      <div className="flex h-full flex-col px-2 py-3">
        <div className="mb-2 px-1.5 flex items-center justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <span className="text-lg font-bold text-white">Birdactyl</span>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="inline-flex cursor-pointer items-center justify-center w-6 h-6 rounded-lg text-neutral-400 hover:text-neutral-100 hover:bg-neutral-700 transition-colors"
            >
              <Icons.sidebarClose className="h-4 w-4" />
            </button>
          )}
        </div>

        <nav className="mt-1 relative space-y-0.5">
          {navItems.map((item) => (
            <div key={item.id} className="w-full mt-1">
              <NavLink href={item.href} icon={item.icon} active={isActive(item.href)}>{item.label}</NavLink>
            </div>
          ))}

          <div className="p-1" />
          <span className="text-[10px] font-medium text-neutral-400 tracking-widest pb-2 ml-2">PLATFORM</span>
          <SidebarSection items={platformItems} />

          {adminItems.length > 0 && (
            <>
              <div className="p-1" />
              <span className="text-[10px] font-medium text-neutral-400 tracking-widest pb-2 ml-2 mt-2">ADMIN</span>
              <SidebarSection items={adminItems} />
            </>
          )}
        </nav>

        <div className="mt-auto space-y-1">
          <div className="flex items-center gap-2 px-1">
            <div className="flex-1 min-w-0">
              <UserDropdown user={user} />
            </div>
          </div>
        </div>
      </div>
    </aside>
  );
}
