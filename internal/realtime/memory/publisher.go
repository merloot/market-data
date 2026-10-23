package memory

import (
	"context"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
)

var _ realtime.EventPublisher = (*Publisher)(nil)

type Publisher struct {
	bus *bus
}

func New() (*Publisher, *Subscriber) {
	b := newBus()
	return &Publisher{bus: b}, &Subscriber{bus: b}
}

func (p *Publisher) PublishPriceUpdated(_ context.Context, event market.PriceUpdated) error {
	p.bus.publishPrice(event)
	return nil
}

func (p *Publisher) PublishMarketCapUpdated(_ context.Context, event market.MarketCapUpdated) error {
	p.bus.publishMarketCap(event)
	return nil
}

func (p *Publisher) LastPrice(coinName string) (market.PriceUpdated, bool) {
	p.bus.mu.RLock()
	defer p.bus.mu.RUnlock()

	e, ok := p.bus.prices[coinName]
	return e, ok
}

func (p *Publisher) LastMarketCap(coinName string) (market.MarketCapUpdated, bool) {
	p.bus.mu.RLock()
	defer p.bus.mu.RUnlock()

	e, ok := p.bus.marketCaps[coinName]
	return e, ok
}
