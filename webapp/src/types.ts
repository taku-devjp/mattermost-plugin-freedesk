export interface ConfigData {
    timezone: string;
    today: string;
    max_advance_months: number;
    bookable_until: string;
    one_desk_per_day: boolean;
    notification_enabled: boolean;
    is_plugin_admin: boolean;
}

export interface MatrixDesk {
    id: string;
    name: string;
    sort_order: number;
    is_active: boolean;
}

export interface MatrixReservation {
    id: string;
    desk_id: string;
    user_id: string;
    user_name: string;
    reserve_date: string;
    is_mine: boolean;
}

export interface MatrixData {
    year: number;
    month: number;
    timezone: string;
    today: string;
    bookable_until: string;
    can_go_prev: boolean;
    can_go_next: boolean;
    desks: MatrixDesk[];
    dates: string[];
    reservations: MatrixReservation[];
}

export interface Desk {
    id: string;
    location_id: string;
    name: string;
    sort_order: number;
    is_active: boolean;
    create_at?: number;
}

export interface Location {
    id: string;
    name: string;
    sort_order: number;
}

export interface Reservation {
    id: string;
    desk_id: string;
    desk_name?: string;
    user_id: string;
    user_name?: string;
    reserve_date: string;
    create_at?: number;
}

export type ConfirmAction = 'book' | 'cancel' | 'proxy_cancel' | 'delete_desk';

export interface ConfirmDialogState {
    action: ConfirmAction;
    deskId?: string;
    deskName: string;
    date?: string;
    reservationId?: string;
    userName?: string;
}

export interface FreeDeskState {
    modalOpen: boolean;
    activeTab: 'reservation' | 'admin';
    config: ConfigData | null;
    matrix: MatrixData | null;
    desks: Desk[];
    locations: Location[];
    loading: boolean;
    error: string | null;
    year: number;
    month: number;
    confirmDialog: ConfirmDialogState | null;
    adminSaving: boolean;
}

export interface APIErrorBody {
    error: {
        code: string;
        message: string;
        details?: Record<string, unknown>;
    };
}
