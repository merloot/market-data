package marketdata

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type MarketService interface {
	ListCurrencies(ctx context.Context) ([]market.Currency, error)
	GetMarketDataInfo(ctx context.Context, currencies []market.Currency, since time.Time) ([]market.MarketDataRow, error)
	GetPriceHistory(ctx context.Context, currency market.Currency, from, to time.Time) (market.MarketDataChart, error)
}
