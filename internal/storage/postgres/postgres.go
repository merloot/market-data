package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

var migrationsFS embed.FS

type Repo struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Repo, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Ping: %w", err)
	}

	return &Repo{pool: pool}, nil
}

func (r *Repo) Close() {
	r.pool.Close()
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
