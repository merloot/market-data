package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"

	"fmt"

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi"
	"github.com/merloot/market-data/internal/httpapi/debug"
	v1 "github.com/merloot/market-data/internal/httpapi/v1"
	"github.com/merloot/market-data/internal/httpapi/v1/marketdata"
	"github.com/merloot/market-data/internal/marketservice"
	providermetrics"github.com/merloot/market-data/internal/provider/metrics"
	"github.com/merloot/market-data/internal/provider/coingecko"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
	"github.com/merloot/market-data/internal/realtime"
	"github.com/merloot/market-data/internal/realtime/memory"
	redisrealtime "github.com/merloot/market-data/internal/realtime/redis"
	"github.com/merloot/market-data/internal/storage/postgres"
)

// TODO IMPL apps

type marketDataV1Service struct {
	currencies CurrencyLister
	history    HistoryLister
}

type CurrencyLister interface {
	List(ctx context.Context) ([]market.Currency, error)
}

type HistoryLister interface {
	GetMarketDataList(ctx context.Context, currencies []string, since time.Time) ([]market.MarketDataRow, error)
	GetPriceHistory(ctx context.Context, currency market.Currency, timeframe market.Timeframe, from, to time.Time) ([]market.MarketDataPricePoint, error)
}

func newMarketDataV1Service(
	currencies CurrencyLister,
	history HistoryLister,
) *marketDataV1Service {
	return &marketDataV1Service{
		currencies: currencies,
		history:    history,
	}
}

var _ marketdata.MarketService = (*marketDataV1Service)(nil)

func (s *marketDataV1Service) ListCurrencies(ctx context.Context) ([]market.Currency, error) {
	return s.currencies.List(ctx)
}

func (s *marketDataV1Service) GetMarketDataInfo(
	ctx context.Context,
	currencies []market.Currency,
	since time.Time,
) ([]market.MarketDataRow, error) {
	names := make([]string, len(currencies))
	for i, c := range currencies {
		names[i] = c.CoinName
	}

	rows, err := s.history.GetMarketDataList(ctx, names, since)
	if err != nil {
		return nil, fmt.Errorf("history: %w", err)
	}

	info := make([]market.MarketDataRow, 0, len(rows))
	for _, r := range rows {
		info = append(info, market.MarketDataRow{
			Currency:      r.Currency,
			Price:         r.Price,
			MarketCap:     r.MarketCap,
			PriceDiff:     r.PriceDiff,
			MarketCapDiff: r.MarketCapDiff,
		})
	}
	return info, nil
}

func (s *marketDataV1Service) GetPriceHistory(
	ctx context.Context,
	c market.Currency,
	timeframe market.Timeframe,
	from, to time.Time,
) ([]market.MarketDataPricePoint, error) {
	return s.history.GetPriceHistory(ctx, c, timeframe, from, to)
}

func main() {
	_ = godotenv.Load()
	if err := run(); err != nil {
		slog.Error("api failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	initCtx := context.Background()

	repo, err := postgres.New(initCtx, cfg.Database, postgres.WithLogger(log),postgres.WithMetrics())
	if err != nil {
		return err
	}
	defer repo.Close()

	currencyRepo := postgres.NewCurrencyRepository(repo)
	historyRepo := postgres.NewMarketDataHistoryRepository(repo)

	cg := coingecko.New(cfg.Provider.Coingecko.ApiKey)
	cmc := coinmarketcap.New(cfg.Provider.Coinmarketcap.ApiKey)

	cgM := providermetrics.New(cg)
	cmcM:= providermetrics.New(cmc)
	marketSvc := marketservice.NewService(cgM, cmcM)

	debugSvc := marketSvc

	v1Svc := newMarketDataV1Service(currencyRepo, historyRepo)

	subscriber, cleanup, err := buildSubscriber(cfg, log)
	if err != nil {
		return err
	}
	defer cleanup()

	deps := &httpapi.Deps{
		Debug: &debug.Deps{
			MarketService: debugSvc,
			Log:           log,
		},
		V1: &v1.Deps{
			MarketService: v1Svc,
			Subscriber:    subscriber,
			Log:           log,
		},
		Log: log,
	}

	router := httpapi.NewRouter(deps)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("api starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Info("shutdown signal", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
	}

	log.Info("api stopped")
	return nil
}

func buildSubscriber(cfg *config.Config, log *slog.Logger) (realtime.EventSubscriber, func(), error) {
	switch cfg.Realtime.Backend {
	case "memory":
		_, sub := memory.New()
		return sub, func() {}, nil
	case "redis":
		rdb := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr})
		log.Info("realtime backend", "backend", "redis", "addr", cfg.Redis.Addr)
		return redisrealtime.NewSubscriber(rdb), func() { rdb.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("Unknown realtime backend: %q", cfg.Realtime.Backend)
	}
}
