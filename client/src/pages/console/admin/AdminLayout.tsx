import { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { SubNavigation } from '../../../components/layout/SubNavigation';
import { registry } from '../../../lib/registry';

const adminTabs = [
    { name: 'Users', path: '/users', icon: 'users' },
    { name: 'Servers', path: '/servers', icon: 'server' },
    { name: 'Nodes', path: '/nodes', icon: 'globe' },
    { name: 'Packages', path: '/packages', icon: 'cube' },
    { name: 'IP Bans', path: '/ip-bans', icon: 'shield' },
    { name: 'Activity', path: '/logs', icon: 'activity' },
    { name: 'DB Hosts', path: '/database-hosts', icon: 'database' },
    { name: 'Marketplace', path: '/marketplace', icon: 'pieChart' },
];

const adminPages = registry.getPages().filter(p => p.path.startsWith('/admin/'));

export default function AdminLayout() {
    const basePath = '/console/admin';

    return (
        <>
            <SubNavigation basePath={basePath} tabs={adminTabs} />

            <Suspense fallback={<div className="flex items-center justify-center h-32 text-neutral-500">Loading...</div>}>
                <Routes>
                    {adminPages.map(({ path, component: Component }) => (
                        <Route key={path} path={path.replace('/admin', '')} element={<Component />} />
                    ))}
                    <Route path="/" element={<Navigate to="/console/admin/users" replace />} />
                    <Route path="*" element={<Navigate to="/console/admin/users" replace />} />
                </Routes>
            </Suspense>
        </>
    );
}
