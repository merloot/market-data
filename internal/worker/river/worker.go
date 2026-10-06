package river

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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

	workers := river.NewWorkers()
	for _, t := range registry.Types() {
		river.AddWorker(workers, &genericWorker{
			taskType: t,
			registry: registry,
			log:      log,
		})
		log.Info("River handler registered", "type", t)
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
	w.log.Info("River worker starting")
	return w.client.Start(ctx)
}

type jobArgs struct {
	KindName string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

func (j jobArgs) Kind() string { return j.KindName }

type genericWorker struct {
	river.WorkerDefaults[jobArgs]
	taskType string
	registry *worker.Registry
	log      *slog.Logger
}

func (w *genericWorker) Work(ctx context.Context, job *river.Job[jobArgs]) error {
	w.log.Info("River job received", "type", w.taskType)
	return w.registry.Handle(ctx, w.taskType, job.Args.Payload)
}
