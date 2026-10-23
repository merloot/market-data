package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
	redisreal "github.com/merloot/market-data/internal/realtime/redis"
)

var _ realtime.EventSubscriber = (*redisreal.Subscriber)(nil)

func TestSubscriber_SubscribePriceUpdated(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)
	pub := redisreal.NewPublisher(testClient)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := waitMessage(t, ch, 2*time.Second)

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

func TestSubscriber_SubscribeMarketCapUpdated(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)
	pub := redisreal.NewPublisher(testClient)

	ch, stop, err := sub.SubscribeMarketCapUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:                     "ETH",
		MarketCap:                    400000000000,
		MarketCapChangePercentage24h: 3.2,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := waitMessage(t, ch, 2*time.Second)

	if got.CoinName != "ETH" {
		t.Errorf("coin = %q, want ETH", got.CoinName)
	}
	if got.MarketCap != 400000000000 {
		t.Errorf("marketCap = %v, want 4e11", got.MarketCap)
	}
}

func TestSubscriber_MultipleMessages(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)
	pub := redisreal.NewPublisher(testClient)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	coins := []string{"BTC", "ETH", "SOL"}
	for _, coin := range coins {
		if err := pub.PublishPriceUpdated(ctx, market.PriceUpdated{
			CoinName: coin,
			Price:    100,
		}); err != nil {
			t.Fatalf("publish %s: %v", coin, err)
		}
	}

	var mu sync.Mutex
	received := make([]string, 0, 3)

	for i := 0; i < 3; i++ {
		got := waitMessage(t, ch, 2*time.Second)
		mu.Lock()
		received = append(received, got.CoinName)
		mu.Unlock()
	}

	if len(received) != 3 {
		t.Fatalf("got %d messages, want 3", len(received))
	}
}

func TestSubscriber_ContextCancel(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithCancel(context.Background())

	sub := redisreal.NewSubscriber(testClient)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	cancel()

	waitClosed(t, ch, 2*time.Second)
}

func TestSubscriber_Cleanup(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	stop()

	waitClosed(t, ch, 2*time.Second)
}

func TestSubscriber_Isolation(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)
	pub := redisreal.NewPublisher(testClient)

	priceCh, stopPrice, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe price: %v", err)
	}
	defer stopPrice()

	if err := pub.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:  "ETH",
		MarketCap: 100,
	}); err != nil {
		t.Fatalf("publish cap: %v", err)
	}

	select {
	case got := <-priceCh:
		t.Errorf("unexpected message in price channel: %+v", got)
	case <-time.After(500 * time.Millisecond):
	}
}

func TestSubscriber_JSONDecodeError_Skipped(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(testClient)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	if err := testClient.Publish(ctx, redisreal.PriceUpdatedChannel, "not json").Err(); err != nil {
		t.Fatalf("publish raw: %v", err)
	}

	select {
	case got := <-ch:
		t.Errorf("unexpected message: %+v", got)
	case <-time.After(500 * time.Millisecond):
	}
}
