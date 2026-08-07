import {isJapaneseHoliday} from './japanese_holidays';

const WEEKDAY_LABELS = ['日', '月', '火', '水', '木', '金', '土'];

export function parseLocalDate(date: string): Date | null {
    const parts = date.split('-');
    if (parts.length !== 3) {
        return null;
    }
    const year = parseInt(parts[0], 10);
    const month = parseInt(parts[1], 10) - 1;
    const day = parseInt(parts[2], 10);
    if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day)) {
        return null;
    }
    return new Date(year, month, day);
}

function getWeekdayLabel(date: string): string | null {
    const d = parseLocalDate(date);
    if (!d) {
        return null;
    }
    return WEEKDAY_LABELS[d.getDay()];
}

export function isOffDay(date: string): boolean {
    if (isJapaneseHoliday(date)) {
        return true;
    }
    const d = parseLocalDate(date);
    if (!d) {
        return false;
    }
    const day = d.getDay();
    return day === 0 || day === 6;
}

export function formatDateLabel(date: string): string {
    const parts = date.split('-');
    if (parts.length !== 3) {
        return date;
    }
    const weekday = getWeekdayLabel(date);
    if (!weekday) {
        return `${parts[1]}/${parts[2]}`;
    }
    return `${parts[1]}/${parts[2]}(${weekday})`;
}

export function formatDateFull(date: string): string {
    const parts = date.split('-');
    if (parts.length !== 3) {
        return date;
    }
    const weekday = getWeekdayLabel(date);
    if (!weekday) {
        return `${parts[0]}/${parts[1]}/${parts[2]}`;
    }
    return `${parts[0]}/${parts[1]}/${parts[2]}(${weekday})`;
}
