import React from 'react';
import {useSelector} from 'react-redux';

import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';

import MatrixTable from './matrix_table';

const ReservationTab: React.FC = () => {
    const {matrix, config, loading, error} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));

    if (loading && !matrix) {
        return <div className='freedesk-loading'>読み込み中...</div>;
    }

    if (error) {
        return <div className='freedesk-error'>{error}</div>;
    }

    if (!matrix || !config) {
        return <div className='freedesk-error'>データがありません。</div>;
    }

    return (
        <div className='freedesk-reservation-tab'>
            {error && <div className='freedesk-error'>{error}</div>}
            <MatrixTable
                matrix={matrix}
                isPluginAdmin={config.is_plugin_admin}
            />
            {loading && <div className='freedesk-loading-overlay'>更新中...</div>}
        </div>
    );
};

export default ReservationTab;
