package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var migrationDIR = "migrations"

func init() {
	goose.SetBaseFS(migrationsFS)
}

func (r *Repo) MigrateUp(ctx context.Context) error {
	return r.runGoose(ctx, func(db *sql.DB) error {
		return goose.UpContext(ctx, db, migrationDIR)
	}, "up")
}

func (r *Repo) MigrateDown(ctx context.Context) error {
	return r.runGoose(ctx, func(db *sql.DB) error {
		return goose.DownContext(ctx, db, migrationDIR)
	}, "down")
}

func (r *Repo) MigrateStatus(ctx context.Context) error {
	return r.runGoose(ctx, func(db *sql.DB) error {
		return goose.StatusContext(ctx, db, migrationDIR)
	}, "status")
}

func (r *Repo) MigrateVersion(ctx context.Context) error {
	return r.runGoose(ctx, func(db *sql.DB) error {
		v, err := goose.GetDBVersionContext(ctx, db)
		if err != nil {
			return err
		}
		r.log.Info("Current migration version", "version", v)
		return nil
	}, "version")
}

func (r *Repo) MigrateUpTo(ctx context.Context, version int64) error {
	return r.runGoose(ctx, func(db *sql.DB) error {
		return goose.UpToContext(ctx, db, migrationDIR, version)
	}, "up-to")
}

func (r *Repo) runGoose(ctx context.Context, fn func(*sql.DB) error, op string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Goose dialect: %w", err)
	}
	db := stdlib.OpenDBFromPool(r.pool)
	defer db.Close()

	r.log.Info("Migrate start", "op", op)
	if err := fn(db); err != nil {
		return fmt.Errorf("Goose %s: %w", op, err)
	}

	r.log.Info("Migrate done", "op", op)
	return nil
}
