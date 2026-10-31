package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/provider/coingecko"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
	providermetrics "github.com/merloot/market-data/internal/provider/metrics"
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
		postgres.WithMetrics(),
	)
	if err != nil {
		return fmt.Errorf("Postgres: %w", err)
	}
	defer repo.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr})
	defer rdb.Close()

	cg := coingecko.New(cfg.Provider.Coingecko.ApiKey)
	cmc := coinmarketcap.New(cfg.Provider.Coinmarketcap.ApiKey)

	cgM := providermetrics.New(cg)
	cmcM := providermetrics.New(cmc)

	currencyRepository := postgres.NewCurrencyRepository(repo)
	marketDataHistoryRepository := postgres.NewMarketDataHistoryRepository(repo)
	provider := marketservice.NewService(cgM, cmcM)

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

	metricsAddr := ":9091"
	metricsSrv := &http.Server{
		Addr:              metricsAddr,
		Handler:           newMetricsMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		log.Info("worker starting", "backend", cfg.Queue.Backend)
		if err := w.Run(runCtx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- fmt.Errorf("worker: %w", err)
		}
	}()

	go func() {
		log.Info("metrics server starting", "addr", metricsAddr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-runCtx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := metricsSrv.Shutdown(shutdownCxt); err != nil {
		log.Error("metrics server shutdown failed", "err", err)
	}
	log.Info("worker shutdown")
	return nil
}

func newMetricsMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
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
		pub, _ := memory.New()
		return pub, nil
	case "redis":
		return redisrealtime.NewPublisher(rdb), nil
	default:
		return nil, fmt.Errorf("Unknown realtime backend: %q", cfg.Realtime.Backend)
	}
}
