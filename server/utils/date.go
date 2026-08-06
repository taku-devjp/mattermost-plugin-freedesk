package utils

import (
	"fmt"
	"time"
)

const (
	Timezone       = "Asia/Tokyo"
	DateLayout     = "2006-01-02"
	DefaultMaxDays = 60
)

var location *time.Location

func init() {
	loc, err := time.LoadLocation(Timezone)
	if err != nil {
		panic(fmt.Sprintf("failed to load timezone %s: %v", Timezone, err))
	}
	location = loc
}

// Now returns the current time in Asia/Tokyo.
func Now() time.Time {
	return time.Now().In(location)
}

// Today returns today's date string (YYYY-MM-DD) in Asia/Tokyo.
func Today() string {
	return Now().Format(DateLayout)
}

// BookableUntil returns the last bookable date for the given max advance days.
func BookableUntil(maxAdvanceDays int) string {
	if maxAdvanceDays <= 0 {
		maxAdvanceDays = DefaultMaxDays
	}
	return Now().AddDate(0, 0, maxAdvanceDays).Format(DateLayout)
}

// ParseDate parses a YYYY-MM-DD date string.
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, s, location)
}

// FormatDate formats a time as YYYY-MM-DD in Asia/Tokyo.
func FormatDate(t time.Time) string {
	return t.In(location).Format(DateLayout)
}

// MonthStart returns the first day of the given year/month.
func MonthStart(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, location)
}

// MonthEnd returns the last day of the given year/month.
func MonthEnd(year, month int) time.Time {
	return MonthStart(year, month).AddDate(0, 1, -1)
}

// CompareDates returns -1 if a < b, 0 if equal, 1 if a > b.
func CompareDates(a, b string) int {
	ta, errA := ParseDate(a)
	tb, errB := ParseDate(b)
	if errA != nil || errB != nil {
		return 0
	}
	if ta.Before(tb) {
		return -1
	}
	if ta.After(tb) {
		return 1
	}
	return 0
}

// DateRange generates inclusive date strings from start to end.
func DateRange(start, end string) ([]string, error) {
	startTime, err := ParseDate(start)
	if err != nil {
		return nil, err
	}
	endTime, err := ParseDate(end)
	if err != nil {
		return nil, err
	}
	if endTime.Before(startTime) {
		return []string{}, nil
	}

	var dates []string
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		dates = append(dates, FormatDate(d))
	}
	return dates, nil
}

// IsValidDateFormat checks YYYY-MM-DD format.
func IsValidDateFormat(s string) bool {
	_, err := ParseDate(s)
	return err == nil
}

// CurrentYearMonth returns the current year and month in Asia/Tokyo.
func CurrentYearMonth() (int, int) {
	now := Now()
	return now.Year(), int(now.Month())
}

// IsBeforeMonth returns true if (year, month) is before (refYear, refMonth).
func IsBeforeMonth(year, month, refYear, refMonth int) bool {
	if year < refYear {
		return true
	}
	if year == refYear && month < refMonth {
		return true
	}
	return false
}

// IsAfterMonth returns true if (year, month) is after (refYear, refMonth).
func IsAfterMonth(year, month, refYear, refMonth int) bool {
	if year > refYear {
		return true
	}
	if year == refYear && month > refMonth {
		return true
	}
	return false
}
