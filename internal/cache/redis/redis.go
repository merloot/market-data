package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrMiss = errors.New("Cache miss")

type Cache struct {
	client *redis.Client
}

func New(addr string) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (c *Cache) Close() {
	c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, ErrMiss
	}

	if err != nil {
		return nil, fmt.Errorf("Redis get: %w", err)
	}

	return val, nil
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("Redis set %w", err)
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("Redis delete %w", err)
	}
	return nil
}
