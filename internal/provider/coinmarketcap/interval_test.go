package coinmarketcap

import (
	"testing"
	"time"
)

func TestIntervalForRange(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name string
		from time.Time
		to   time.Time
		want string
	}{
		{
			name: "less than 2 days",
			from: now.Add(-24 * time.Hour),
			to:   now,
			want: "5m",
		},
		{
			name: "3 days",
			from: now.Add(-3 * 24 * time.Hour),
			to:   now,
			want: "1h",
		},
		{
			name: "4 month",
			from: now.Add(-120 * 24 * time.Hour),
			to:   now,
			want: "1d",
		},
		{
			name: "exactly 2 days",
			from: now.Add(-2 * 24 * time.Hour),
			to:   now,
			want: "5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := intervalForRange(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("IntervalForRange() = %q, want %q", got, tt.want)
			}
		})
	}
}
