package asynq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/merloot/market-data/internal/scheduling"
)

type Scheduler struct {
	scheduler *asynq.Scheduler
	log       *slog.Logger
}

func NewSchedule(ctx context.Context, redisAddr string, tasks []scheduling.Task, log *slog.Logger) (*Scheduler, error) {
	s := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			Location: time.UTC,
			Logger:   newAsynqLogger(log),
		},
	)

	for _, task := range tasks {
		if err := registerTask(ctx, s, task, log); err != nil {
			return nil, err
		}
	}
	return &Scheduler{scheduler: s, log: log}, nil
}

func registerTask(ctx context.Context, s *asynq.Scheduler, task scheduling.Task, log *slog.Logger) error {
	t := asynq.NewTask(task.Type, task.Payload)

	entryID, err := s.Register(
		task.Spec,
		t, asynq.Queue("critical"),
		asynq.MaxRetry(3),
		asynq.Unique(55*time.Second),
	)
	if err != nil {
		return fmt.Errorf("Register %q: %w", task.Name, err)
	}

	log.Info("Cron registered",
		"name", task.Name,
		"spec", task.Spec,
		"type", task.Type,
		"entry_id", entryID,
	)

	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	s.log.Info("Asynq scheduler starting")

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.scheduler.Run()
	}()

	select {
	case <-ctx.Done():
		s.log.Info("Asynq scheduler stopping")
		s.scheduler.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (s *Scheduler) Stop(_ context.Context) error {
	s.scheduler.Shutdown()
	return nil
}
