package scheduling

import (
	"context"
	"fmt"
	"log/slog"
)

type Builder interface {
	Name() string
	Build() ([]Task, error)
}

type Registry struct {
	builders []Builder
	log      *slog.Logger
}

func NewRegistry(log *slog.Logger, builders ...Builder) *Registry {
	return &Registry{log: log, builders: builders}
}

func (r *Registry) Tasks(ctx context.Context) ([]Task, error) {
	var tasks []Task
	for _, b := range r.builders {
		built, err := b.Build()
		if err != nil {
			return nil, fmt.Errorf("Builder %q: %w", b.Name(), err)
		}
		r.log.Info("Tasks builder", "builder", b.Name(), "count", len(built))
		tasks = append(tasks, built...)
	}
	return tasks, nil
}
