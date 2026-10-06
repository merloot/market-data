package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/scheduling/asynq"
	"github.com/merloot/market-data/internal/scheduling/river"
	"github.com/merloot/market-data/internal/worker"
)

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("Worker failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	registry, err := worker.NewRegistry(log)
	if err != nil {
		return fmt.Errorf("Registry: %w", err)
	}

	log.Info("Handlers registered", "count", len(registry.Types()), "types", registry.Types())

	w, cleanup, err := buildWorker(ctx, cfg, registry, log)
	if err != nil {
		return fmt.Errorf("Build worker: %w", err)
	}
	defer cleanup()

	log.Info("Worker staring", "backend", cfg.Queue.Backend)
	if err := w.Run(ctx); err != nil {
		return err
	}

	log.Info("Worker stopped")
	return nil
}

func buildWorker(
	ctx context.Context,
	cfg *config.Config,
	registry *worker.Registry,
	log *slog.Logger,
) (worker.Worker, func(), error) {
	switch cfg.Queue.Backend {
	case "redis":
		w, err := asynq.NewWorker(cfg.Redis.Addr, registry, log)
		if err != nil {
			return nil, nil, err
		}
		return w, func() {}, nil

	case "postgres":
		w, cleanup, err := river.NewWorker(ctx, cfg.Database.URL, registry, log)
		if err != nil {
			return nil, nil, err
		}
		return w, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("Unknown queue backend: %q", cfg.Queue.Backend)
	}
}
