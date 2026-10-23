package memory

import (
	"context"
	"sync"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
)

var _ realtime.EventSubscriber = (*Subscriber)(nil)

type Subscriber struct {
	bus *bus
}

func (s *Subscriber) SubscribePriceUpdated(ctx context.Context) (<-chan market.PriceUpdated, func(), error) {
	ch := make(chan market.PriceUpdated, 100)

	s.bus.mu.Lock()
	s.bus.priceSubs = append(s.bus.priceSubs, ch)
	s.bus.mu.Unlock()

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			s.bus.mu.Lock()
			defer s.bus.mu.Unlock()
			for i, sub := range s.bus.priceSubs {
				if sub == ch {
					s.bus.priceSubs = append(s.bus.priceSubs[:i], s.bus.priceSubs[i+1:]...)
					close(ch)
					return
				}
			}
		})
	}

	go func() {
		<-ctx.Done()
		cleanup()
	}()

	return ch, cleanup, nil
}

func (s *Subscriber) SubscribeMarketCapUpdated(ctx context.Context) (<-chan market.MarketCapUpdated, func(), error) {
	ch := make(chan market.MarketCapUpdated, 100)

	s.bus.mu.Lock()
	s.bus.marketCapSubs = append(s.bus.marketCapSubs, ch)
	s.bus.mu.Unlock()

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			s.bus.mu.Lock()
			defer s.bus.mu.Unlock()
			for i, sub := range s.bus.marketCapSubs {
				if sub == ch {
					s.bus.marketCapSubs = append(s.bus.marketCapSubs[:i], s.bus.marketCapSubs[i+1:]...)
					close(ch)
					return
				}
			}
		})
	}

	go func() {
		<-ctx.Done()
		cleanup()
	}()

	return ch, cleanup, nil
}
