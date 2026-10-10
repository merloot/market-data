package marketservice

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type Provider interface {
	Provider() string

	GetLogo(ctx context.Context, currency string) (string, error)

	CheckCurrencyData(ctx context.Context, currencies []market.CurrencyToFind) (map[string]market.CurrencyData, error)

	GetMarketDataList(ctx context.Context, currencies []string) (map[string]market.MarketData, error)

	GetMarketDataChart(ctx context.Context, currency string, days int) (market.MarketDataChart, error)

	GetMarketDataChartRange(ctx context.Context, currency string, from, to time.Time) (market.MarketDataChart, error)
}
