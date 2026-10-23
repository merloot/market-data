package redis_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	redisreal "github.com/merloot/market-data/internal/realtime/redis"
)

func TestSubscriber_SubscribePriceUpdated(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(client)
	pub := redisreal.NewPublisher(client)

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

	select {
	case payload := <-ch:
		var got market.PriceUpdated
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.CoinName != "BTC" {
			t.Errorf("coin = %q, want BTC", got.CoinName)
		}
		if got.Price != 50000 {
			t.Errorf("price = %v, want 50000", got.Price)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestSubscriber_SubscribeMarketCapUpdated(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(client)
	pub := redisreal.NewPublisher(client)

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

	select {
	case payload := <-ch:
		var got market.MarketCapUpdated
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.CoinName != "ETH" {
			t.Errorf("coin = %q, want ETH", got.CoinName)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestSubscriber_MultipleMessages(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(client)
	pub := redisreal.NewPublisher(client)

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
	var received []string

	for i := 0; i < 3; i++ {
		select {
		case payload := <-ch:
			var got market.PriceUpdated
			_ = json.Unmarshal(payload, &got)
			mu.Lock()
			received = append(received, got.CoinName)
			mu.Unlock()
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout after %d messages", i)
		}
	}

	if len(received) != 3 {
		t.Fatalf("got %d messages, want 3", len(received))
	}
}

func TestSubscriber_ContextCancel(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	sub := redisreal.NewSubscriber(client)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer stop()

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after context cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after context cancel")
	}
}

func TestSubscriber_Cleanup(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(client)

	ch, stop, err := sub.SubscribePriceUpdated(ctx)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	stop()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after cleanup")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after cleanup")
	}
}

func TestSubscriber_Isolation(t *testing.T) {
	t.Parallel()

	client, cleanup := setupRedis(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub := redisreal.NewSubscriber(client)
	pub := redisreal.NewPublisher(client)

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
	case payload := <-priceCh:
		t.Errorf("unexpected message in price channel: %s", payload)
	case <-time.After(500 * time.Millisecond):
	}
}
