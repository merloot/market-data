package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	pg "github.com/merloot/market-data/internal/storage/postgres"
)

func setupPostgres(t *testing.T) (*pg.Repo, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpha",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Database system is ready to accept connection").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Start container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Connection string: %v", err)
	}

	repo, err := pg.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Repo: %v", err)
	}

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate :%v", err)
	}
	return repo, func() {
		repo.Close()
		_ = container.Terminate(ctx)
	}
}
