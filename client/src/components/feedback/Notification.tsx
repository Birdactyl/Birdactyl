import { useEffect, useState } from 'react';
import { Icons } from '../Icons';

export type NotificationType = 'error' | 'success' | 'info';

export interface NotificationData {
  id: string;
  title: string;
  message: string;
  type?: NotificationType;
}

interface NotificationItemProps extends NotificationData {
  onClose: (id: string) => void;
}

const typeStyles: Record<NotificationType, string> = {
  error: 'text-red-400',
  success: 'text-emerald-400',
  info: 'text-sky-400',
};

function NotificationItem({ id, title, message, type = 'error', onClose }: NotificationItemProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [isLeaving, setIsLeaving] = useState(false);
  const styles = typeStyles[type];

  useEffect(() => {
    requestAnimationFrame(() => setIsVisible(true));
    const timer = setTimeout(() => handleClose(), 5000);
    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setIsLeaving(true);
    setTimeout(() => onClose(id), 300);
  };

  return (
    <div
      className={`relative overflow-hidden rounded-xl border border-neutral-800 bg-neutral-900 shadow-2xl px-4 py-3 flex items-center gap-3 pointer-events-auto transition-all duration-300 ease-out ${
        isVisible && !isLeaving ? 'opacity-100 translate-y-0 scale-100' : 'opacity-0 -translate-y-4 scale-95'
      }`}
      role="status"
      aria-live="polite"
    >
      <div className={`flex items-center justify-center ${styles}`}>
        {type === 'error' && <Icons.errorCircle className="w-5 h-5" />}
        {type === 'success' && <Icons.successCircle className="w-5 h-5" />}
        {type === 'info' && <Icons.infoCircle className="w-5 h-5" />}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-semibold leading-tight truncate text-neutral-100">{title}</div>
        <div className="text-xs text-neutral-400 truncate">{message}</div>
      </div>
      <button
        type="button"
        aria-label="Close alert"
        onClick={handleClose}
        className="shrink-0 ml-2 inline-flex h-6 w-6 items-center justify-center rounded-lg text-neutral-500 hover:text-neutral-200 hover:bg-neutral-800 transition"
      >
        <Icons.xFilled className="w-4 h-4" />
      </button>
    </div>
  );
}

interface NotificationContainerProps {
  notifications: NotificationData[];
  onClose: (id: string) => void;
}

export function NotificationContainer({ notifications, onClose }: NotificationContainerProps) {
  return (
    <div className="fixed top-4 left-1/2 -translate-x-1/2 z-[10000] pointer-events-none">
      <div className="relative">
        {notifications.map((notification, index) => {
          const reverseIndex = notifications.length - 1 - index;
          const scale = 1 - reverseIndex * 0.05;
          const translateY = reverseIndex * 8;
          const opacity = 1 - reverseIndex * 0.2;
          const blur = reverseIndex * 2;

          return (
            <div
              key={notification.id}
              className="transition-all duration-300"
              style={{
                position: index === notifications.length - 1 ? 'relative' : 'absolute',
                top: 0,
                left: '50%',
                transform: `translateX(-50%) translateY(${translateY}px) scale(${scale})`,
                opacity: Math.max(opacity, 0.3),
                filter: `blur(${blur}px)`,
                zIndex: notifications.length - reverseIndex,
              }}
            >
              <NotificationItem {...notification} onClose={onClose} />
            </div>
          );
        })}
      </div>
    </div>
  );
}

let notificationId = 0;
let addNotificationFn: ((notification: Omit<NotificationData, 'id'>) => void) | null = null;

export function setNotificationHandler(handler: (notification: Omit<NotificationData, 'id'>) => void) {
  addNotificationFn = handler;
}

export function notify(title: string, message: string, type: NotificationType = 'error') {
  if (addNotificationFn) {
    addNotificationFn({ title, message, type });
  }
}

export function useNotifications() {
  const [notifications, setNotifications] = useState<NotificationData[]>([]);

  useEffect(() => {
    setNotificationHandler((notification) => {
      const id = `notification-${++notificationId}`;
      setNotifications((prev) => [...prev, { ...notification, id }]);
    });
    return () => setNotificationHandler(() => {});
  }, []);

  const removeNotification = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  return { notifications, removeNotification };
}
