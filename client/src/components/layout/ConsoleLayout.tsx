import { useState, useEffect } from 'react';
import Sidebar from './Sidebar';
import PageTransition from './PageTransition';
import { Icons } from '../Icons';

export default function ConsoleLayout({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [sidebarVisible, setSidebarVisible] = useState(true);
  const [isAnimating, setIsAnimating] = useState(false);

  useEffect(() => {
    if (sidebarOpen) {
      setSidebarVisible(true);
      requestAnimationFrame(() => {
        requestAnimationFrame(() => setIsAnimating(false));
      });
    } else {
      setIsAnimating(true);
      const timer = setTimeout(() => setSidebarVisible(false), 300);
      return () => clearTimeout(timer);
    }
  }, [sidebarOpen]);

  const handleClose = () => {
    setIsAnimating(true);
    setSidebarOpen(false);
  };

  const handleOpen = () => {
    setSidebarOpen(true);
  };

  return (
    <div className="w-screen h-screen bg-black p-1 pl-0 overflow-hidden">
      {sidebarVisible && (
        <div
          className={`fixed inset-y-0 left-0 z-40 w-64 transition-transform duration-300 ease-out ${
            sidebarOpen && !isAnimating ? 'translate-x-0' : '-translate-x-full'
          }`}
        >
          <Sidebar onClose={handleClose} />
        </div>
      )}
      
      <button
        onClick={handleOpen}
        className={`fixed left-0 top-1/2 -translate-y-1/2 z-[60] h-20 w-[20px] md:w-[18px] rounded-r-lg border border-neutral-600 border-l-0 bg-neutral-800/95 backdrop-blur-sm transition-all duration-300 ease-out hover:w-[24px] md:hover:w-[22px] shadow-lg hover:shadow-xl focus:outline-none cursor-pointer hover:bg-neutral-700 hover:border-neutral-500 focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-neutral-950 group flex items-center justify-center touch-pan-y ${
          sidebarOpen ? 'opacity-0 pointer-events-none -translate-x-full' : 'opacity-100 translate-x-0'
        }`}
        aria-label="Open sidebar"
        type="button"
      >
        <Icons.chevronRight className="h-4 w-4 text-neutral-400 group-hover:text-neutral-100 transition-colors opacity-70 group-hover:opacity-100" />
        <span className="sr-only">Open sidebar</span>
      </button>

      <main 
        className="h-full pr-1 py-1 transition-all duration-300 ease-out overflow-x-auto" 
        style={{ marginLeft: sidebarOpen ? '252px' : '0', paddingLeft: sidebarOpen ? '0' : '4px', minWidth: sidebarOpen ? 'calc(320px + 252px)' : '320px' }}
      >
        <div className="w-full h-full bg-neutral-950 rounded-lg overflow-auto px-6 sm:px-8 py-6">
          <PageTransition>{children}</PageTransition>
        </div>
      </main>
    </div>
  );
}
