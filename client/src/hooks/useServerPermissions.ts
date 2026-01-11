import { useState, useEffect } from 'react';
import { getServerPermissions } from '../lib/api';

export function useServerPermissions(serverId: string | undefined) {
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [isOwner, setIsOwner] = useState(false);

  useEffect(() => {
    if (!serverId) return;
    getServerPermissions(serverId).then(res => {
      if (res.success && res.data) {
        setPermissions(res.data);
        setIsOwner(res.data.includes('*'));
      }
      setLoading(false);
    });
  }, [serverId]);

  const can = (perm: string) => permissions.includes('*') || permissions.includes(perm);

  return { permissions, loading, can, isOwner };
}
