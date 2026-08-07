import {isOffDay} from './date';
import {isJapaneseHoliday} from './japanese_holidays';

describe('date utils', () => {
    test('identifies saturday as off day', () => {
        expect(isOffDay('2026-08-08')).toBe(true);
    });

    test('identifies sunday as off day', () => {
        expect(isOffDay('2026-08-09')).toBe(true);
    });

    test('identifies weekday as not off day', () => {
        expect(isOffDay('2026-08-07')).toBe(false);
    });

    test('identifies holiday as off day', () => {
        expect(isJapaneseHoliday('2026-01-01')).toBe(true);
        expect(isOffDay('2026-01-01')).toBe(true);
    });

    test('identifies coming-of-age day', () => {
        expect(isJapaneseHoliday('2026-01-12')).toBe(true);
        expect(isOffDay('2026-01-12')).toBe(true);
    });
});
