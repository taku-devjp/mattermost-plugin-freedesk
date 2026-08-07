import React, {useEffect} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import AdminTab from './admin_tab';
import ErrorDialog from './error_dialog';
import ReservationTab from './reservation_tab';

import {closeModal, hideError, loadModalData, setTab} from '../actions';
import {labels} from '../labels';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';

const FreeDeskModal: React.FC = () => {
    const dispatch = useDispatch();
    const {modalOpen, activeTab, config, loading, errorDialog} = useSelector((state: GlobalState) => getPluginStateFromGlobal(state));

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
                aria-label={labels.modalTitle}
            >
                <div className='freedesk-modal-header'>
                    <div className='freedesk-modal-header-spacer'/>
                    <h2 className='freedesk-modal-title'>{labels.modalTitle}</h2>
                    <button
                        type='button'
                        className='freedesk-modal-close'
                        onClick={() => dispatch(closeModal())}
                        aria-label={labels.close}
                    >
                        {labels.closeSymbol}
                    </button>
                </div>
                <div className='freedesk-modal-tabs'>
                    <button
                        type='button'
                        className={`freedesk-tab ${activeTab === 'reservation' ? 'freedesk-tab--active' : ''}`}
                        onClick={() => dispatch(setTab('reservation'))}
                    >
                        {labels.tabReservation}
                    </button>
                    {showAdminTab && (
                        <button
                            type='button'
                            className={`freedesk-tab ${activeTab === 'admin' ? 'freedesk-tab--active' : ''}`}
                            onClick={() => dispatch(setTab('admin'))}
                        >
                            {labels.tabAdmin}
                        </button>
                    )}
                </div>
                <div className='freedesk-modal-body'>
                    {loading && !config ? (
                        <div className='freedesk-loading'>{labels.loading}</div>
                    ) : (
                        <>
                            {activeTab === 'reservation' && <ReservationTab/>}
                            {activeTab === 'admin' && showAdminTab && <AdminTab/>}
                        </>
                    )}
                </div>
            </div>
            {errorDialog && (
                <ErrorDialog
                    message={errorDialog}
                    onClose={() => dispatch(hideError())}
                />
            )}
        </div>
    );
};

export default FreeDeskModal;
