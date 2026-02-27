import { useEffect, useState, useRef, ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { getReadyPromise, resetLoader, isPageLoading } from '../../lib/pageLoader';

export default function PageTransition({ children }: { children: ReactNode }) {
  const location = useLocation();
  const [isVisible, setIsVisible] = useState(true);
  const [content, setContent] = useState<ReactNode>(children);
  const isFirstRender = useRef(true);
  const prevPath = useRef(location.pathname);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }

    if (prevPath.current === location.pathname) {
      return;
    }
    prevPath.current = location.pathname;

    setIsVisible(false);
    resetLoader();

    const transition = async () => {
      await new Promise(r => setTimeout(r, 50));
      
      const checkReady = async () => {
        const maxWait = Date.now() + 10000;
        while (Date.now() < maxWait) {
          if (!isPageLoading()) {
            const promise = getReadyPromise();
            if (!promise) break;
          }
          await new Promise(r => setTimeout(r, 16));
          const promise = getReadyPromise();
          if (promise) {
            await Promise.race([promise, new Promise(r => setTimeout(r, 10000))]);
            break;
          }
        }
      };
      
      await checkReady();
      
      setContent(children);
      requestAnimationFrame(() => {
        setIsVisible(true);
      });
    };

    transition();
  }, [location.pathname, children]);

  return (
    <div
      style={{
        transition: 'opacity 250ms cubic-bezier(0.68, -0.55, 0.27, 1.55), transform 250ms cubic-bezier(0.68, -0.55, 0.27, 1.55)',
        opacity: isVisible ? 1 : 0,
        transform: isVisible ? 'translateY(0) scale(1)' : 'translateY(12px) scale(0.98)',
      }}
    >
      {content}
    </div>
  );
}
