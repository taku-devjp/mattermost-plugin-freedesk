// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from 'manifest';
import type {Reducer, Store, UnknownAction} from 'redux';

import type {PluginRegistry} from 'types/mattermost-webapp';

import {openModal} from './actions';
import FreeDeskModal from './components/freedesk_modal';
import reducer from './reducer';
import './styles.scss';

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store) {
        registry.registerReducer({reducer: reducer as Reducer<unknown, UnknownAction>});
        registry.registerRootComponent(FreeDeskModal);

        const iconUrl = `/plugins/${manifest.id}/${manifest.icon_path}`;
        registry.registerAppBarComponent(
            iconUrl,
            () => store.dispatch(openModal()),
            'フリーデスク予約',
            null,
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
