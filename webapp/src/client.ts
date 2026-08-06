import manifest from 'manifest';

import type {APIErrorBody, ConfigData, Desk, Location, MatrixData, Reservation} from './types';

const API_BASE = `/plugins/${manifest.id}/api/v1`;

export class APIError extends Error {
    code: string;
    details?: Record<string, unknown>;

    constructor(body: APIErrorBody['error']) {
        super(body.message);
        this.code = body.code;
        this.details = body.details;
    }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await fetch(`${API_BASE}${path}`, {
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {}),
        },
        ...options,
    });

    if (response.status === 204) {
        return undefined as T;
    }

    const body = await response.json();
    if (!response.ok) {
        throw new APIError(body.error);
    }
    return body.data as T;
}

export function getConfig(): Promise<ConfigData> {
    return request<ConfigData>('/config');
}

export function getMatrix(year?: number, month?: number, locationId?: string): Promise<MatrixData> {
    const params = new URLSearchParams();
    if (year) {
        params.set('year', String(year));
    }
    if (month) {
        params.set('month', String(month));
    }
    if (locationId) {
        params.set('location_id', locationId);
    }
    const qs = params.toString();
    return request<MatrixData>(`/matrix${qs ? `?${qs}` : ''}`);
}

export function getDesks(includeInactive = false, locationId?: string): Promise<{desks: Desk[]}> {
    const params = new URLSearchParams();
    if (includeInactive) {
        params.set('include_inactive', 'true');
    }
    if (locationId) {
        params.set('location_id', locationId);
    }
    const qs = params.toString();
    return request<{desks: Desk[]}>(`/desks${qs ? `?${qs}` : ''}`);
}

export function getLocations(): Promise<{locations: Location[]}> {
    return request<{locations: Location[]}>('/locations');
}

export function createReservation(deskId: string, reserveDate: string): Promise<Reservation> {
    return request<Reservation>('/reservations', {
        method: 'POST',
        body: JSON.stringify({desk_id: deskId, reserve_date: reserveDate}),
    });
}

export function deleteReservation(id: string): Promise<void> {
    return request<void>(`/reservations/${id}`, {method: 'DELETE'});
}

export function createDesk(payload: {
    location_id: string;
    name: string;
    sort_order?: number;
    is_active?: boolean;
}): Promise<Desk> {
    return request<Desk>('/admin/desks', {
        method: 'POST',
        body: JSON.stringify(payload),
    });
}

export function updateDesk(id: string, payload: {
    name?: string;
    sort_order?: number;
    is_active?: boolean;
}): Promise<Desk> {
    return request<Desk>(`/admin/desks/${id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
    });
}

export function deleteDesk(id: string): Promise<void> {
    return request<void>(`/admin/desks/${id}`, {method: 'DELETE'});
}
