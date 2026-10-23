package redis_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
	redisreal "github.com/merloot/market-data/internal/realtime/redis"
)

var _ realtime.EventPublisher = (*redisreal.Publisher)(nil)

func TestPublisher_PublishPriceUpdated(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := testClient.Subscribe(ctx, redisreal.PriceUpdatedChannel)
	defer sub.Close()

	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	p := redisreal.NewPublisher(testClient)

	if err := p.PublishPriceUpdated(ctx, market.PriceUpdated{
		CoinName:                 "BTC",
		Price:                    50000,
		PriceChangePercentage24h: 5.5,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		var got market.PriceUpdated
		if err := json.Unmarshal([]byte(msg.Payload), &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.CoinName != "BTC" {
			t.Errorf("coin = %q, want BTC", got.CoinName)
		}
		if got.Price != 50000 {
			t.Errorf("price = %v, want 50000", got.Price)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestPublisher_PublishMarketCapUpdated(t *testing.T) {
	testFlush(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sub := testClient.Subscribe(ctx, redisreal.MarketCapUpdatedChannel)
	defer sub.Close()

	if _, err := sub.Receive(ctx); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	p := redisreal.NewPublisher(testClient)

	if err := p.PublishMarketCapUpdated(ctx, market.MarketCapUpdated{
		CoinName:                     "ETH",
		MarketCap:                    400000000000,
		MarketCapChangePercentage24h: 3.2,
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		var got market.MarketCapUpdated
		if err := json.Unmarshal([]byte(msg.Payload), &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.CoinName != "ETH" {
			t.Errorf("coin = %q, want ETH", got.CoinName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}
