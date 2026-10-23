package scheduling

import (
	"context"
	"fmt"
	"log/slog"
)

type Builder interface {
	Name() string
	Build(context.Context) ([]Task, error)
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
		built, err := b.Build(ctx)
		if err != nil {
			return nil, fmt.Errorf("builder %q: %w", b.Name(), err)
		}

		for _, t := range built {
			if t.Type == "" {
				return nil, fmt.Errorf("builder %q: task %q has empty Type", b.Name(), t.Name)
			}
			if t.Spec == "" {
				return nil, fmt.Errorf("builder %q: task %q has empty Spec", b.Name(), t.Name)
			}
		}

		r.log.Info("tasks built", "builder", b.Name(), "count", len(built))
		tasks = append(tasks, built...)
	}
	return tasks, nil
}
