package redis_test

import (
	"github.com/merloot/market-data/internal/realtime"
	"github.com/merloot/market-data/internal/realtime/redis"
)

var _ realtime.EventPublisher = (*redis.Publisher)(nil)
