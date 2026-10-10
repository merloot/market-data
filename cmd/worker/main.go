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
	goredis "github.com/redis/go-redis/v9"

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/provider/coingecko"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
	"github.com/merloot/market-data/internal/realtime"
	"github.com/merloot/market-data/internal/realtime/memory"
	redisrealtime "github.com/merloot/market-data/internal/realtime/redis"
	"github.com/merloot/market-data/internal/storage/postgres"
	"github.com/merloot/market-data/internal/tasks/oracle"
	"github.com/merloot/market-data/internal/worker"
	"github.com/merloot/market-data/internal/worker/asynq"
	"github.com/merloot/market-data/internal/worker/river"
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

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr})
	defer rdb.Close()

	cg := coingecko.New(cfg.Provider.Coingecko.ApiKey)
	coinmarketcap := coinmarketcap.New(cfg.Provider.Coinmarketcap.ApiKey)

	currencyRepository := postgres.NewCurrencyRepository(repo)
	marketDataHistoryRepository := postgres.NewMarketDataHistoryRepository(repo)
	provider := marketservice.NewService(cg, coinmarketcap)

	publisher, err := buildPublisher(cfg, rdb)
	if err != nil {
		return fmt.Errorf("Publisher: %w", err)
	}

	oracleService := oracle.NewService(
		log,
		currencyRepository,
		marketDataHistoryRepository,
		provider,
		publisher,
	)

	registry, err := worker.NewRegistry(log,
		oracle.NewHandler(oracleService, log),
	)
	if err != nil {
		return fmt.Errorf("Registry: %w", err)
	}

	log.Info("Handlers registered",
		"count", len(registry.Types()),
		"types", registry.Types(),
	)

	w, cleanup, err := buildWorker(initCtx, cfg, registry, log)
	if err != nil {
		return fmt.Errorf("Build worker: %w", err)
	}
	defer cleanup()

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("Worker starting", "backend", cfg.Queue.Backend)
	if err := w.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("Run: %w", err)
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

func buildPublisher(cfg *config.Config, rdb *goredis.Client) (realtime.EventPublisher, error) {
	switch cfg.Realtime.Backend {
	case "memory":
		return memory.NewPublisher(), nil
	case "redis":
		return redisrealtime.NewPublisher(rdb), nil
	default:
		return nil, fmt.Errorf("Unknown realtime backend: %q", cfg.Realtime.Backend)
	}
}
