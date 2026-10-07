package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
	"github.com/merloot/market-data/internal/realtime/memory"
)

var _ realtime.EventPublisher = (*memory.Publisher)(nil)

func TestPublisher_PublishPriceUpdated_StoresLast(t *testing.T) {
	t.Parallel()

	p := memory.NewPublisher()
	ctx := context.Background()

	err := p.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	})
	if err != nil {
		t.Fatalf("Err = %v", err)
	}

	got, ok := p.LastPrice("BTC")
	if !ok {
		t.Fatalf("No price stored")
	}

	if got.Price != 50000 {
		t.Errorf("Price =%v, want 50000", got.Price)
	}

	if got.PriceChangePercentage24h != 5.5 {
		t.Errorf("PriceChangePercentage24h=%v, want 5.5", got.PriceChangePercentage24h)
	}
}

func TestPublisher_PublishMarketCapUpdated(t *testing.T) {
	t.Parallel()

	p := memory.NewPublisher()
	ctx := context.Background()

	err := p.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:                     "BTC",
		MarketCap:                    11150000,
		MarketCapChangePercentage24h: 15.5,
	})
	if err != nil {
		t.Fatalf("Err = %v", err)
	}

	got, ok := p.LastMarketCap("BTC")
	if !ok {
		t.Fatalf("No price stored")
	}

	if got.MarketCap != 11150000 {
		t.Errorf("MarketCap =%v, want 11150000", got.MarketCap)
	}

	if got.MarketCapChangePercentage24h != 15.5 {
		t.Errorf("MarketCapChangePercentage24hh=%v, want 15.5", got.MarketCapChangePercentage24h)
	}
}

func TestPublisher_Subscribe_ReceivesEvents(t *testing.T) {
	t.Parallel()

	p := memory.NewPublisher()
	ctx := context.Background()

	ch := p.Subscribe()
	defer p.Unsubscribe(ch)

	if err := p.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    50000,
	}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case ev := <-ch:
		if ev.Type != "price.updated" {
			t.Errorf("Type = %q, want price.updated", ev.Type)
		}
		if ev.CoinName != "BTC" {
			t.Errorf("CoinName =%q, want BTC", ev.CoinName)
		}
		if ev.Price != 50000 {
			t.Errorf("Price = %v, want 50000", ev.Price)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestPublisher_Unsubscribe_ClosesChannel(t *testing.T) {
	t.Parallel()

	p := memory.NewPublisher()
	ch := p.Subscribe()
	p.Unsubscribe(ch)

	_, ok := <-ch
	if ok {
		t.Error("Channel not closed after unsubscribe")

	}
}
