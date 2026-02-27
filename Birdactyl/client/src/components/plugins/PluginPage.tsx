import { useParams } from 'react-router-dom';
import PluginRenderer from './PluginRenderer';
import { getLoadedPlugin, evaluatePluginGuard } from '../../lib/pluginLoader';
import { Icons } from '../Icons';

export default function PluginPage() {
  const { pluginId, '*': subPath } = useParams<{ pluginId: string; '*': string }>();
  
  if (!pluginId) {
    return <PluginNotFound />;
  }

  const plugin = getLoadedPlugin(pluginId);
  
  if (!plugin) {
    return <PluginNotFound pluginId={pluginId} />;
  }

  const pagePath = '/' + (subPath || '');
  const page = plugin.manifest.pages.find(p => p.path === pagePath);

  if (!page) {
    return <PageNotFound pluginId={pluginId} path={pagePath} />;
  }

  if (!evaluatePluginGuard(pluginId, page.guard)) {
    return <AccessDenied />;
  }

  return (
    <div className="plugin-page">
      <PluginRenderer 
        pluginId={pluginId} 
        component={page.component}
        props={{ path: pagePath }}
      />
    </div>
  );
}

function PluginNotFound({ pluginId }: { pluginId?: string }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-12rem)]">
      <div className="rounded-xl bg-neutral-800/30 p-8 text-center max-w-md">
        <div className="w-12 h-12 rounded-full bg-neutral-700/50 flex items-center justify-center mx-auto mb-4">
          <Icons.puzzle className="w-6 h-6 text-neutral-400" />
        </div>
        <h1 className="text-xl font-semibold text-neutral-100 mb-2">Plugin Not Found</h1>
        <p className="text-sm text-neutral-400">
          {pluginId 
            ? `The plugin "${pluginId}" is not installed or has no UI.`
            : 'No plugin specified.'
          }
        </p>
      </div>
    </div>
  );
}

function PageNotFound({ pluginId, path }: { pluginId: string; path: string }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-12rem)]">
      <div className="rounded-xl bg-neutral-800/30 p-8 text-center max-w-md">
        <div className="w-12 h-12 rounded-full bg-neutral-700/50 flex items-center justify-center mx-auto mb-4">
          <Icons.fileQuestion className="w-6 h-6 text-neutral-400" />
        </div>
        <h1 className="text-xl font-semibold text-neutral-100 mb-2">Page Not Found</h1>
        <p className="text-sm text-neutral-400">
          The plugin "{pluginId}" does not have a page at "{path}".
        </p>
      </div>
    </div>
  );
}

function AccessDenied() {
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
