package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/merloot/market-data/internal/domain/market"
	goredis "github.com/redis/go-redis/v9"
)

type Publisher struct {
	client *goredis.Client
}

func NewPublisher(client *goredis.Client) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) PublishPriceUpdated(ctx context.Context, event market.PriceUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}

	if err := p.client.Publish(ctx, "price.updated", payload).Err(); err != nil {
		return fmt.Errorf("Publish: %w", err)
	}

	return nil
}

func (p *Publisher) PublishMarketCapUpdated(ctx context.Context, event market.MarketCapUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}

	if err := p.client.Publish(ctx, "marketcap.updated", payload).Err(); err != nil {
		return fmt.Errorf("Publish: %w", err)
	}
	return nil
}
