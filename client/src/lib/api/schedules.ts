import { api } from './client';

export interface ScheduleTask {
    sequence: number;
    action: 'command' | 'power' | 'delay' | 'backup';
    payload: string;
}

export interface Schedule {
    id: string;
    server_id: string;
    name: string;
    cron_expression: string;
    is_active: boolean;
    only_when_online: boolean;
    last_run_at: string | null;
    next_run_at: string | null;
    tasks: ScheduleTask[];
    created_at: string;
    updated_at: string;
}

export interface CreateScheduleRequest {
    name: string;
    cron_expression: string;
    is_active: boolean;
    only_when_online: boolean;
    tasks: ScheduleTask[];
}

export const getSchedules = (serverId: string) => api.get<Schedule[]>(`/servers/${serverId}/schedules`);
export const getSchedule = (serverId: string, scheduleId: string) => api.get<Schedule>(`/servers/${serverId}/schedules/${scheduleId}`);
export const createSchedule = (serverId: string, data: CreateScheduleRequest) => api.post<Schedule>(`/servers/${serverId}/schedules`, data);
export const updateSchedule = (serverId: string, scheduleId: string, data: CreateScheduleRequest) => api.put<Schedule>(`/servers/${serverId}/schedules/${scheduleId}`, data);
export const deleteSchedule = (serverId: string, scheduleId: string) => api.delete<void>(`/servers/${serverId}/schedules/${scheduleId}`);
export const runScheduleNow = (serverId: string, scheduleId: string) => api.post<void>(`/servers/${serverId}/schedules/${scheduleId}/run`);
