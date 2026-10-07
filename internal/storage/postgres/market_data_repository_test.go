package postgres_test

import (
	"github.com/merloot/market-data/internal/storage/postgres"
	"github.com/merloot/market-data/internal/tasks/oracle"
)

var _ oracle.MarketDataHistoryRepository = (*postgres.MarketDataHistoryRepository)(nil)