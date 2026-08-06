const WEEKDAY_LABELS = ['日', '月', '火', '水', '木', '金', '土'];

function getWeekdayLabel(date: string): string | null {
    const parts = date.split('-');
    if (parts.length !== 3) {
        return null;
    }
    const year = parseInt(parts[0], 10);
    const month = parseInt(parts[1], 10) - 1;
    const day = parseInt(parts[2], 10);
    const d = new Date(year, month, day);
    return WEEKDAY_LABELS[d.getDay()];
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
