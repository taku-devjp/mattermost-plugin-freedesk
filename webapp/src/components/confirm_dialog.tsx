import React from 'react';

import {labels} from '../labels';
import type {ConfirmDialogState} from '../types';

interface Props {
    dialog: ConfirmDialogState;
    onConfirm: () => void;
    onCancel: () => void;
}

function getMessage(dialog: ConfirmDialogState): string {
    switch (dialog.action) {
    case 'book':
        return '予約しますか？';
    case 'cancel':
        return '予約を取り消しますか？';
    case 'proxy_cancel':
        return '代理で予約を取り消しますか？';
    case 'delete_desk':
        return 'このデスクを削除しますか？';
    default:
        return '実行しますか？';
    }
}

const ConfirmDialog: React.FC<Props> = ({dialog, onConfirm, onCancel}) => {
    return (
        <div
            className='freedesk-confirm-overlay'
            onClick={onCancel}
            role='presentation'
        >
            <div
                className='freedesk-confirm-dialog'
                onClick={(e) => e.stopPropagation()}
                role='dialog'
                aria-modal='true'
            >
                <h3 className='freedesk-confirm-title'>{labels.confirm}</h3>
                <div className='freedesk-confirm-body'>
                    <p><strong>{labels.deskLabel}</strong> {dialog.deskName}</p>
                    {dialog.date && <p><strong>{labels.dateLabel}</strong> {dialog.date}</p>}
                    {dialog.userName && <p><strong>{labels.reserverLabel}</strong> {dialog.userName}</p>}
                    <p>{getMessage(dialog)}</p>
                </div>
                <div className='freedesk-confirm-actions'>
                    <button
                        type='button'
                        className='btn btn-tertiary'
                        onClick={onCancel}
                    >
                        {labels.cancel}
                    </button>
                    <button
                        type='button'
                        className='btn btn-primary'
                        onClick={onConfirm}
                    >
                        {labels.ok}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ConfirmDialog;
