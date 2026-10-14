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

	"fmt"

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi"
	"github.com/merloot/market-data/internal/httpapi/debug"
	v1 "github.com/merloot/market-data/internal/httpapi/v1"
	"github.com/merloot/market-data/internal/httpapi/v1/marketdata"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/provider/coingecko"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
	"github.com/merloot/market-data/internal/storage/postgres"
)

// TODO IMPL apps

type marketDataV1Service struct {
	currencies CurrencyLister
	market     *marketservice.Service
	history    HistoryLister
}

type CurrencyLister interface {
	List(ctx context.Context) ([]market.Currency, error)
}

type HistoryLister interface {
	GetMarketDataList(ctx context.Context, currencies []string, since time.Time) ([]market.MarketDataRow, error)
}

func newMarketDataV1Service(
	currencies CurrencyLister,
	market *marketservice.Service,
	history HistoryLister,
) *marketDataV1Service {
	return &marketDataV1Service{
		currencies: currencies,
		market:     market,
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
	from, to time.Time,
) (market.MarketDataChart, error) {
	return s.market.GetMarketDataChartRange(ctx, c, from, to)
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

	repo, err := postgres.New(initCtx, cfg.Database, postgres.WithLogger(log))
	if err != nil {
		return err
	}
	defer repo.Close()

	currencyRepo := postgres.NewCurrencyRepository(repo)
	historyRepo := postgres.NewMarketDataHistoryRepository(repo)

	cg := coingecko.New(cfg.Provider.Coingecko.ApiKey)
	cmc := coinmarketcap.New(cfg.Provider.Coinmarketcap.ApiKey)
	marketSvc := marketservice.NewService(cg, cmc)

	debugSvc := marketSvc

	v1Svc := newMarketDataV1Service(currencyRepo, marketSvc, historyRepo)

	deps := &httpapi.Deps{
		Debug: &debug.Deps{
			MarketService: debugSvc,
			Log:           log,
		},
		V1: &v1.Deps{
			MarketService: v1Svc,
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
