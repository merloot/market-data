package main

import (
	"context"
	"flag"
	"fmt"
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
		slog.Error("Migrate failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("Usage: migrate <up|down|status|version up-to> [args]")
	}
	command := args[0]

	fs := flag.NewFlagSet(command, flag.ExitOnError)

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("Config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	repo, err := postgres.New(ctx, cfg.Database, postgres.WithLogger(log), postgres.WithPingRetries(5))
	if err != nil {
		return fmt.Errorf("Postgres: %w", err)
	}
	defer repo.Close()

	switch command {
	case "up":
		return repo.MigrateUp(ctx)
	case "down":
		return repo.MigrateDown(ctx)
	case "status":
		return repo.MigrateStatus(ctx)
	case "version":
		return repo.MigrateVersion(ctx)
	case "up-to":
		if len(args) < 2 {
			return fmt.Errorf("Usage: migrate up <version>")
		}
		var version int64
		if _, err := fmt.Sscanf(args[1], "%d", &version); err != nil {
			return fmt.Errorf("Invalid version %q: %w", args[1], err)
		}
		_ = fs
		return repo.MigrateUpTo(ctx, version)
	default:
		return fmt.Errorf("Unknown command: %q", command)
	}
}
