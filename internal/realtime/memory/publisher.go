package memory

import (
	"context"
	"sync"

	"github.com/merloot/market-data/internal/domain/market"
)

type Publisher struct {
	mu          sync.RWMutex
	prices      map[string]market.PriceUpdated
	marketCaps  map[string]market.MarketCapUpdated
	subscribers []chan Event
}

type Event struct {
	Type                     string  `json:"type"`
	CoinName                 string  `json:"coinName"`
	Price                    float64 `json:"price,omitempty"`
	PriceChangePercentage24h float64 `json:"priceChangePercentage24h,omitempty"`
	MarketCap                float64 `json:"marketCap,omitempty"`
	MarketCapPercentage24h   float64 `json:"marketCapPercentage24h,omitempty"`
}

func NewPublisher() *Publisher {
	return &Publisher{
		prices:     make(map[string]market.PriceUpdated),
		marketCaps: make(map[string]market.MarketCapUpdated),
	}
}

func (p *Publisher) PublishPriceUpdated(_ context.Context, event market.PriceUpdated) error {
	p.mu.Lock()
	p.prices[event.CoinName] = event
	subs := p.subscribers
	p.mu.Unlock()

	ev := Event{
		Type:                     "price.updated",
		CoinName:                 event.CoinName,
		Price:                    event.Price,
		PriceChangePercentage24h: event.PriceChangePercentage24h,
	}

	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
	return nil
}

func (p *Publisher) PublishMarketCapUpdated(_ context.Context, event market.MarketCapUpdated) error {
	p.mu.Lock()
	p.marketCaps[event.CoinName] = event
	subs := p.subscribers
	p.mu.Unlock()

	ev := Event{
		Type:                   "marketcap.updated",
		CoinName:               event.CoinName,
		MarketCap:              event.MarketCap,
		MarketCapPercentage24h: event.MarketCapChangePercentage24h,
	}

	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
	return nil
}

func (p *Publisher) LastPrice(coinName string) (market.PriceUpdated, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	e, ok := p.prices[coinName]
	return e, ok
}

func (p *Publisher) LastMarketCap(coinName string) (market.MarketCapUpdated, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	e, ok := p.marketCaps[coinName]
	return e, ok
}

func (p *Publisher) Subscribe() <-chan Event {
	ch := make(chan Event, 100)
	p.mu.Lock()
	p.subscribers = append(p.subscribers, ch)
	p.mu.Unlock()
	return ch
}

func (p *Publisher) Unsubscribe(ch <-chan Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, sub := range p.subscribers {
		if sub == ch {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			close(sub)
			return
		}
	}
}
