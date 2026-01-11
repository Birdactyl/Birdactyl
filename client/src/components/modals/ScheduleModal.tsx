import { useState, useEffect, useRef } from 'react';
import { createSchedule, updateSchedule, type Schedule, type ScheduleTask } from '../../lib/api';
import { notify, Modal, Input, Button, Icons, Checkbox, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '../';

const TASK_ACTIONS = [
    { value: 'command', label: 'Send Command' },
    { value: 'power', label: 'Power Action' },
    { value: 'delay', label: 'Delay' },
    { value: 'backup', label: 'Create Backup' },
] as const;

const POWER_OPTIONS = [
    { value: 'start', label: 'Start' },
    { value: 'stop', label: 'Stop' },
    { value: 'restart', label: 'Restart' },
    { value: 'kill', label: 'Kill' },
];

const CRON_PRESETS = [
    { value: '0 */5 * * * *', label: 'Every 5 minutes' },
    { value: '0 */15 * * * *', label: 'Every 15 minutes' },
    { value: '0 */30 * * * *', label: 'Every 30 minutes' },
    { value: '0 0 * * * *', label: 'Every hour' },
    { value: '0 0 */2 * * *', label: 'Every 2 hours' },
    { value: '0 0 */4 * * *', label: 'Every 4 hours' },
    { value: '0 0 */6 * * *', label: 'Every 6 hours' },
    { value: '0 0 */12 * * *', label: 'Every 12 hours' },
    { value: '0 0 0 * * *', label: 'Daily at midnight' },
    { value: '0 0 4 * * *', label: 'Daily at 4:00 AM' },
    { value: '0 0 12 * * *', label: 'Daily at noon' },
    { value: '0 0 0 * * 0', label: 'Every Sunday' },
    { value: '0 0 0 * * 1', label: 'Every Monday' },
    { value: '0 0 0 1 * *', label: 'First of the month' },
];

interface Props {
    serverId: string;
    schedule?: Schedule;
    open: boolean;
    onClose: () => void;
    onSaved?: (schedule: Schedule) => void;
}

export default function ScheduleModal({ serverId, schedule, open, onClose, onSaved }: Props) {
    const [name, setName] = useState('');
    const [cron, setCron] = useState('0 0 4 * * *');
    const [useCustomCron, setUseCustomCron] = useState(false);
    const [isActive, setIsActive] = useState(true);
    const [onlyWhenOnline, setOnlyWhenOnline] = useState(false);
    const [tasks, setTasks] = useState<ScheduleTask[]>([]);
    const [loading, setLoading] = useState(false);
    const submittingRef = useRef(false);

    useEffect(() => {
        if (open && schedule) {
            submittingRef.current = false;
            setLoading(false);
            setName(schedule.name);
            setCron(schedule.cron_expression);
            setUseCustomCron(!CRON_PRESETS.some(p => p.value === schedule.cron_expression));
            setIsActive(schedule.is_active);
            setOnlyWhenOnline(schedule.only_when_online);
            setTasks(schedule.tasks || []);
        } else if (open) {
            submittingRef.current = false;
            setLoading(false);
            setName('');
            setCron('0 0 4 * * *');
            setUseCustomCron(false);
            setIsActive(true);
            setOnlyWhenOnline(false);
            setTasks([]);
        }
    }, [open, schedule]);

    const addTask = () => setTasks([...tasks, { sequence: tasks.length + 1, action: 'command', payload: '' }]);
    const removeTask = (idx: number) => setTasks(tasks.filter((_, i) => i !== idx).map((t, i) => ({ ...t, sequence: i + 1 })));
    const updateTask = (idx: number, field: keyof ScheduleTask, value: string | number) => {
        const updated = [...tasks];
        updated[idx] = { ...updated[idx], [field]: value };
        if (field === 'action' && value === 'backup') updated[idx].payload = '';
        if (field === 'action' && value === 'power') updated[idx].payload = 'restart';
        setTasks(updated);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!name || !cron || submittingRef.current) return;
        submittingRef.current = true;
        setLoading(true);
        const data = { name, cron_expression: cron, is_active: isActive, only_when_online: onlyWhenOnline, tasks };
        const res = schedule
            ? await updateSchedule(serverId, schedule.id, data)
            : await createSchedule(serverId, data);
        if (res.success && res.data) {
            notify('Saved', schedule ? 'Schedule updated' : 'Schedule created', 'success');
            onSaved?.(res.data);
            onClose();
        } else {
            notify('Error', res.error || 'Failed to save', 'error');
            setLoading(false);
            submittingRef.current = false;
        }
    };

    const getActionLabel = (action: string) => TASK_ACTIONS.find(a => a.value === action)?.label || action;
    const getPowerLabel = (value: string) => POWER_OPTIONS.find(p => p.value === value)?.label || value;
    const getCronLabel = () => useCustomCron ? 'Custom' : (CRON_PRESETS.find(p => p.value === cron)?.label || 'Custom');

    return (
        <Modal open={open} onClose={onClose} title={schedule ? 'Edit Schedule' : 'Create Schedule'} description="Configure when and what actions to run.">
            <form onSubmit={handleSubmit} className="space-y-4">
                <Input label="Name" placeholder="Daily Restart" value={name} onChange={e => setName(e.target.value)} required />

                <div>
                    <label className="block text-xs font-medium text-neutral-400 mb-1.5">When to Run</label>
                    <DropdownMenu className="w-full">
                        <DropdownMenuTrigger asChild>
                            <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-sm rounded-lg border border-neutral-700 bg-neutral-800/80 text-neutral-100 hover:border-neutral-500 transition-colors">
                                {getCronLabel()}
                                <Icons.selector className="w-4 h-4 text-neutral-500" />
                            </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)] max-h-64 overflow-y-auto">
                            {CRON_PRESETS.map(p => (
                                <DropdownMenuItem key={p.value} onSelect={() => { setCron(p.value); setUseCustomCron(false); }} className={!useCustomCron && cron === p.value ? 'bg-neutral-700/50' : ''}>
                                    {p.label}
                                </DropdownMenuItem>
                            ))}
                            <DropdownMenuItem onSelect={() => setUseCustomCron(true)} className={useCustomCron ? 'bg-neutral-700/50' : ''}>
                                Custom...
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                    {useCustomCron && (
                        <div className="mt-2">
                            <Input placeholder="0 0 4 * * *" value={cron} onChange={e => setCron(e.target.value)} className="font-mono" />
                            <p className="text-xs text-neutral-500 mt-1">Format: second minute hour day month weekday</p>
                        </div>
                    )}
                    {!useCustomCron && (
                        <p className="text-xs text-neutral-500 mt-1.5 font-mono">{cron}</p>
                    )}
                </div>

                <div className="flex items-center gap-6">
                    <Checkbox checked={isActive} onChange={v => setIsActive(v ?? false)} label="Active" />
                    <Checkbox checked={onlyWhenOnline} onChange={v => setOnlyWhenOnline(v ?? false)} label="Only when online" />
                </div>

                <div className="space-y-3">
                    <div className="flex items-center justify-between">
                        <label className="text-xs font-medium text-neutral-400">Tasks</label>
                        <button type="button" onClick={addTask} className="text-xs text-neutral-400 hover:text-neutral-200 flex items-center gap-1">
                            <Icons.plus className="w-3.5 h-3.5" /> Add Task
                        </button>
                    </div>

                    {tasks.length === 0 ? (
                        <div className="text-sm text-neutral-500 text-center py-6 border border-dashed border-neutral-700 rounded-lg">
                            No tasks yet. Add a task to get started.
                        </div>
                    ) : (
                        <div className="space-y-3 max-h-64 overflow-y-auto pr-1">
                            {tasks.map((task, idx) => (
                                <div key={idx} className="p-3 rounded-lg bg-neutral-800/50 border border-neutral-700 space-y-3">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                            <span className="text-xs font-medium text-neutral-500">Task {task.sequence}</span>
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <button type="button" className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md border border-neutral-600 bg-neutral-900 text-neutral-200 hover:border-neutral-500 transition-colors">
                                                        {getActionLabel(task.action)}
                                                        <Icons.selector className="w-4 h-4 text-neutral-500" />
                                                    </button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent>
                                                    {TASK_ACTIONS.map(a => (
                                                        <DropdownMenuItem key={a.value} onSelect={() => updateTask(idx, 'action', a.value)} className={task.action === a.value ? 'bg-neutral-700/50' : ''}>
                                                            {a.label}
                                                        </DropdownMenuItem>
                                                    ))}
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </div>
                                        <button type="button" onClick={() => removeTask(idx)} className="text-neutral-500 hover:text-red-400 p-1.5 rounded hover:bg-neutral-700/50 transition-colors">
                                            <Icons.trash className="w-4 h-4" />
                                        </button>
                                    </div>

                                    {task.action === 'command' && (
                                        <Input placeholder="say Server restarting in 60 seconds!" value={task.payload} onChange={e => updateTask(idx, 'payload', e.target.value)} />
                                    )}

                                    {task.action === 'power' && (
                                        <DropdownMenu className="w-full">
                                            <DropdownMenuTrigger asChild>
                                                <button type="button" className="w-full flex items-center justify-between px-3 py-2 text-sm rounded-md border border-neutral-600 bg-neutral-900 text-neutral-200 hover:border-neutral-500 transition-colors">
                                                    {getPowerLabel(task.payload)}
                                                    <Icons.selector className="w-4 h-4 text-neutral-500" />
                                                </button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent className="!min-w-0 w-[var(--trigger-width)]">
                                                {POWER_OPTIONS.map(p => (
                                                    <DropdownMenuItem key={p.value} onSelect={() => updateTask(idx, 'payload', p.value)} className={task.payload === p.value ? 'bg-neutral-700/50' : ''}>
                                                        {p.label}
                                                    </DropdownMenuItem>
                                                ))}
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    )}

                                    {task.action === 'delay' && (
                                        <div className="flex items-center gap-3">
                                            <Input type="number" placeholder="30" value={task.payload} onChange={e => updateTask(idx, 'payload', e.target.value)} className="w-24" />
                                            <span className="text-sm text-neutral-400">seconds</span>
                                        </div>
                                    )}

                                    {task.action === 'backup' && (
                                        <p className="text-sm text-neutral-400">This task will create a server backup.</p>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                <div className="flex justify-end gap-3 pt-4">
                    <Button variant="ghost" onClick={onClose} disabled={loading}>Cancel</Button>
                    <Button type="submit" loading={loading} disabled={!name || loading}>{schedule ? 'Save' : 'Create'}</Button>
                </div>
            </form>
        </Modal>
    );
}
