package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime/memory"
)

func TestSubscriber_SubscribePriceUpdated(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	ch, cleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cleanup()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    50000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case got := <-ch:
		if got.CoinName != "BTC" {
			t.Errorf("coin = %q, want BTC", got.CoinName)
		}
		if got.Price != 50000 {
			t.Errorf("price = %v, want 50000", got.Price)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestSubscriber_ContextCancel(t *testing.T) {
	t.Parallel()

	_, sub := memory.New()
	ctx, cancel := context.WithCancel(context.Background())

	ch, cleanup, _ := sub.SubscribePriceUpdated(ctx)
	defer cleanup()

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed")
	}
}
