package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"

	rediscache "github.com/merloot/market-data/internal/cache/redis"
)

func setupRedis(t *testing.T) (*rediscache.Cache, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := rediscontainer.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	addr, _ := container.Endpoint(ctx, "")
	c := rediscache.New(addr)
	return c, func() {
		c.Close()
		_ = testcontainers.TerminateContainer(container)
	}
}

func TestCache_SetGetDelete(t *testing.T) {
	t.Parallel()

	c, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()

	if _, err := c.Get(ctx, "missing"); err == nil {
		t.Fatal("expected miss")
	}

	_ = c.Set(ctx, "k", []byte("v"), time.Minute)
	got, _ := c.Get(ctx, "k")
	if string(got) != "v" {
		t.Errorf("got %q", got)
	}

	_ = c.Delete(ctx, "k")
	if _, err := c.Get(ctx, "k"); err == nil {
		t.Fatal("expected miss after delete")
	}
}
