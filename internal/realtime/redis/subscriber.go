package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Subscriber struct {
	client *goredis.Client
}

func NewSubscriber(client *goredis.Client) *Subscriber {
	return &Subscriber{client: client}
}

func (s *Subscriber) SubscribePriceUpdated(ctx context.Context) (<-chan []byte, func(), error) {
	return s.subscribe(ctx, PriceUpdatedChannel)
}

func (s *Subscriber) SubscribeMarketCapUpdated(ctx context.Context) (<-chan []byte, func(), error) {
	return s.subscribe(ctx, MarketCapUpdatedChannel)
}

func (s *Subscriber) subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	pubsub := s.client.Subscribe(ctx, channel)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, nil, fmt.Errorf("Subscribe %s: %w", channel, err)
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

	cleanup := func() { _ = pubsub.Close() }
	return out, cleanup, nil
}
