import React from 'react';
import {useSelector} from 'react-redux';

import MatrixTable from './matrix_table';

import {labels} from '../labels';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';

const ReservationTab: React.FC = () => {
    const {matrix, config, loading} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));

    if (loading && !matrix) {
        return <div className='freedesk-loading'>{labels.loading}</div>;
    }

    if (!matrix || !config) {
        return null;
    }

    return (
        <div className='freedesk-reservation-tab'>
            <MatrixTable
                matrix={matrix}
                isPluginAdmin={config.is_plugin_admin}
            />
            {loading && <div className='freedesk-loading-overlay'>{labels.updating}</div>}
        </div>
    );
};

export default ReservationTab;
