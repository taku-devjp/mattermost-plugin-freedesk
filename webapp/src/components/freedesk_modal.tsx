import React, {useEffect} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import {closeModal, loadModalData, setTab} from '../actions';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';

import AdminTab from './admin_tab';
import ReservationTab from './reservation_tab';

const FreeDeskModal: React.FC = () => {
    const dispatch = useDispatch();
    const {modalOpen, activeTab, config, loading} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));

    useEffect(() => {
        if (modalOpen) {
            loadModalData(dispatch);
        }
    }, [dispatch, modalOpen]);

    if (!modalOpen) {
        return null;
    }

    const showAdminTab = config?.is_plugin_admin;

    return (
        <div
            className='freedesk-modal-overlay'
            onClick={() => dispatch(closeModal())}
            role='presentation'
        >
            <div
                className='freedesk-modal'
                onClick={(e) => e.stopPropagation()}
                role='dialog'
                aria-modal='true'
                aria-label='フリーデスク予約'
            >
                <div className='freedesk-modal-header'>
                    <h2>フリーデスク予約</h2>
                    <button
                        type='button'
                        className='freedesk-modal-close'
                        onClick={() => dispatch(closeModal())}
                        aria-label='閉じる'
                    >
                        ×
                    </button>
                </div>
                <div className='freedesk-modal-tabs'>
                    <button
                        type='button'
                        className={`freedesk-tab ${activeTab === 'reservation' ? 'freedesk-tab--active' : ''}`}
                        onClick={() => dispatch(setTab('reservation'))}
                    >
                        予約
                    </button>
                    {showAdminTab && (
                        <button
                            type='button'
                            className={`freedesk-tab ${activeTab === 'admin' ? 'freedesk-tab--active' : ''}`}
                            onClick={() => dispatch(setTab('admin'))}
                        >
                            管理
                        </button>
                    )}
                </div>
                <div className='freedesk-modal-body'>
                    {loading && !config ? (
                        <div className='freedesk-loading'>読み込み中...</div>
                    ) : (
                        <>
                            {activeTab === 'reservation' && <ReservationTab/>}
                            {activeTab === 'admin' && showAdminTab && <AdminTab/>}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
};

export default FreeDeskModal;
