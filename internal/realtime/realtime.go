package realtime

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
)

type EventPublisher interface {
	PublishPriceUpdated(ctx context.Context, event market.PriceUpdated) error
	PublishMarketCapUpdated(ctx context.Context, event market.MarketCapUpdated) error
}

type EventSubscriber interface {
	SubscribePriceUpdated(ctx context.Context) (<-chan market.PriceUpdated, func(), error)
	SubscribeMarketCapUpdated(ctx context.Context) (<-chan market.MarketCapUpdated, func(), error)
}
