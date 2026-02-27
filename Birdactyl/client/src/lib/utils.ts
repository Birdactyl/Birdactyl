export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024, sizes = ['B', 'KiB', 'MiB', 'GiB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
};

export const formatDate = (date: number | string): string => {
  const d = typeof date === 'number' ? new Date(date * 1000) : new Date(date);
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
};

export const parseUserAgent = (ua: string): string => {
  if (ua.includes('Chrome')) return 'Chrome';
  if (ua.includes('Firefox')) return 'Firefox';
  if (ua.includes('Safari')) return 'Safari';
  if (ua.includes('Edge')) return 'Edge';
  return 'Unknown Browser';
};
