import type {FreeDeskState} from './types';

export const ActionTypes = {
    OPEN_MODAL: 'OPEN_MODAL',
    CLOSE_MODAL: 'CLOSE_MODAL',
    SET_TAB: 'SET_TAB',
    SET_LOADING: 'SET_LOADING',
    SET_ERROR: 'SET_ERROR',
    SET_CONFIG: 'SET_CONFIG',
    SET_MATRIX: 'SET_MATRIX',
    SET_DESKS: 'SET_DESKS',
    SET_LOCATIONS: 'SET_LOCATIONS',
    SET_MONTH: 'SET_MONTH',
    SHOW_CONFIRM: 'SHOW_CONFIRM',
    HIDE_CONFIRM: 'HIDE_CONFIRM',
    SET_ADMIN_SAVING: 'SET_ADMIN_SAVING',
} as const;

export const initialState: FreeDeskState = {
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

type Action =
    | {type: typeof ActionTypes.OPEN_MODAL}
    | {type: typeof ActionTypes.CLOSE_MODAL}
    | {type: typeof ActionTypes.SET_TAB; tab: 'reservation' | 'admin'}
    | {type: typeof ActionTypes.SET_LOADING; loading: boolean}
    | {type: typeof ActionTypes.SET_ERROR; error: string | null}
    | {type: typeof ActionTypes.SET_CONFIG; config: FreeDeskState['config']}
    | {type: typeof ActionTypes.SET_MATRIX; matrix: FreeDeskState['matrix']}
    | {type: typeof ActionTypes.SET_DESKS; desks: FreeDeskState['desks']}
    | {type: typeof ActionTypes.SET_LOCATIONS; locations: FreeDeskState['locations']}
    | {type: typeof ActionTypes.SET_MONTH; year: number; month: number}
    | {type: typeof ActionTypes.SHOW_CONFIRM; dialog: FreeDeskState['confirmDialog']}
    | {type: typeof ActionTypes.HIDE_CONFIRM}
    | {type: typeof ActionTypes.SET_ADMIN_SAVING; saving: boolean};

export default function reducer(state: FreeDeskState = initialState, action: Action): FreeDeskState {
    switch (action.type) {
    case ActionTypes.OPEN_MODAL:
        return {...state, modalOpen: true, error: null};
    case ActionTypes.CLOSE_MODAL:
        return {...state, modalOpen: false, confirmDialog: null, error: null};
    case ActionTypes.SET_TAB:
        return {...state, activeTab: action.tab};
    case ActionTypes.SET_LOADING:
        return {...state, loading: action.loading};
    case ActionTypes.SET_ERROR:
        return {...state, error: action.error};
    case ActionTypes.SET_CONFIG:
        return {...state, config: action.config};
    case ActionTypes.SET_MATRIX:
        return {
            ...state,
            matrix: action.matrix,
            year: action.matrix?.year ?? state.year,
            month: action.matrix?.month ?? state.month,
        };
    case ActionTypes.SET_DESKS:
        return {...state, desks: action.desks};
    case ActionTypes.SET_LOCATIONS:
        return {...state, locations: action.locations};
    case ActionTypes.SET_MONTH:
        return {...state, year: action.year, month: action.month};
    case ActionTypes.SHOW_CONFIRM:
        return {...state, confirmDialog: action.dialog};
    case ActionTypes.HIDE_CONFIRM:
        return {...state, confirmDialog: null};
    case ActionTypes.SET_ADMIN_SAVING:
        return {...state, adminSaving: action.saving};
    default:
        return state;
    }
}
