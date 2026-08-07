import React, {useCallback, useLayoutEffect, useRef, useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import ConfirmDialog from './confirm_dialog';

import {executeConfirm, hideConfirm, loadMatrix, showConfirm} from '../actions';
import {labels} from '../labels';
import type {MatrixData, MatrixReservation} from '../types';
import type {GlobalState} from '../types/mattermost';
import {getPluginStateFromGlobal} from '../types/mattermost';
import {formatDateLabel} from '../utils/date';

interface Props {
    matrix: MatrixData;
    isPluginAdmin: boolean;
}

function findReservation(reservations: MatrixReservation[], deskId: string, date: string): MatrixReservation | undefined {
    return reservations.find((r) => r.desk_id === deskId && r.reserve_date === date);
}

function getCellClassName(isPast: boolean, isMine: boolean, isOthers: boolean, isClickable: boolean): string {
    let cellClass = 'freedesk-matrix-cell';
    if (isMine) {
        cellClass += ' freedesk-matrix-cell--mine';
    } else if (isOthers) {
        cellClass += ' freedesk-matrix-cell--occupied';
    } else {
        cellClass += ' freedesk-matrix-cell--empty';
    }
    if (isPast) {
        cellClass += ' freedesk-matrix-cell--past';
    }
    if (isClickable) {
        cellClass += ' freedesk-matrix-cell--clickable';
    }
    return cellClass;
}

function getCellContent(reservation: MatrixReservation | undefined): string {
    if (reservation) {
        return reservation.user_name;
    }
    return labels.emptySlot;
}

const MatrixTable: React.FC<Props> = ({matrix, isPluginAdmin}) => {
    const dispatch = useDispatch();
    const confirmDialog = useSelector((state: GlobalState) => getPluginStateFromGlobal(state).confirmDialog);
    const tableWrapRef = useRef<HTMLDivElement>(null);
    const headerRef = useRef<HTMLDivElement>(null);
    const scrollRef = useRef<HTMLDivElement>(null);
    const [hoveredDeskId, setHoveredDeskId] = useState<string | null>(null);

    useLayoutEffect(() => {
        const scrollEl = scrollRef.current;
        const wrapEl = tableWrapRef.current;
        if (!scrollEl || !wrapEl) {
            return undefined;
        }

        const updateScrollbarGutter = () => {
            const scrollbarWidth = scrollEl.offsetWidth - scrollEl.clientWidth;
            wrapEl.style.setProperty('--freedesk-scrollbar-width', `${scrollbarWidth}px`);
        };

        updateScrollbarGutter();

        const observer = new ResizeObserver(updateScrollbarGutter);
        observer.observe(scrollEl);

        return () => observer.disconnect();
    }, [matrix.dates.length, matrix.desks.length]);

    const handleBodyScroll = useCallback(() => {
        const scrollEl = scrollRef.current;
        const headerEl = headerRef.current;
        if (scrollEl && headerEl) {
            headerEl.scrollLeft = scrollEl.scrollLeft;
        }
    }, []);

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

    const handlePrevMonth = useCallback(() => {
        let prevYear = matrix.year;
        let prevMonth = matrix.month - 1;
        if (prevMonth < 1) {
            prevMonth = 12;
            prevYear -= 1;
        }
        loadMatrix(dispatch, prevYear, prevMonth);
    }, [dispatch, matrix.month, matrix.year]);

    const handleNextMonth = useCallback(() => {
        let nextYear = matrix.year;
        let nextMonth = matrix.month + 1;
        if (nextMonth > 12) {
            nextMonth = 1;
            nextYear += 1;
        }
        loadMatrix(dispatch, nextYear, nextMonth);
    }, [dispatch, matrix.month, matrix.year]);

    const renderColgroup = () => (
        <colgroup>
            <col className='freedesk-matrix-date-col'/>
            {matrix.desks.map((desk) => (
                <col
                    key={desk.id}
                    className='freedesk-matrix-desk-col'
                />
            ))}
        </colgroup>
    );

    const renderHeaderRow = () => (
        <tr>
            <th className='freedesk-matrix-date-header'>{labels.dateHeader}</th>
            {matrix.desks.map((desk) => (
                <th
                    key={desk.id}
                    className={`freedesk-matrix-desk-header${hoveredDeskId === desk.id ? ' freedesk-matrix-desk-header--highlighted' : ''}`}
                    title={desk.name}
                >
                    {desk.name}
                </th>
            ))}
        </tr>
    );

    return (
        <div className='freedesk-matrix-container'>
            <div className='freedesk-matrix-toolbar'>
                <div className='freedesk-matrix-toolbar-action freedesk-matrix-toolbar-action--prev'>
                    {matrix.can_go_prev && (
                        <button
                            type='button'
                            className='btn btn-tertiary btn-sm'
                            onClick={handlePrevMonth}
                        >
                            {labels.prevMonth}
                        </button>
                    )}
                </div>
                <span className='freedesk-matrix-month'>
                    {matrix.year}
                    {labels.yearSuffix}
                    {matrix.month}
                    {labels.monthSuffix}
                </span>
                <div className='freedesk-matrix-toolbar-action'>
                    {matrix.can_go_next && (
                        <button
                            type='button'
                            className='btn btn-tertiary btn-sm'
                            onClick={handleNextMonth}
                        >
                            {labels.nextMonth}
                        </button>
                    )}
                </div>
            </div>
            <div
                className='freedesk-matrix-table-wrap'
                ref={tableWrapRef}
            >
                <div
                    className='freedesk-matrix-header'
                    ref={headerRef}
                >
                    <table className='freedesk-matrix freedesk-matrix--header'>
                        {renderColgroup()}
                        <thead>
                            {renderHeaderRow()}
                        </thead>
                    </table>
                </div>
                <div
                    className='freedesk-matrix-scroll'
                    ref={scrollRef}
                    onScroll={handleBodyScroll}
                >
                    <table className='freedesk-matrix freedesk-matrix--body'>
                        {renderColgroup()}
                        <tbody>
                            {matrix.dates.map((date) => {
                                const isPast = date < matrix.today;
                                const isToday = date === matrix.today;
                                const rowClass = isPast ?
                                    'freedesk-matrix-row--past' :
                                    (isToday ? 'freedesk-matrix-row--today' : '');
                                return (
                                <tr
                                    key={date}
                                    className={rowClass || undefined}
                                >
                                    <td className='freedesk-matrix-date'>{formatDateLabel(date)}</td>
                                    {matrix.desks.map((desk) => {
                                        const reservation = findReservation(matrix.reservations, desk.id, date);
                                        const isBookable = !isPast && !reservation;
                                        const isMine = Boolean(reservation?.is_mine);
                                        const isOthers = Boolean(reservation && !reservation.is_mine);
                                        const isClickable = isBookable || isMine || (isOthers && isPluginAdmin);
                                        const cellClass = getCellClassName(isPast, isMine, isOthers, isClickable);
                                        const cellContent = getCellContent(reservation);

                                        return (
                                            <td
                                                key={`${date}-${desk.id}`}
                                                className={cellClass}
                                                onClick={isClickable ? () => handleCellClick(desk.id, desk.name, date) : undefined}
                                                onMouseEnter={isClickable ? () => setHoveredDeskId(desk.id) : undefined}
                                                onMouseLeave={isClickable ? () => setHoveredDeskId(null) : undefined}
                                                title={reservation?.user_name || (isBookable ? labels.available : '')}
                                            >
                                                {cellContent}
                                            </td>
                                        );
                                    })}
                                </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
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
