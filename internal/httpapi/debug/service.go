package debug

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
)

type MarketService interface {
	GetMarketDataList(ctx context.Context, currencies []market.Currency) (map[string]market.MarketData, error)
	GetMarketDataChart(ctx context.Context, currency market.Currency, days int) (market.MarketDataChart, error)
}
