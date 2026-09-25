package market

import (
	"context"
	"time"
)

type Provider interface {
	Provider() string

	GetLogo(ctx context.Context, currency string) (string, error)

	CheckCurrentData(ctx context.Context, currencies []string) (map[string]CurrencyData, error)

	GetMarketDataList(ctx context.Context, currency string, days int) (MarketDataChart, error)

	GetMarketDataRange(ctx context.Context, currency string, form, to time.Time) (MarketDataChart, error)
}
