package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/realtime"
	goredis "github.com/redis/go-redis/v9"
)

var _ realtime.EventSubscriber = (*Subscriber)(nil)

type Subscriber struct {
	client *goredis.Client
}

func NewSubscriber(client *goredis.Client) *Subscriber {
	return &Subscriber{client: client}
}

func (s *Subscriber) SubscribePriceUpdated(ctx context.Context) (<-chan market.PriceUpdated, func(), error) {
	return subscribeJSON[market.PriceUpdated](ctx, s.client, PriceUpdatedChannel)
}

func (s *Subscriber) SubscribeMarketCapUpdated(ctx context.Context) (<-chan market.MarketCapUpdated, func(), error) {
	return subscribeJSON[market.MarketCapUpdated](ctx, s.client, MarketCapUpdatedChannel)
}

func subscribeJSON[T any](ctx context.Context, client *goredis.Client, channel string) (<-chan T, func(), error) {
	raw, cleanup, err := subscribeRaw(ctx, client, channel)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan T, 100)

	go func() {
		defer close(out)
		for payload := range raw {
			var event T
			if err := json.Unmarshal(payload, &event); err != nil {
				continue
			}

			select {
			case out <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, cleanup, nil
}

func subscribeRaw(ctx context.Context, client *goredis.Client, channel string) (<-chan []byte, func(), error) {
	pubsub := client.Subscribe(ctx, channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, nil, fmt.Errorf("subscribe %s: %w", channel, err)
	}

	out := make(chan []byte, 100)

	go func() {
		<-ctx.Done()
		_ = pubsub.Close()
	}()

	go func() {
		defer close(out)
		for msg := range pubsub.Channel() {
			select {
			case out <- []byte(msg.Payload):
			case <-ctx.Done():
				return
			}
		}
	}()

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			_ = pubsub.Close()
		})
	}
	return out, cleanup, nil
}
