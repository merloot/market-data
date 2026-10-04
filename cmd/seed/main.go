package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/storage/postgres"
)

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("Seed failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	repo, err := postgres.New(ctx, cfg.Database, postgres.WithLogger(log), postgres.WithPingRetries(5))

	if err != nil {
		return nil
	}

	defer repo.Close()

	return repo.SeedCurrencies(ctx)
}
