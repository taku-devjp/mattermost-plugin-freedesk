import React, {useCallback} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import {executeConfirm, hideConfirm, loadMatrix, showConfirm} from '../actions';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';
import type {MatrixData, MatrixReservation} from '../types';

import ConfirmDialog from './confirm_dialog';

interface Props {
    matrix: MatrixData;
    isPluginAdmin: boolean;
}

function formatDateLabel(date: string): string {
    const parts = date.split('-');
    if (parts.length !== 3) {
        return date;
    }
    return `${parts[1]}/${parts[2]}`;
}

function findReservation(reservations: MatrixReservation[], deskId: string, date: string): MatrixReservation | undefined {
    return reservations.find((r) => r.desk_id === deskId && r.reserve_date === date);
}

const MatrixTable: React.FC<Props> = ({matrix, isPluginAdmin}) => {
    const dispatch = useDispatch();
    const confirmDialog = useSelector((state: GlobalState) => getPluginStateFromGlobal(state).confirmDialog);

    const handleCellClick = useCallback((deskId: string, deskName: string, date: string) => {
        if (date < matrix.today) {
            return;
        }

        const reservation = findReservation(matrix.reservations, deskId, date);
        if (!reservation) {
            dispatch(showConfirm({
                action: 'book',
                deskId,
                deskName,
                date,
            }));
            return;
        }

        if (reservation.is_mine) {
            dispatch(showConfirm({
                action: 'cancel',
                deskId,
                deskName,
                date,
                reservationId: reservation.id,
            }));
            return;
        }

        if (isPluginAdmin) {
            dispatch(showConfirm({
                action: 'proxy_cancel',
                deskId,
                deskName,
                date,
                reservationId: reservation.id,
                userName: reservation.user_name,
            }));
        }
    }, [dispatch, isPluginAdmin, matrix.reservations, matrix.today]);

    const handleNextMonth = useCallback(() => {
        let nextYear = matrix.year;
        let nextMonth = matrix.month + 1;
        if (nextMonth > 12) {
            nextMonth = 1;
            nextYear += 1;
        }
        loadMatrix(dispatch, nextYear, nextMonth);
    }, [dispatch, matrix.month, matrix.year]);

    return (
        <div className='freedesk-matrix-container'>
            <div className='freedesk-matrix-toolbar'>
                <span className='freedesk-matrix-month'>{matrix.year}年{matrix.month}月</span>
                {matrix.can_go_next && (
                    <button
                        type='button'
                        className='btn btn-tertiary btn-sm'
                        onClick={handleNextMonth}
                    >
                        翌月 →
                    </button>
                )}
            </div>
            <div className='freedesk-matrix-scroll'>
                <table className='freedesk-matrix'>
                    <thead>
                        <tr>
                            <th className='freedesk-matrix-date-header'>日付</th>
                            {matrix.desks.map((desk) => (
                                <th
                                    key={desk.id}
                                    className='freedesk-matrix-desk-header'
                                    title={desk.name}
                                >
                                    {desk.name}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {matrix.dates.map((date) => (
                            <tr key={date}>
                                <td className='freedesk-matrix-date'>{formatDateLabel(date)}</td>
                                {matrix.desks.map((desk) => {
                                    const reservation = findReservation(matrix.reservations, desk.id, date);
                                    const isPast = date < matrix.today;
                                    const isBookable = !isPast && !reservation;
                                    const isMine = Boolean(reservation?.is_mine);
                                    const isOthers = Boolean(reservation && !reservation.is_mine);
                                    const isClickable = isBookable || isMine || (isOthers && isPluginAdmin);

                                    let cellClass = 'freedesk-matrix-cell';
                                    if (isPast) {
                                        cellClass += ' freedesk-matrix-cell--past';
                                    } else if (isMine) {
                                        cellClass += ' freedesk-matrix-cell--mine';
                                    } else if (isOthers) {
                                        cellClass += ' freedesk-matrix-cell--occupied';
                                    } else {
                                        cellClass += ' freedesk-matrix-cell--empty';
                                    }
                                    if (isClickable) {
                                        cellClass += ' freedesk-matrix-cell--clickable';
                                    }

                                    return (
                                        <td
                                            key={`${date}-${desk.id}`}
                                            className={cellClass}
                                            onClick={isClickable ? () => handleCellClick(desk.id, desk.name, date) : undefined}
                                            title={reservation?.user_name || (isBookable ? '空き' : '')}
                                        >
                                            {reservation ? reservation.user_name : (isPast ? '' : '—')}
                                        </td>
                                    );
                                })}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            {confirmDialog && (
                <ConfirmDialog
                    dialog={confirmDialog}
                    onConfirm={() => executeConfirm(dispatch, confirmDialog, matrix.year, matrix.month)}
                    onCancel={() => dispatch(hideConfirm())}
                />
            )}
        </div>
    );
};

export default MatrixTable;
