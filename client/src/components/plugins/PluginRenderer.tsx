import React, { Suspense, ComponentType } from 'react';
import { useNavigate } from 'react-router-dom';
import { getPluginComponent } from '../../lib/pluginLoader';
import { createPluginAPI, PluginAPIProvider } from '../../lib/pluginHost';
import { Icons } from '../Icons';

interface PluginRendererProps {
  pluginId: string;
  component: string;
  props?: Record<string, unknown>;
}

function PluginError({ pluginId, component, error }: { pluginId: string; component: string; error?: string }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[200px] p-8">
      <div className="rounded-xl bg-neutral-800/30 p-8 text-center max-w-md">
        <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center mx-auto mb-4">
          <Icons.warning className="w-6 h-6 text-red-400" />
        </div>
        <h3 className="text-lg font-semibold text-neutral-100 mb-2">Plugin Error</h3>
        <p className="text-sm text-neutral-400 mb-2">
          Failed to load component "{component}" from plugin "{pluginId}"
        </p>
        {error && <p className="text-xs text-red-400 font-mono">{error}</p>}
      </div>
    </div>
  );
}

function PluginLoading() {
  return (
    <div className="flex items-center justify-center min-h-[200px]">
      <div className="flex items-center gap-3 text-neutral-400">
        <div className="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin" />
        <span>Loading plugin...</span>
      </div>
    </div>
  );
}

export default function PluginRenderer({ pluginId, component, props = {} }: PluginRendererProps) {
  const navigate = useNavigate();
  
  const Component = getPluginComponent(pluginId, component) as ComponentType<Record<string, unknown>> | null;
  
  if (!Component) {
    return <PluginError pluginId={pluginId} component={component} />;
  }

  const pluginApi = createPluginAPI(pluginId, navigate);

  return (
    <PluginAPIProvider value={pluginApi}>
      <Suspense fallback={<PluginLoading />}>
        <PluginErrorBoundary pluginId={pluginId} component={component}>
          <Component {...props} />
        </PluginErrorBoundary>
      </Suspense>
    </PluginAPIProvider>
  );
}

interface ErrorBoundaryProps {
  pluginId: string;
  component: string;
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: string;
}

class PluginErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error: error.message };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error(`[plugin:${this.props.pluginId}] Component ${this.props.component} crashed:`, error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <PluginError 
          pluginId={this.props.pluginId} 
          component={this.props.component} 
          error={this.state.error} 
        />
      );
    }

    return this.props.children;
  }
}
