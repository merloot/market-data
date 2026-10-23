package memory

import (
	"sync"

	"github.com/merloot/market-data/internal/domain/market"
)

type bus struct {
	mu            sync.RWMutex
	priceSubs     []chan market.PriceUpdated
	marketCapSubs []chan market.MarketCapUpdated

	prices     map[string]market.PriceUpdated
	marketCaps map[string]market.MarketCapUpdated
}

func newBus() *bus {
	return &bus{
		prices:     make(map[string]market.PriceUpdated),
		marketCaps: make(map[string]market.MarketCapUpdated),
	}
}

func (b *bus) publishPrice(event market.PriceUpdated) {
	b.mu.Lock()
	b.prices[event.CoinName] = event
	subs := b.priceSubs
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *bus) publishMarketCap(event market.MarketCapUpdated) {
	b.mu.Lock()
	b.marketCaps[event.CoinName] = event
	subs := b.marketCapSubs
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
