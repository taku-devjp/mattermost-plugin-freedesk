import type {GlobalState as MMGlobalState} from '@mattermost/types/store';

import {PLUGIN_STATE_KEY} from '../actions';
import type {FreeDeskState} from '../types';

export interface GlobalState extends MMGlobalState {
    [PLUGIN_STATE_KEY]?: FreeDeskState;
}

export function getPluginStateFromGlobal(state: GlobalState): FreeDeskState {
    return state[PLUGIN_STATE_KEY] ?? {
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
