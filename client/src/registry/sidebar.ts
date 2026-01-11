import { registry } from '../lib/registry';

registry.registerSidebarItems([
  { id: 'dashboard', label: 'Dashboard', icon: 'home', href: '/console', section: 'nav', order: 0 },

  { 
    id: 'hosting', 
    label: 'Hosting', 
    icon: 'server', 
    href: '/console/servers', 
    section: 'platform', 
    order: 0,
    children: [{ label: 'Servers', href: '/console/servers' }]
  },

  { 
    id: 'admin', 
    label: 'Admin', 
    icon: 'shield', 
    href: '/console/admin', 
    section: 'admin', 
    order: 0,
    guard: 'admin',
    children: [
      { label: 'Users', href: '/console/admin/users' },
      { label: 'Servers', href: '/console/admin/servers' },
      { label: 'Nodes', href: '/console/admin/nodes' },
      { label: 'Package Delivery', href: '/console/admin/packages' },
      { label: 'IP Bans', href: '/console/admin/ip-bans' },
      { label: 'Activity Logs', href: '/console/admin/logs' },
      { label: 'Database Hosts', href: '/console/admin/database-hosts' },
      { label: 'Marketplace', href: '/console/admin/marketplace' },
    ]
  },
]);
