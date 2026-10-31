package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	pgxpool_prometheus "github.com/cmackenzie1/pgxpool-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/merloot/market-data/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

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
	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newPgxLogger(o.logger),
		LogLevel: tracelog.LogLevelDebug,
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Create pool: %w", err)
	}

	repo := &Repo{pool: pool, log: o.logger}

	if err := repo.pingWithRetry(ctx, cfg.ConnectionTimeout, o.pingRetries); err != nil {
		pool.Close()
		return nil, fmt.Errorf("Ping: %w", err)
	}

	if o.metrics {
		collector := pgxpool_prometheus.NewPgxPoolStatsCollector(pool, "marketdata")
		if err := prometheus.Register(collector); err != nil {
			var already prometheus.AlreadyRegisteredError
			if !errors.As(err, &already) {
				repo.log.Warn("pgxpool metrics register failed", "err", err)
			}
		} else {
			repo.log.Info("pgxpool metrics registered")
		}
	}

	repo.log.Info("Postgres connected", "max_conns", cfg.MaxConns, "min_cons", cfg.MinConns)

	return repo, nil
}

func (r *Repo) poolConn() *pgxpool.Pool {
	return r.pool
}

func (r *Repo) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
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
