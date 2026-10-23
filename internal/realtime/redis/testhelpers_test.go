package redis_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

var (
	testClient *goredis.Client
	testFlush  func(t *testing.T)
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	container, err := rediscontainer.Run(ctx, "redis:7-alpine")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start redis: %v\n", err)
		os.Exit(1)
	}

	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		fmt.Fprintf(os.Stderr, "Endpoint: %v\n", err)
		os.Exit(1)
	}

	testClient = goredis.NewClient(&goredis.Options{Addr: addr})

	testFlush = func(t *testing.T) {
		t.Helper()
		if err := testClient.FlushAll(ctx).Err(); err != nil {
			t.Fatalf("Flush: %v", err)
		}
	}

	code := m.Run()

	_ = testClient.Close()
	_ = testcontainers.TerminateContainer(container)
	os.Exit(code)
}

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

func waitClosed[T any](t *testing.T, ch <-chan T, timeout time.Duration) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed")
		}
	case <-time.After(timeout):
		t.Fatal("timeout waiting for channel close")
	}
}
