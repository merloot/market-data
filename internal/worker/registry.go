package worker

import (
	"context"
	"fmt"
	"log/slog"
)

type Registry struct {
	handlers map[string]Handler
	log      *slog.Logger
}

func NewRegistry(log *slog.Logger, handlers ...Handler) (*Registry, error) {
	m := make(map[string]Handler, len(handlers))
	for _, h := range handlers {
		t := h.Type()
		if t == "" {
			return nil, fmt.Errorf("Handler %T: empty type", h)
		}
		if _, exists := m[t]; exists {
			return nil, fmt.Errorf("Duplicate handler for type %q", t)
		}
		m[t] = h
		log.Info("Handle registered", "type", t)
	}
	return &Registry{handlers: m, log: log}, nil
}

func (r *Registry) Handle(ctx context.Context, taskType string, payload []byte) error {
	h, ok := r.handlers[taskType]
	if !ok {
		return fmt.Errorf("No handler for type %q", taskType)
	}
	return h.Handle(ctx, payload)
}

func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
