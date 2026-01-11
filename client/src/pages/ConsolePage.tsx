import { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConsoleLayout, Icons } from '../components';
import { registry } from '../registry';
import { isAdmin } from '../lib/auth';

function AdminGuard({ children }: { children: React.ReactNode }) {
  if (!isAdmin()) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[calc(100vh-6rem)]">
        <div className="rounded-xl bg-neutral-800/30 p-8 text-center max-w-md">
          <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center mx-auto mb-4">
            <Icons.noAccess className="w-6 h-6 text-red-400" />
          </div>
          <h1 className="text-xl font-semibold text-neutral-100 mb-2">Access Denied</h1>
          <p className="text-sm text-neutral-400">You do not have permission to access this page.</p>
        </div>
      </div>
    );
  }
  return <>{children}</>;
}

function PageLoader() {
  return <div className="flex items-center justify-center h-32 text-neutral-500">Loading...</div>;
}

export default function ConsolePage() {
  const pages = registry.getPages();

  return (
    <ConsoleLayout>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          {pages.map(({ path, component: Component, guard }) => (
            <Route
              key={path}
              path={path}
              element={
                guard === 'admin' ? (
                  <AdminGuard><Component /></AdminGuard>
                ) : (
                  <Component />
                )
              }
            />
          ))}
          <Route path="/admin" element={<Navigate to="/console/admin/users" replace />} />
          <Route path="*" element={<Navigate to="/console" replace />} />
        </Routes>
      </Suspense>
    </ConsoleLayout>
  );
}
