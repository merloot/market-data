package market_test

import (
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

func TestCalculateBucket(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		timeframe market.Timeframe
		from      time.Time
		to        time.Time
		want      market.Bucket
	}{
		{"8h", market.Timeframe8h, now.Add(-8 * time.Hour), now, market.Bucket5m},
		{"day", market.TimeframeDay, now.Add(-24 * time.Hour), now, market.Bucket5m},
		{"week", market.TimeframeWeek, now.Add(-7 * 24 * time.Hour), now, market.Bucket1h},
		{"month", market.TimeframeMonth, now.Add(-30 * 24 * time.Hour), now, market.Bucket1h},
		{"3 months", market.TimeframeThreeMonths, now.Add(-90 * 24 * time.Hour), now, market.Bucket3h},
		{"year", market.TimeframeYear, now.Add(-365 * 24 * time.Hour), now, market.Bucket1d},
		{"custom 12h", market.TimeframeCustom, now.Add(-12 * time.Hour), now, market.Bucket5m},
		{"custom 10 days", market.TimeframeCustom, now.Add(-10 * 24 * time.Hour), now, market.Bucket1h},
		{"custom 60 days", market.TimeframeCustom, now.Add(-60 * 24 * time.Hour), now, market.Bucket3h},
		{"custom 200 days", market.TimeframeCustom, now.Add(-200 * 24 * time.Hour), now, market.Bucket1d},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := market.CalculateBucket(tt.timeframe, tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CalculateBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}
