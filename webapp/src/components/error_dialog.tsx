import React from 'react';

import {labels} from '../labels';

interface Props {
    message: string;
    onClose: () => void;
}

const ErrorDialog: React.FC<Props> = ({message, onClose}) => {
    return (
        <div
            className='freedesk-error-overlay'
            onClick={onClose}
            role='presentation'
        >
            <div
                className='freedesk-error-dialog'
                onClick={(e) => e.stopPropagation()}
                role='alertdialog'
                aria-modal='true'
            >
                <h3 className='freedesk-error-title'>{labels.errorTitle}</h3>
                <div className='freedesk-error-body'>
                    <p>{message}</p>
                </div>
                <div className='freedesk-error-actions'>
                    <button
                        type='button'
                        className='btn btn-primary'
                        onClick={onClose}
                    >
                        {labels.ok}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ErrorDialog;
