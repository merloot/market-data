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
	"github.com/merloot/market-data/internal/schedule"
)

type Scheduler struct {
	client *river.Client[pgx.Tx]
	log    *slog.Logger
}

func Build(
	ctx context.Context,
	dsn string,
	tasks []schedule.Task,
	log *slog.Logger,
) (*Scheduler, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("Pool: %", err)
	}

	jobs, err := buildPeriodicJobs(tasks)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	workers := river.NewWorkers()
	river.AddWorker(workers, &noopWorker{})

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers:      workers,
		Queues:       map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 1}},
		PeriodicJobs: jobs,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Create river client: %w", err)
	}
	return &Scheduler{client: client, log: log}, pool.Close, nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	s.log.Info("River scheduler starting")
	return s.client.Start(ctx)
}

func buildPeriodicJobs(tasks []schedule.Task) ([]*river.PeriodicJob, error) {
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

type noopWorker struct {
	river.WorkerDefaults[GenericArgs]
}

func (w *noopWorker) Work(ctx context.Context, job *river.Job[GenericArgs]) error {
	return fmt.Errorf("Noop worker received job: %s", job.Args.KindName)
}

type GenericArgs struct {
	KindName string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

func (g GenericArgs) Kind() string { return g.KindName }
