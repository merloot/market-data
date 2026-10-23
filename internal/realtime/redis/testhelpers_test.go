package redis_test

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) (*goredis.Client, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := rediscontainer.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}

	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{Addr: addr})
	return client, func() {
		_ = client.Close()
		_ = testcontainers.TerminateContainer(container)
	}
}
