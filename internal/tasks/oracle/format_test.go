package oracle

import (
	"testing"

	"github.com/merloot/market-data/internal/domain/market"
)

func TestCalcChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		currency float64
		diff     float64
		want     float64
	}{
		{"zero diff", 100, 0, 0},
		{"positive", 110, 10, 10},
		{"negative", 90, -10, -10},
		{"zero base", 0, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcChange(tt.currency, tt.diff)
			if got != tt.want {
				t.Errorf("calcChange(%v, %v) = %v, want %v", tt.currency, tt.diff, got, tt.want)
			}
		})
	}
}

func TestFormatChart(t *testing.T) {
	t.Parallel()

	chart := market.MarketDataChart{
		Prices: [][2]float64{
			{17000000000, 50000},
			{17000006000, 50100},
		},
		MarketCaps: [][2]float64{
			{17000000000, 10000000000},
			{17000006000, 10010000000},
		},
	}

	got := formatChart("BTC", chart)

	if len(got) != 2 {
		t.Fatalf("Got %d points, want 2", len(got))
	}

	if got[0].Currency != "BTC" {
		t.Errorf("Currency =%q, want BTC", got[0].Currency)
	}

	if got[0].Price != 50000 {
		t.Errorf("Price = %v want 50000", got[0].Price)
	}

	if got[1].Price != 50100 {
		t.Errorf("Price = %v want 50100", got[1].Price)
	}

	if got[0].Timestamp.IsZero() {
		t.Error("TImestamp is zero")
	}
}
