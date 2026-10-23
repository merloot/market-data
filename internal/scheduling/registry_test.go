package scheduling_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/merloot/market-data/internal/scheduling"
)

type fakeBuilder struct {
	name  string
	tasks []scheduling.Task
	err   error
}

func (f *fakeBuilder) Name() string { return f.name }
func (f *fakeBuilder) Build(context.Context) ([]scheduling.Task, error) {
	return f.tasks, f.err
}

func TestRegistry_Tasks(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := scheduling.NewRegistry(log,
		&fakeBuilder{name: "a", tasks: []scheduling.Task{
			{Name: "1", Type: "t1", Spec: "@every 1m"},
			{Name: "2", Type: "t2", Spec: "@every 2m"},
		}},
		&fakeBuilder{name: "b", tasks: []scheduling.Task{
			{Name: "3", Type: "t3", Spec: "@every 3m"},
		}},
	)

	tasks, err := r.Tasks(context.Background())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("got %d, want 3", len(tasks))
	}
}

func TestRegistry_EmptyType(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := scheduling.NewRegistry(log,
		&fakeBuilder{name: "a", tasks: []scheduling.Task{
			{Name: "1", Type: "", Spec: "@every 1m"},
		}},
	)

	_, err := r.Tasks(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistry_BuilderError(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := scheduling.NewRegistry(log,
		&fakeBuilder{name: "a", err: errors.New("boom")},
	)

	_, err := r.Tasks(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
