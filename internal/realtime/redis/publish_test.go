package redis_test

import (
	"github.com/merloot/market-data/internal/realtime/redis"
	"github.com/merloot/market-data/internal/tasks/oracle"
)

var _ oracle.EventPublisher = (*redis.Publisher)(nil)
