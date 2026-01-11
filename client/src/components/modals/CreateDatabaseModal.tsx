import { useState, useEffect, useRef } from 'react';
import { getDatabaseHosts, createServerDatabase, DatabaseHost, ServerDatabase } from '../../lib/api';
import { notify } from '../feedback/Notification';
import { Modal, Input, Button, Icons, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../';

interface Props {
  serverId: string;
  open: boolean;
  onClose: () => void;
  onCreated?: (db: ServerDatabase) => void;
}

export default function CreateDatabaseModal({ serverId, open, onClose, onCreated }: Props) {
  const [hosts, setHosts] = useState<DatabaseHost[]>([]);
  const [selectedHost, setSelectedHost] = useState<DatabaseHost | null>(null);
  const [dbName, setDbName] = useState('');
  const [loading, setLoading] = useState(false);
  const submittingRef = useRef(false);

  useEffect(() => {
    if (open) {
      submittingRef.current = false;
      setLoading(false);
      getDatabaseHosts(serverId).then(res => {
        if (res.success && res.data) {
          setHosts(res.data);
          const available = res.data.find(h => h.max_databases === 0 || h.used < h.max_databases);
          if (available) setSelectedHost(available);
        }
      });
      setDbName('');
    }
  }, [open, serverId]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
  };

  const handleCreate = async () => {
    if (!selectedHost || submittingRef.current) return;
    submittingRef.current = true;
    setLoading(true);
    const res = await createServerDatabase(serverId, selectedHost.id, dbName || undefined);
    if (res.success && res.data) {
      notify('Success', 'Database created', 'success');
      onCreated?.(res.data);
      onClose();
    } else {
      notify('Error', res.error || 'Failed to create database', 'error');
      setLoading(false);
      submittingRef.current = false;
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Create Database" description="Create a new MySQL database for this server.">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="w-full">
          <label className="block text-xs font-medium text-neutral-400 mb-1.5">Database Host</label>
          {hosts.length === 0 ? <p className="text-neutral-500 text-sm">No database hosts available.</p> : (
            <DropdownMenu className="w-full block">
              <DropdownMenuTrigger asChild>
                <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-[13px] rounded-lg border border-neutral-800 bg-neutral-800/80 text-left hover:border-neutral-500 transition-colors">
                  <span className="truncate text-neutral-100">{selectedHost ? `${selectedHost.name} (${selectedHost.host}:${selectedHost.port})` : 'Select a database host'}</span>
                  <Icons.selector className="w-4 h-4 text-neutral-500 flex-shrink-0 ml-2" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
                {hosts.map(h => {
                  const full = h.max_databases > 0 && h.used >= h.max_databases;
                  return (
                    <DropdownMenuItem key={h.id} onSelect={() => !full && setSelectedHost(h)} disabled={full} className={selectedHost?.id === h.id ? 'bg-neutral-700/50' : ''}>
                      <span className="truncate flex-1">{h.name} ({h.host}:{h.port}) - {h.max_databases > 0 ? `${h.used}/${h.max_databases}` : h.used} used</span>
                      {selectedHost?.id === h.id && <Icons.check className="w-4 h-4 text-neutral-400 flex-shrink-0" />}
                    </DropdownMenuItem>
                  );
                })}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
        <Input label="Database Name (optional)" value={dbName} onChange={e => setDbName(e.target.value)} placeholder="default" />
        <p className="text-xs text-neutral-500">Leave empty for "default". Will be prefixed with server ID.</p>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
          <Button onClick={handleCreate} loading={loading} disabled={!selectedHost || hosts.length === 0 || loading}>Create</Button>
        </div>
      </form>
    </Modal>
  );
}
