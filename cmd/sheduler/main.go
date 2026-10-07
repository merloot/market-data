package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/scheduling"
	"github.com/merloot/market-data/internal/scheduling/asynq"
	"github.com/merloot/market-data/internal/scheduling/river"
	"github.com/merloot/market-data/internal/storage/postgres"
	"github.com/merloot/market-data/internal/tasks/oracle"
)

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("Scheduler failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Config: %w", err)
	}

	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	repo, err := postgres.New(initCtx, cfg.Database,
		postgres.WithLogger(log),
		postgres.WithPingRetries(5),
	)
	if err != nil {
		return fmt.Errorf("Postgres: %w", err)
	}
	defer repo.Close()

	registry := scheduling.NewRegistry(log,
		oracle.NewBuild(),
	)

	tasks, err := registry.Tasks(initCtx)
	if err != nil {
		return fmt.Errorf("Build tasks: %w", err)
	}
	if len(tasks) == 0 {
		return fmt.Errorf("No tasks to schedule")
	}

	log.Info("Tasks ready", "total", len(tasks), "backend", cfg.Queue.Backend)

	scheduler, cleanup, err := buildScheduler(initCtx, cfg, tasks, log)
	if err != nil {
		return fmt.Errorf("Build scheduler: %w", err)
	}
	defer cleanup()

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("scheduler starting", "backend", cfg.Queue.Backend)
	if err := scheduler.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("Run: %w", err)
	}

	log.Info("Scheduler stopped")
	return nil
}

func buildScheduler(
	ctx context.Context,
	cfg *config.Config,
	tasks []scheduling.Task,
	log *slog.Logger,
) (scheduling.Scheduler, func(), error) {
	switch cfg.Queue.Backend {
	case "redis":
		s, err := asynq.NewSchedule(ctx, cfg.Redis.Addr, tasks, log)
		if err != nil {
			return nil, nil, err
		}
		return s, func() {}, nil

	case "postgres":
		s, cleanup, err := river.NewSchedule(ctx, cfg.Database.URL, tasks, log)
		if err != nil {
			return nil, nil, err
		}
		return s, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("Unknown queue backend: %q", cfg.Queue.Backend)
	}
}
