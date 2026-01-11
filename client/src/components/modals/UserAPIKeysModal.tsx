import { useState, useEffect } from 'react';
import { adminGetUserAPIKeys, adminCreateUserAPIKey, adminDeleteUserAPIKey, type APIKey, type APIKeyCreated } from '../../lib/api';
import { notify, Modal, Input, Button, Icons, Table } from '../';

interface User {
    id: string;
    username: string;
    email: string;
    is_admin: boolean;
    is_root_admin: boolean;
}

interface Props {
    open: boolean;
    user: User | null;
    onClose: () => void;
}

export default function UserAPIKeysModal({ open, user, onClose }: Props) {
    const [keys, setKeys] = useState<APIKey[]>([]);
    const [loading, setLoading] = useState(false);
    const [createForm, setCreateForm] = useState({ open: false, name: '', expiresIn: '', loading: false });
    const [newKey, setNewKey] = useState<APIKeyCreated | null>(null);
    const [deleteKey, setDeleteKey] = useState<{ key: APIKey; loading: boolean } | null>(null);

    const loadKeys = async () => {
        if (!user) return;
        setLoading(true);
        const res = await adminGetUserAPIKeys(user.id);
        if (res.success && res.data) setKeys(res.data);
        setLoading(false);
    };

    useEffect(() => {
        if (open && user) loadKeys();
        if (!open) {
            setKeys([]);
            setCreateForm({ open: false, name: '', expiresIn: '', loading: false });
            setNewKey(null);
            setDeleteKey(null);
        }
    }, [open, user]);

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;
        setCreateForm(f => ({ ...f, loading: true }));
        const expiresIn = createForm.expiresIn ? parseInt(createForm.expiresIn) : undefined;
        const res = await adminCreateUserAPIKey(user.id, createForm.name || 'API Key', expiresIn);
        if (res.success && res.data) {
            setNewKey(res.data);
            setCreateForm({ open: false, name: '', expiresIn: '', loading: false });
            loadKeys();
        } else {
            notify('Error', res.error || 'Failed to create API key', 'error');
            setCreateForm(f => ({ ...f, loading: false }));
        }
    };

    const handleDelete = async () => {
        if (!user || !deleteKey) return;
        setDeleteKey(d => d && { ...d, loading: true });
        const res = await adminDeleteUserAPIKey(user.id, deleteKey.key.id);
        if (res.success) {
            notify('Deleted', 'API key deleted', 'success');
            setDeleteKey(null);
            loadKeys();
        } else {
            notify('Error', res.error || 'Failed to delete', 'error');
            setDeleteKey(d => d && { ...d, loading: false });
        }
    };

    const copyKey = () => {
        if (newKey) {
            navigator.clipboard.writeText(newKey.key);
            notify('Copied', 'API key copied to clipboard', 'success');
        }
    };

    const columns = [
        { key: 'name', header: 'Name', render: (k: APIKey) => <span className="text-sm font-medium text-neutral-100">{k.name}</span> },
        { key: 'prefix', header: 'Key', render: (k: APIKey) => <span className="text-sm font-mono text-neutral-400">{k.key_prefix}...</span> },
        { key: 'expires', header: 'Expires', render: (k: APIKey) => <span className="text-sm text-neutral-400">{k.expires_at ? new Date(k.expires_at).toLocaleDateString() : 'Never'}</span> },
        { key: 'created', header: 'Created', render: (k: APIKey) => <span className="text-sm text-neutral-400">{new Date(k.created_at).toLocaleDateString()}</span> },
        {
            key: 'actions', header: '', align: 'right' as const, render: (k: APIKey) => (
                <button onClick={() => setDeleteKey({ key: k, loading: false })} className="text-xs text-red-400 hover:text-red-300">Delete</button>
            )
        },
    ];

    const subModalOpen = !!newKey || !!deleteKey || createForm.open;

    return (
        <>
            <Modal open={open} onClose={onClose} title={`API Keys - ${user?.username || ''}`} description="Manage API keys for this user." className="max-w-2xl">
                <div className={`space-y-4 ${subModalOpen ? 'opacity-0 pointer-events-none h-0 overflow-hidden' : ''}`}>
                    <div className="flex items-center justify-between">
                        <span className="text-xs text-neutral-400">{keys.length} API key{keys.length !== 1 ? 's' : ''}</span>
                        <Button onClick={() => setCreateForm(f => ({ ...f, open: true }))}><Icons.plus className="w-4 h-4" />Create Key</Button>
                    </div>
                    <div className="bg-neutral-900/40 rounded-lg p-1">
                        <Table columns={columns} data={keys} keyField="id" loading={loading} emptyText="No API keys" />
                    </div>
                    <div className="flex justify-end pt-2">
                        <Button variant="ghost" onClick={onClose}>Close</Button>
                    </div>
                </div>
            </Modal>

            <Modal open={createForm.open} onClose={() => !createForm.loading && setCreateForm(f => ({ ...f, open: false }))} title="Create API Key" description={`Generate a new API key for ${user?.username}.`}>
                <form onSubmit={handleCreate} className="space-y-4">
                    <Input label="Name" placeholder="My CLI Key" value={createForm.name} onChange={e => setCreateForm(f => ({ ...f, name: e.target.value }))} />
                    <Input label="Expires In (days)" placeholder="Leave blank for never" type="number" min={1} value={createForm.expiresIn} onChange={e => setCreateForm(f => ({ ...f, expiresIn: e.target.value }))} />
                    <div className="flex justify-end gap-3 pt-4">
                        <Button variant="ghost" onClick={() => setCreateForm(f => ({ ...f, open: false }))} disabled={createForm.loading}>Cancel</Button>
                        <Button type="submit" loading={createForm.loading}>Create</Button>
                    </div>
                </form>
            </Modal>

            <Modal open={!!newKey} onClose={() => setNewKey(null)} title="API Key Created" description="Copy this key now. You won't be able to see it again!">
                <div className="space-y-4">
                    <div className="p-3 bg-neutral-900 rounded-lg font-mono text-sm text-emerald-400 break-all">{newKey?.key}</div>
                    <div className="flex justify-end gap-3">
                        <Button onClick={copyKey}><Icons.clipboard className="w-4 h-4" />Copy</Button>
                        <Button variant="ghost" onClick={() => setNewKey(null)}>Done</Button>
                    </div>
                </div>
            </Modal>

            <Modal open={!!deleteKey} onClose={() => !deleteKey?.loading && setDeleteKey(null)} title="Delete API Key" description={`Delete "${deleteKey?.key.name}"? This cannot be undone.`}>
                <div className="flex justify-end gap-3 pt-4">
                    <Button variant="ghost" onClick={() => setDeleteKey(null)} disabled={deleteKey?.loading}>Cancel</Button>
                    <Button onClick={handleDelete} loading={deleteKey?.loading} variant="danger">Delete</Button>
                </div>
            </Modal>
        </>
    );
}
