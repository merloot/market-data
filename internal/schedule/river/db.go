package river

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func NewClient(ctx context.Context, dns string, cfg *river.Config) (*river.Client[pgx.Tx], func(), error) {
	pool, err := pgxpool.New(ctx, dns)
	if err != nil {
		return nil, nil, fmt.Errorf("Pool: %w", err)
	}

	client, err := river.NewClient(riverpgxv5.New(pool), cfg)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("River client: %w", err)
	}
	cleanup := func() {
		pool.Close()
	}
	return client, cleanup, nil
}
