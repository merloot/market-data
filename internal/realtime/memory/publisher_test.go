package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime/memory"
)

func waitMessage[T any](t *testing.T, ch <-chan T, timeout time.Duration) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for message")
		var zero T
		return zero
	}
}

func TestPublisher_PublishPriceUpdated(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	ch, cleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cleanup()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := waitMessage(t, ch, time.Second)

	if got.CoinName != "BTC" {
		t.Errorf("coin = %q, want BTC", got.CoinName)
	}
	if got.Price != 50000 {
		t.Errorf("price = %v, want 50000", got.Price)
	}
	if got.PriceChangePercentage24h != 5.5 {
		t.Errorf("change = %v, want 5.5", got.PriceChangePercentage24h)
	}
}

func TestPublisher_PublishMarketCapUpdated(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	ch, cleanup, err := sub.SubscribeMarketCapUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cleanup()

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:                     "ETH",
		MarketCap:                    400000000000,
		MarketCapChangePercentage24h: 3.2,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := waitMessage(t, ch, time.Second)

	if got.CoinName != "ETH" {
		t.Errorf("coin = %q, want ETH", got.CoinName)
	}
	if got.MarketCap != 400000000000 {
		t.Errorf("marketCap = %v, want 4e11", got.MarketCap)
	}
	if got.MarketCapChangePercentage24h != 3.2 {
		t.Errorf("change = %v, want 3.2", got.MarketCapChangePercentage24h)
	}
}

func TestPublisher_LastPrice(t *testing.T) {
	t.Parallel()

	pub, _ := memory.New()
	ctx := context.Background()

	if _, ok := pub.LastPrice("BTC"); ok {
		t.Error("expected no last price before publish")
	}

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    50000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got, ok := pub.LastPrice("BTC")
	if !ok {
		t.Fatal("expected last price after publish")
	}
	if got.Price != 50000 {
		t.Errorf("price = %v, want 50000", got.Price)
	}

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    51000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got, _ = pub.LastPrice("BTC")
	if got.Price != 51000 {
		t.Errorf("price = %v, want 51000 (last)", got.Price)
	}
}

func TestPublisher_LastMarketCap(t *testing.T) {
	t.Parallel()

	pub,_ := memory.New()
	ctx := context.Background()

	if _, ok := pub.LastMarketCap("ETH"); ok {
		t.Error("expected no last marketcap before publish")
	}

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:  "ETH",
		MarketCap: 400000000000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got, ok := pub.LastMarketCap("ETH")
	if !ok {
		t.Fatal("expected last marketcap after publish")
	}
	if got.MarketCap != 400000000000 {
		t.Errorf("marketCap = %v, want 4e11", got.MarketCap)
	}
}

func TestPublisher_MultipleSubscribers(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	const subscribers = 3
	chans := make([]<-chan market.PriceUpdated, subscribers)
	cleanups := make([]func(), subscribers)

	for i := 0; i < subscribers; i++ {
		ch, cleanup, err := sub.SubscribePriceUpdated(ctx)
		if err != nil {
			t.Fatalf("subscribe %d: %v", i, err)
		}
		chans[i] = ch
		cleanups[i] = cleanup
	}
	defer func() {
		for _, c := range cleanups {
			c()
		}
	}()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    50000,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	for i, ch := range chans {
		got := waitMessage(t, ch, time.Second)
		if got.CoinName != "BTC" {
			t.Errorf("subscriber %d: coin = %q, want BTC", i, got.CoinName)
		}
	}
}

func TestPublisher_SlowSubscriberDoesNotBlock(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	slowCh, slowCleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe slow: %v", err)
	}
	defer slowCleanup()
	_ = slowCh

	fastCh, fastCleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe fast: %v", err)
	}
	defer fastCleanup()

	for i := 0; i < 200; i++ {
		if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
			CoinName: "BTC",
			Price:    float64(i),
		}); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}

	got := waitMessage(t, fastCh, time.Second)
	if got.CoinName != "BTC" {
		t.Errorf("coin = %q, want BTC", got.CoinName)
	}
}

func TestPublisher_Isolation(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	priceCh, priceCleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe price: %v", err)
	}
	defer priceCleanup()

	capCh, capCleanup, err := sub.SubscribeMarketCapUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe cap: %v", err)
	}
	defer capCleanup()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName: "BTC",
		Price:    50000,
	}); err != nil {
		t.Fatalf("publish price: %v", err)
	}

	got := waitMessage(t, priceCh, time.Second)
	if got.CoinName != "BTC" {
		t.Errorf("coin = %q, want BTC", got.CoinName)
	}

	select {
	case got := <-capCh:
		t.Errorf("unexpected message in cap channel: %+v", got)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestPublisher_ConcurrentPublish(t *testing.T) {
	t.Parallel()

	pub, sub := memory.New()
	ctx := context.Background()

	ch, cleanup, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cleanup()

	const workers = 10
	const perWorker = 10

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				_ = pub.PublishPriceUpdated(ctx, market.PriceUpdated{
					CoinName: "BTC",
					Price:    float64(worker*perWorker + j),
				})
			}
		}(i)
	}
	wg.Wait()

	var received int
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case <-ch:
			received++
		case <-timeout:
			break loop
		}
	}

	if received == 0 {
		t.Error("expected at least one message")
	}
}
