package postgres

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/merloot/market-data/internal/config"
	"github.com/pressly/goose/v3"
)

var migrationsFS embed.FS

type Repo struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func New(ctx context.Context, cfg config.DatabaseConfig, opts ...Options) (*Repo, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("Parse dsn: %w", err)
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectionTimeout

	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Create pool: %w", err)
	}

	repo := &Repo{pool: pool, log: o.logger}

	if err := repo.pingWithRetry(ctx, cfg.ConnectionTimeout, o.pingRetries); err != nil {
		pool.Close()
		return nil, fmt.Errorf("Ping: %w", err)

	}

	repo.log.Info("Postgres connected", "max_conns", cfg.MaxConns, "min_cons", cfg.MinConns)

	return repo, nil
}

func (r *Repo) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repo) Migrate(ctx context.Context) error {
	goose.SetBaseFS(migrationsFS)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Goose dialect : %w", err)
	}

	db := stdlib.OpenDBFromPool(r.pool)
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("Goose up: %w", err)
	}

	return nil
}

func (r *Repo) pingWithRetry(ctx context.Context, timeout time.Duration, retries int) error {
	if retries <= 0 {
		return r.ping(ctx, timeout)
	}

	var lastErr error
	for i := 0; i < retries; i++ {
		if err := r.ping(ctx, timeout); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Экспоненциальный backoff: 500ms, 1s,2s, ...
		delay := time.Duration(500*(1<<i)) * time.Microsecond
		r.log.Warn("Ping failed, retrying",
			"attempt", i+1,
			"delay", delay,
			"err", lastErr,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("After %d retries: %w", retries, lastErr)
}

func (r *Repo) ping(ctx context.Context, timeout time.Duration) error {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return r.pool.Ping(pingCtx)
}
