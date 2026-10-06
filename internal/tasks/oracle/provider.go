package oracle

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
)

type MarketDataProvider interface {
	getMarketDataList(ctx context.Context, currency market.Currency) (map[string]market.MarketData, error)

	GetMarketChartRange(ctx context.Context, currency market.Currency, days int) (market.MarketDataChart, error)

	GetLogo(ctx context.Context, currency market.Currency) (string, error)
}
