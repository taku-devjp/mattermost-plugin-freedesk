package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeMaxAdvanceMonths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
		want  int
	}{
		{name: "default for zero", input: 0, want: DefaultMaxMonths},
		{name: "default for negative", input: -1, want: DefaultMaxMonths},
		{name: "minimum", input: 1, want: 1},
		{name: "maximum", input: 6, want: 6},
		{name: "clamp below minimum", input: 0, want: DefaultMaxMonths},
		{name: "clamp above maximum", input: 12, want: MaxMaxMonths},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, NormalizeMaxAdvanceMonths(tt.input))
		})
	}
}

func TestBookableUntilFrom(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 7, 12, 0, 0, 0, location)

	tests := []struct {
		name   string
		months int
		want   string
	}{
		{name: "two months", months: 2, want: "2026-10-31"},
		{name: "one month", months: 1, want: "2026-09-30"},
		{name: "six months", months: 6, want: "2027-02-28"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, bookableUntilFrom(now, tt.months))
		})
	}
}
