package oracle

import (
	"context"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

type MarketDataHistoryRepository interface {
	Create(ctx context.Context, data []market.MarketDataHistory) error
	LastUpdated(ctx context.Context, currency string) (time.Time, error)

	GetMarketDataList(ctx context.Context, currencies []string, since time.Time) ([]market.MarketDataRow, error)
}
