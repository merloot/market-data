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

	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/httpapi"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/provider/coingecko"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
)

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
		return fmt.Errorf("config: %w", err)
	}

	cg := coingecko.New(cfg.Provider.Coingecko.ApiKey)
	cmc := coinmarketcap.New(cfg.Provider.Coinmarketcap.ApiKey)

	marketSvc := marketservice.NewService(cg, cmc)

	deps := &httpapi.Deps{MarketService: marketSvc}
	router := httpapi.NewRouter(deps, log)

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
