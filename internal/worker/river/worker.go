package river

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/merloot/market-data/internal/worker"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type Worker struct {
	client *river.Client[pgx.Tx]
	log    *slog.Logger
}

type oracleWorker struct {
	river.WorkerDefaults[oracleArgs]
	registry *worker.Registry
	log      *slog.Logger
}

func (w *oracleWorker) Work(ctx context.Context, job *river.Job[oracleArgs]) error {
	w.log.Info("River job receiver", "type", job.Kind)
	return w.registry.Handle(ctx, job.Kind, job.Args.Payload)
}

func NewWorker(
	ctx context.Context,
	dsn string,
	registry *worker.Registry,
	log *slog.Logger,
) (*Worker, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("Pool: %w", err)
	}

	types := make(map[string]bool)
	workers := river.NewWorkers()
	for _, t := range registry.Types() {
		types[t] = true

		// TODO
		if types["market-data-oracle"] {
			river.AddWorker(workers, &oracleWorker{
				registry: registry,
				log:      log,
			})
			log.Info("River handler registered", "type", t)
		}
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
	})
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("River client: %w", err)
	}

	return &Worker{client: client, log: log}, pool.Close, nil
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("River worker starting", "ctxErr", ctx.Err())

	startErrCh := make(chan error, 1)
	go func() {
		startErrCh <- w.client.Start(ctx)
	}()

	<-ctx.Done()
	w.log.Info("River worker stopping", "ctxErr", ctx.Err())

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := w.client.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		w.log.Error("River client stop", "err", err)
	}

	return ctx.Err()
}
