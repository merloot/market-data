package river

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
)

type noopArgs struct{}

func (noopArgs) Kind() string { return "noop" }

type noopWorker struct {
	river.WorkerDefaults[noopArgs]
}

func (w *noopWorker) Work(ctx context.Context, _ *river.Job[noopArgs]) error {
	return fmt.Errorf("Noop worker received job")
}
