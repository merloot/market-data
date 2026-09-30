package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/merloot/market-data/internal/cache"
	"github.com/merloot/market-data/internal/config"
	"github.com/merloot/market-data/internal/storage/postgres"
)

type App struct {
	Config *config.Config
	Logger *slog.Logger
	Repo   *postgres.Repo
	Cache  cache.Cache
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	repo, err := postgres.New(
		ctx,
		cfg.Database,
		postgres.WithLogger(logger),
		postgres.WithPingRetries(5),
	)

	if err != nil {
		return nil, err
	}

	return &App{
		Config: cfg,
		Logger: logger,
		Repo:   repo,
	}, nil
}

func (a *App) Close() {
	a.Repo.Close()
}
