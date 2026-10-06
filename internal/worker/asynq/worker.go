package asynq

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/merloot/market-data/internal/worker"
)

type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	log    *slog.Logger
}

func NewWorker(redisAddr string, registry *worker.Registry, log *slog.Logger) (*Worker, error) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: newAsynqLogger(log),
		},
	)

	mux := asynq.NewServeMux()
	for _, t := range registry.Types() {
		taskType := t
		mux.HandleFunc(taskType, func(ctx context.Context, t *asynq.Task) error {
			return registry.Handle(ctx, taskType, t.Payload())
		})
		log.Info("Asynq handler registered", "type", taskType)
	}

	return &Worker{server: srv, mux: mux, log: log}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("Asynq worker starting")

	errCh := make(chan error, 1)
	go func() {
		errCh <- w.server.Run(w.mux)
	}()

	select {
	case <-ctx.Done():
		w.log.Info("Asynq worker stopping")
		w.server.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
