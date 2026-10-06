package oracle

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
)

type EventPublisher interface {
	PublishPriceUpdated(ctx context.Context, event market.PriceUpdated) error 

	PublishMarketCapUpdated(ctx context.Context, event market.MarketCapUpdated) error
}