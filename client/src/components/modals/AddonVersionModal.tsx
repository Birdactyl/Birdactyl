import { useState, useEffect } from 'react';
import { getAddonVersions, installAddon, type Addon, type AddonVersion } from '../../lib/api';
import { notify, Modal, Button } from '../';

interface Props {
  serverId: string;
  sourceId: string;
  addon: Addon;
  open: boolean;
  onClose: () => void;
  onInstalled?: () => void;
}

export default function AddonVersionModal({ serverId, sourceId, addon, open, onClose, onInstalled }: Props) {
  const [versions, setVersions] = useState<AddonVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [installing, setInstalling] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    getAddonVersions(serverId, sourceId, addon.id).then(res => {
      if (res.success && res.data) setVersions(res.data);
      setLoading(false);
    });
  }, [open, serverId, sourceId, addon.id]);

  const handleInstall = async (version: AddonVersion) => {
    setInstalling(version.id);
    const res = await installAddon(serverId, sourceId, version.download_url, version.file_name || `${addon.name}.jar`);
    if (res.success) {
      notify('Installed', `${addon.name} has been installed`, 'success');
      onInstalled?.();
      onClose();
    } else {
      notify('Error', res.error || 'Failed to install', 'error');
    }
    setInstalling(null);
  };

  return (
    <Modal open={open} onClose={onClose} title={`Install ${addon.name}`} description="Select a version to install">
      <div className="space-y-2 max-h-64 overflow-y-auto">
        {loading ? (
          <div className="text-sm text-neutral-400 py-4 text-center">Loading versions...</div>
        ) : versions.length === 0 ? (
          <div className="text-sm text-neutral-400 py-4 text-center">No versions available</div>
        ) : (
          versions.slice(0, 10).map(v => (
            <div key={v.id} className="flex items-center justify-between p-2 rounded-lg hover:bg-neutral-800/50">
              <span className="text-sm text-neutral-200">{v.name}</span>
              <Button variant="ghost" onClick={() => handleInstall(v)} disabled={!!installing} loading={installing === v.id}>Install</Button>
            </div>
          ))
        )}
      </div>
    </Modal>
  );
}
