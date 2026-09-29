package market

import (
	"context"
	"time"
)

type Provider interface {
	Provider() string

	GetLogo(ctx context.Context, currency string) (string, error)

	CheckCurrencyData(ctx context.Context, currencies []CurrencyToFind) (map[string]CurrencyData, error)

	GetMarketDataList(ctx context.Context, currencies []string) (map[string]MarketData, error)

	GetMarketDataChart(ctx context.Context, currency string, days int) (MarketDataChart, error)

	GetMarketDataChartRange(ctx context.Context, currency string, from, to time.Time) (MarketDataChart, error)
}
