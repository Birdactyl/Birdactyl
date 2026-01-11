import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { getServer, getSchedules, runScheduleNow, type Server, type Schedule, type ScheduleTask } from '../../../lib/api';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { Button, Icons, Table, PermissionDenied, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../../../components';
import { ScheduleModal, DeleteScheduleModal } from '../../../components/modals';
import { notify } from '../../../components/feedback/Notification';

export default function SchedulesPage() {
    const { id } = useParams<{ id: string }>();
    const [server, setServer] = useState<Server | null>(null);
    const [schedules, setSchedules] = useState<Schedule[]>([]);
    const [loading, setLoading] = useState(true);
    const [expanded, setExpanded] = useState<Set<string>>(new Set());
    const [editModal, setEditModal] = useState<{ open: boolean; schedule?: Schedule }>({ open: false });
    const [deleteModal, setDeleteModal] = useState<Schedule | null>(null);
    const { can, loading: permsLoading } = useServerPermissions(id);

    useEffect(() => {
        if (!id) return;
        getServer(id).then(res => res.success && res.data && setServer(res.data));
        loadSchedules();
    }, [id]);

    const loadSchedules = async () => {
        if (!id) return;
        const res = await getSchedules(id);
        if (res.success && res.data) setSchedules(res.data);
        setLoading(false);
    };

    const handleRunNow = async (schedule: Schedule) => {
        if (!id) return;
        const res = await runScheduleNow(id, schedule.id);
        if (res.success) {
            notify('Started', `Running "${schedule.name}"`, 'success');
        } else {
            notify('Error', res.error || 'Failed to run', 'error');
        }
    };

    const toggleExpand = (scheduleId: string) => setExpanded(prev => { const next = new Set(prev); next.has(scheduleId) ? next.delete(scheduleId) : next.add(scheduleId); return next; });
    const formatDateTime = (ts: string | null) => ts ? new Date(ts).toLocaleString() : '—';

    const getTaskLabel = (task: ScheduleTask) => {
        switch (task.action) {
            case 'command': return `Run: ${task.payload}`;
            case 'power': return `Power: ${task.payload}`;
            case 'delay': return `Wait ${task.payload}s`;
            case 'backup': return 'Create backup';
            default: return task.action;
        }
    };

    if (permsLoading || loading) return null;
    if (!can('schedule.list')) return <PermissionDenied message="You don't have permission to view schedules" />;

    const columns = [
        {
            key: 'expand', header: '', className: 'w-8', render: (s: Schedule) => (
                <button onClick={() => toggleExpand(s.id)} className="text-neutral-400 hover:text-neutral-200 transition">
                    <Icons.chevronRight className={`w-4 h-4 transition-transform ${expanded.has(s.id) ? 'rotate-90' : ''}`} />
                </button>
            )
        },
        {
            key: 'name', header: 'Name', render: (s: Schedule) => (
                <div className="flex items-center gap-3">
                    <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${s.is_active ? 'bg-emerald-500/20' : 'bg-neutral-700/50'}`}>
                        <Icons.clock className={`w-4 h-4 ${s.is_active ? 'text-emerald-400' : 'text-neutral-500'}`} />
                    </div>
                    <div>
                        <div className="text-sm font-medium text-neutral-100">{s.name}</div>
                        <div className="text-xs text-neutral-500 font-mono">{s.cron_expression}</div>
                    </div>
                </div>
            )
        },
        {
            key: 'status', header: 'Status', render: (s: Schedule) => (
                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${s.is_active ? 'bg-emerald-500/20 text-emerald-400' : 'bg-neutral-700 text-neutral-400'}`}>
                    {s.is_active ? 'Active' : 'Inactive'}
                </span>
            )
        },
        { key: 'lastRun', header: 'Last Run', render: (s: Schedule) => <span className="text-sm text-neutral-400">{formatDateTime(s.last_run_at)}</span> },
        { key: 'nextRun', header: 'Next Run', render: (s: Schedule) => <span className="text-sm text-neutral-400">{formatDateTime(s.next_run_at)}</span> },
        {
            key: 'actions', header: '', align: 'right' as const, render: (s: Schedule) => (
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        {can('schedule.run') && <DropdownMenuItem onSelect={() => handleRunNow(s)}>Run Now</DropdownMenuItem>}
                        {can('schedule.update') && <DropdownMenuItem onSelect={() => setEditModal({ open: true, schedule: s })}>Edit</DropdownMenuItem>}
                        {can('schedule.delete') && <DropdownMenuItem onSelect={() => setDeleteModal(s)} className="text-red-400">Delete</DropdownMenuItem>}
                    </DropdownMenuContent>
                </DropdownMenu>
            )
        },
    ];

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-1 text-sm text-neutral-400">
                <span className="font-medium text-neutral-200">{server?.name || 'Server'}</span>
                <span className="text-neutral-400">/</span>
                <span className="font-semibold text-neutral-100">Schedules</span>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h1 className="text-xl font-semibold text-neutral-100">Schedules</h1>
                    <p className="text-sm text-neutral-400">Automate tasks with scheduled actions.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="ghost" onClick={loadSchedules}><Icons.refresh className="h-4 w-4" /></Button>
                    {can('schedule.create') && <Button onClick={() => setEditModal({ open: true })}><Icons.plus className="h-4 w-4 mr-1.5" />New Schedule</Button>}
                </div>
            </div>

            <div className="rounded-xl bg-neutral-800/30">
                <div className="px-4 py-2 text-xs text-neutral-400">{schedules.length} schedule{schedules.length !== 1 ? 's' : ''}</div>
                <div className="bg-neutral-900/40 rounded-lg p-1">
                    <Table
                        columns={columns}
                        data={schedules}
                        keyField="id"
                        emptyText="No schedules yet"
                        expandable={{
                            isExpanded: s => expanded.has(s.id),
                            render: s => (
                                <div className="space-y-3 py-2">
                                    <div className="flex items-center gap-4 text-xs">
                                        {s.only_when_online && <span className="text-amber-400">Only runs when server is online</span>}
                                        <span className="text-neutral-500">{s.tasks?.length || 0} task{(s.tasks?.length || 0) !== 1 ? 's' : ''}</span>
                                    </div>
                                    {s.tasks && s.tasks.length > 0 && (
                                        <div className="flex flex-wrap gap-2">
                                            {s.tasks.map((task, i) => (
                                                <span key={i} className="inline-flex items-center gap-1 px-2 py-1 rounded bg-neutral-800 text-xs text-neutral-300">
                                                    <span className="text-neutral-500">{task.sequence}.</span>
                                                    {getTaskLabel(task)}
                                                </span>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            )
                        }}
                    />
                </div>
            </div>

            <ScheduleModal
                serverId={id!}
                schedule={editModal.schedule}
                open={editModal.open}
                onClose={() => setEditModal({ open: false })}
                onSaved={() => loadSchedules()}
            />

            <DeleteScheduleModal
                serverId={id!}
                schedule={deleteModal}
                open={!!deleteModal}
                onClose={() => setDeleteModal(null)}
                onDeleted={scheduleId => setSchedules(s => s.filter(x => x.id !== scheduleId))}
            />
        </div>
    );
}
