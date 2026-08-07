import {parseLocalDate} from './date';

const holidayCache = new Map<number, Set<string>>();

function formatDateParts(year: number, month: number, day: number): string {
    return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
}

function nthWeekdayOfMonth(year: number, month: number, weekday: number, n: number): number {
    const daysInMonth = new Date(year, month - 1, 0).getDate();
    let count = 0;
    for (let day = 1; day <= daysInMonth; day++) {
        if (new Date(year, month - 1, day).getDay() === weekday) {
            count++;
            if (count === n) {
                return day;
            }
        }
    }
    return -1;
}

function vernalEquinoxDay(year: number): number {
    return Math.floor(20.8431 + 0.242194 * (year - 1980) - Math.floor((year - 1980) / 4));
}

function autumnalEquinoxDay(year: number): number {
    return Math.floor(23.2488 + 0.242194 * (year - 1980) - Math.floor((year - 1980) / 4));
}

function getBaseHolidays(year: number): Set<string> {
    const holidays = new Set<string>();
    const add = (month: number, day: number) => {
        if (day > 0) {
            holidays.add(formatDateParts(year, month, day));
        }
    };

    add(1, 1);

    if (year >= 2000) {
        if (year < 2020) {
            add(1, 15);
        } else {
            add(1, nthWeekdayOfMonth(year, 1, 1, 2));
        }
    }

    add(2, 11);
    if (year >= 2020) {
        add(2, 23);
    }

    add(3, vernalEquinoxDay(year));
    add(4, 29);
    add(5, 3);
    add(5, 4);
    add(5, 5);

    if (year >= 2016) {
        if (year === 2020) {
            add(7, 23);
        } else if (year === 2021) {
            add(7, 22);
        } else if (year >= 2003) {
            add(7, nthWeekdayOfMonth(year, 7, 1, 3));
        } else {
            add(7, 20);
        }
    }

    if (year >= 2016) {
        if (year === 2020) {
            add(8, 10);
        } else if (year === 2021) {
            add(8, 8);
        } else {
            add(8, 11);
        }
    }

    add(9, nthWeekdayOfMonth(year, 9, 1, 3));
    add(9, autumnalEquinoxDay(year));
    add(10, nthWeekdayOfMonth(year, 10, 1, 2));
    add(11, 3);
    add(11, 23);

    return holidays;
}

function applySubstituteAndCitizensHolidays(baseHolidays: Set<string>, year: number): Set<string> {
    const holidays = new Set(baseHolidays);

    for (let month = 1; month <= 12; month++) {
        const daysInMonth = new Date(year, month, 0).getDate();
        for (let day = 2; day < daysInMonth; day++) {
            const date = formatDateParts(year, month, day);
            const weekday = new Date(year, month - 1, day).getDay();
            if (weekday === 0 || weekday === 6 || holidays.has(date)) {
                continue;
            }
            const prev = formatDateParts(year, month, day - 1);
            const next = formatDateParts(year, month, day + 1);
            if (holidays.has(prev) && holidays.has(next)) {
                holidays.add(date);
            }
        }
    }

    for (const date of baseHolidays) {
        const parsed = parseLocalDate(date);
        if (!parsed || parsed.getDay() !== 0) {
            continue;
        }
        const substitute = new Date(parsed);
        while (true) {
            substitute.setDate(substitute.getDate() + 1);
            const substituteDate = formatDateParts(
                substitute.getFullYear(),
                substitute.getMonth() + 1,
                substitute.getDate(),
            );
            if (!holidays.has(substituteDate)) {
                holidays.add(substituteDate);
                break;
            }
        }
    }

    return holidays;
}

function getHolidaysForYear(year: number): Set<string> {
    if (!holidayCache.has(year)) {
        const base = getBaseHolidays(year);
        holidayCache.set(year, applySubstituteAndCitizensHolidays(base, year));
    }
    return holidayCache.get(year)!;
}

export function isJapaneseHoliday(date: string): boolean {
    const parsed = parseLocalDate(date);
    if (!parsed) {
        return false;
    }
    return getHolidaysForYear(parsed.getFullYear()).has(date);
}
