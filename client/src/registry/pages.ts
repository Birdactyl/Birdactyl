import { registry, lazyPage } from '../lib/registry';

registry.registerPages([
  { path: '/', component: lazyPage(() => import('../pages/console/DashboardPage')) },
  { path: '/servers', component: lazyPage(() => import('../pages/console/ServersPage')) },
  { path: '/server/:id/*', component: lazyPage(() => import('../pages/console/server/ServerConsolePage')) },
  { path: '/settings', component: lazyPage(() => import('../pages/console/SettingsPage')) },

  { path: '/admin/users', component: lazyPage(() => import('../pages/console/admin/UsersPage')), guard: 'admin' },
  { path: '/admin/nodes', component: lazyPage(() => import('../pages/console/admin/NodesPage')), guard: 'admin' },
  { path: '/admin/packages', component: lazyPage(() => import('../pages/console/admin/PackagesPage')), guard: 'admin' },
  { path: '/admin/servers', component: lazyPage(() => import('../pages/console/admin/ServersPage')), guard: 'admin' },
  { path: '/admin/ip-bans', component: lazyPage(() => import('../pages/console/admin/IPBansPage')), guard: 'admin' },
  { path: '/admin/logs', component: lazyPage(() => import('../pages/console/admin/LogsPage')), guard: 'admin' },
  { path: '/admin/database-hosts', component: lazyPage(() => import('../pages/console/admin/DatabaseHostsPage')), guard: 'admin' },
  { path: '/admin/database-hosts/:id', component: lazyPage(() => import('../pages/console/admin/DatabaseHostPage')), guard: 'admin' },
  { path: '/admin/marketplace', component: lazyPage(() => import('../pages/console/admin/MarketplacePage')), guard: 'admin' },
  
  { path: '/plugins/:pluginId/*', component: lazyPage(() => import('../components/plugins/PluginPage')) },
]);
