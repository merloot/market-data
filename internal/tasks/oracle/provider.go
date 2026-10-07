package oracle

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type MarketDataProvider interface {
	
	GetMarketDataList(ctx context.Context, currencies []market.Currency) (map[string]market.MarketData, error)

	GetMarketDataChart(ctx context.Context, currency market.Currency, days int) (market.MarketDataChart, error)

	GetMarketDataChartRange(ctx context.Context, currency market.Currency, from, to time.Time) (market.MarketDataChart, error)

	GetLogo(ctx context.Context, currency market.Currency) (string, error)
}
