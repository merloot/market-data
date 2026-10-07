package river

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/robfig/cron/v3"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/merloot/market-data/internal/scheduling"
)

type Scheduler struct {
	client *river.Client[pgx.Tx]
	log    *slog.Logger
	tasks  []scheduling.Task
}

func NewSchedule(
	ctx context.Context,
	dsn string,
	tasks []scheduling.Task,
	log *slog.Logger,
) (*Scheduler, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("Pool: %w", err)
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, &noopWorker{})

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		LeaderElectionDisabled: true,
		SoftStopTimeout:        5 * time.Second,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Create river client: %w", err)
	}
	return &Scheduler{client: client, log: log, tasks: tasks}, pool.Close, nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	s.log.Info("River scheduler starting", "tasks", len(s.tasks))

	c := cron.New()
	for _, task := range s.tasks {
		t := task
		if _, err := c.AddFunc(t.Spec, func() {
			s.enqueue(ctx, t)
		}); err != nil {
			return fmt.Errorf("Add cron %q: %w", t.Name, err)
		}
		s.log.Info("Cron registered", "name", t.Name, "spec", t.Spec, "type", t.Type)
	}
	c.Start()
	defer c.Stop()

	<-ctx.Done()
	s.log.Info("River scheduler stopping")
	return ctx.Err()
}

func (s *Scheduler) Stop(ctx context.Context) error {
	return nil
}

func (s *Scheduler) enqueue(ctx context.Context, task scheduling.Task) {
	_, err := s.client.Insert(ctx, GenericArgs{
		KindName: task.Type,
		Payload:  task.Payload,
	}, &river.InsertOpts{
		Queue: river.QueueDefault,
		UniqueOpts: river.UniqueOpts{
			ByPeriod: time.Minute,
		},
	})
	if err != nil {
		s.log.Error("Enqueue", "name", task.Name, "err", err)
		return
	}
	s.log.Info("Task enqueued", "name", task.Name, "type", task.Type)
}

func buildPeriodicJobs(tasks []scheduling.Task) ([]*river.PeriodicJob, error) {
	jobs := make([]*river.PeriodicJob, 0, len(tasks))
	for _, task := range tasks {
		schedule, err := parseSpec(task.Spec)
		if err != nil {
			return nil, fmt.Errorf("Parse %q: %w", task.Name, err)
		}

		t := task
		jobs = append(jobs, river.NewPeriodicJob(
			schedule,
			func() (river.JobArgs, *river.InsertOpts) {
				return GenericArgs{
					KindName: t.Type,
					Payload:  t.Payload,
				}, &river.InsertOpts{
					Queue: river.QueueDefault,
					UniqueOpts: river.UniqueOpts{
						ByPeriod: time.Minute,
					},
				}
			}, &river.PeriodicJobOpts{RunOnStart: true},
		))
	}
	return jobs, nil
}

func parseSpec(spec string) (river.PeriodicSchedule, error) {
	if len(spec) > 7 && spec[:7] == "@every " {
		d, err := time.ParseDuration(spec[7:])
		if err != nil {
			return nil, fmt.Errorf("Parse @every: %w", err)
		}
		return river.PeriodicInterval(d), nil
	}
	return cron.ParseStandard(spec)
}

type GenericArgs struct {
	KindName string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

func (g GenericArgs) Kind() string { return g.KindName }
