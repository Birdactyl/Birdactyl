import Topbar from './Topbar';
import PageTransition from './PageTransition';
import { SubNavProvider } from './SubNavigation';

export default function ConsoleLayout({ children }: { children: React.ReactNode }) {
  return (
    <SubNavProvider>
      <div className="h-screen flex flex-col bg-[#0a0a0a] overflow-hidden" style={{ height: '100dvh' }}>
        <Topbar />
        <main className="flex-1 overflow-y-auto custom-scrollbar" style={{ scrollbarGutter: 'stable' }}>
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-4 sm:py-6">
            <PageTransition>{children}</PageTransition>
          </div>
        </main>
      </div>
    </SubNavProvider>
  );
}
