package oracle

import (
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

func formatChart(currency string, chart market.MarketDataChart) []market.MarketDataHistory {
	n := len(chart.Prices)
	if len(chart.MarketCaps) < n {
		n = len(chart.MarketCaps)
	}

	out := make([]market.MarketDataHistory, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, market.MarketDataHistory{
			Currency:  currency,
			Price:     chart.Prices[i][1],
			MarketCap: chart.MarketCaps[i][1],
			Timestamp: time.UnixMilli(int64(chart.Prices[i][0])).UTC(),
		})
	}
	return out
}

func calcChange(current, diff float64) float64 {
	base := current - diff
	if base == 0 {
		return 100
	}
	return (diff / base) * 100
}
