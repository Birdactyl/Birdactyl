import { request, ParsedResponse } from './client';
import type { Server } from './servers';
import type { Node } from './admin';
import type { Package } from './packages';

export interface Mount {
    id: string;
    name: string;
    description: string;
    source: string;
    target: string;
    read_only: boolean;
    user_mountable: boolean;
    navigable: boolean;
    created_at: string;
    updated_at: string;
    servers?: Server[];
    nodes?: Node[];
    packages?: Package[];
}

export interface PaginatedMounts {
    mounts: Mount[];
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
}

export async function adminGetMounts(page = 1, perPage = 20, search = ''): Promise<ParsedResponse<PaginatedMounts>> {
    const params = new URLSearchParams({
        page: page.toString(),
        per_page: perPage.toString()
    });
    if (search) params.append('search', search);

    return request<PaginatedMounts>(`/admin/mounts?${params}`, {
        method: 'GET'
    });
}

export async function adminCreateMount(name: string, description: string, source: string, target: string, readOnly: boolean, userMountable: boolean, navigable: boolean, nodes: string[], packages: string[]): Promise<ParsedResponse<Mount>> {
    return request<Mount>('/admin/mounts', {
        method: 'POST',
        body: JSON.stringify({ name, description, source, target, read_only: readOnly, user_mountable: userMountable, navigable: navigable, nodes: nodes, packages: packages })
    });
}

export async function adminUpdateMount(id: string, updates: Partial<{ name: string; description: string; source: string; target: string; read_only: boolean; user_mountable: boolean; navigable: boolean; nodes: string[]; packages: string[] }>): Promise<ParsedResponse<Mount>> {
    return request<Mount>(`/admin/mounts/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(updates)
    });
}

export async function adminDeleteMount(id: string): Promise<ParsedResponse<void>> {
    return request<void>(`/admin/mounts/${id}`, {
        method: 'DELETE'
    });
}

export async function adminAttachMount(id: string, serverId: string): Promise<ParsedResponse<void>> {
    return request<void>(`/admin/mounts/${id}/servers/${serverId}`, {
        method: 'POST'
    });
}

export async function adminDetachMount(id: string, serverId: string): Promise<ParsedResponse<void>> {
    return request<void>(`/admin/mounts/${id}/servers/${serverId}`, {
        method: 'DELETE'
    });
}
