import type {Dispatch} from 'redux';

import * as client from './client';
import {ActionTypes} from './reducer';
import type {ConfirmDialogState, FreeDeskState} from './types';

export const PLUGIN_STATE_KEY = 'plugins-com.freedesk.mattermost';

export function getPluginState(state: { [key: string]: unknown }): FreeDeskState {
    const pluginState = state[PLUGIN_STATE_KEY] as FreeDeskState | undefined;
    return pluginState ?? {
        modalOpen: false,
        activeTab: 'reservation',
        config: null,
        matrix: null,
        desks: [],
        locations: [],
        loading: false,
        error: null,
        year: 0,
        month: 0,
        confirmDialog: null,
        adminSaving: false,
    };
}

export function openModal() {
    return {type: ActionTypes.OPEN_MODAL};
}

export function closeModal() {
    return {type: ActionTypes.CLOSE_MODAL};
}

export function setTab(tab: 'reservation' | 'admin') {
    return {type: ActionTypes.SET_TAB, tab};
}

export function showConfirm(dialog: ConfirmDialogState) {
    return {type: ActionTypes.SHOW_CONFIRM, dialog};
}

export function hideConfirm() {
    return {type: ActionTypes.HIDE_CONFIRM};
}

export async function loadModalData(dispatch: Dispatch) {
    dispatch({type: ActionTypes.SET_LOADING, loading: true});
    dispatch({type: ActionTypes.SET_ERROR, error: null});
    try {
        const config = await client.getConfig();
        dispatch({type: ActionTypes.SET_CONFIG, config});

        const matrix = await client.getMatrix();
        dispatch({type: ActionTypes.SET_MATRIX, matrix});

        if (config.is_plugin_admin) {
            const [{desks}, {locations}] = await Promise.all([
                client.getDesks(true),
                client.getLocations(),
            ]);
            dispatch({type: ActionTypes.SET_DESKS, desks});
            dispatch({type: ActionTypes.SET_LOCATIONS, locations});
        }
    } catch (err) {
        const message = err instanceof client.APIError ? err.message : 'データの取得に失敗しました。';
        dispatch({type: ActionTypes.SET_ERROR, error: message});
    } finally {
        dispatch({type: ActionTypes.SET_LOADING, loading: false});
    }
}

export async function loadMatrix(dispatch: Dispatch, year: number, month: number) {
    dispatch({type: ActionTypes.SET_LOADING, loading: true});
    dispatch({type: ActionTypes.SET_ERROR, error: null});
    try {
        const matrix = await client.getMatrix(year, month);
        dispatch({type: ActionTypes.SET_MATRIX, matrix});
    } catch (err) {
        const message = err instanceof client.APIError ? err.message : 'マトリクスの取得に失敗しました。';
        dispatch({type: ActionTypes.SET_ERROR, error: message});
    } finally {
        dispatch({type: ActionTypes.SET_LOADING, loading: false});
    }
}

export async function reloadAdminDesks(dispatch: Dispatch) {
    try {
        const {desks} = await client.getDesks(true);
        dispatch({type: ActionTypes.SET_DESKS, desks});
    } catch (err) {
        const message = err instanceof client.APIError ? err.message : 'デスク一覧の取得に失敗しました。';
        dispatch({type: ActionTypes.SET_ERROR, error: message});
    }
}

export async function executeConfirm(dispatch: Dispatch, dialog: ConfirmDialogState, year: number, month: number) {
    dispatch({type: ActionTypes.SET_LOADING, loading: true});
    dispatch({type: ActionTypes.HIDE_CONFIRM});
    dispatch({type: ActionTypes.SET_ERROR, error: null});
    try {
        switch (dialog.action) {
        case 'book':
            if (dialog.deskId && dialog.date) {
                await client.createReservation(dialog.deskId, dialog.date);
            }
            break;
        case 'cancel':
        case 'proxy_cancel':
            if (dialog.reservationId) {
                await client.deleteReservation(dialog.reservationId);
            }
            break;
        case 'delete_desk':
            if (dialog.deskId) {
                await client.deleteDesk(dialog.deskId);
                await reloadAdminDesks(dispatch);
            }
            break;
        default:
            break;
        }

        if (dialog.action === 'book' || dialog.action === 'cancel' || dialog.action === 'proxy_cancel') {
            const reloadYear = year || new Date().getFullYear();
            const reloadMonth = month || (new Date().getMonth() + 1);
            await loadMatrix(dispatch, reloadYear, reloadMonth);
        }
    } catch (err) {
        const message = err instanceof client.APIError ? err.message : '操作に失敗しました。';
        dispatch({type: ActionTypes.SET_ERROR, error: message});
    } finally {
        dispatch({type: ActionTypes.SET_LOADING, loading: false});
    }
}

export async function saveDesk(
    dispatch: Dispatch,
    deskId: string | null,
    payload: {location_id: string; name: string; sort_order: number; is_active: boolean},
) {
    dispatch({type: ActionTypes.SET_ADMIN_SAVING, saving: true});
    dispatch({type: ActionTypes.SET_ERROR, error: null});
    try {
        if (deskId) {
            await client.updateDesk(deskId, {
                name: payload.name,
                sort_order: payload.sort_order,
                is_active: payload.is_active,
            });
        } else {
            await client.createDesk(payload);
        }
        await reloadAdminDesks(dispatch);
        const matrix = await client.getMatrix();
        dispatch({type: ActionTypes.SET_MATRIX, matrix});
    } catch (err) {
        const message = err instanceof client.APIError ? err.message : 'デスクの保存に失敗しました。';
        dispatch({type: ActionTypes.SET_ERROR, error: message});
    } finally {
        dispatch({type: ActionTypes.SET_ADMIN_SAVING, saving: false});
    }
}
