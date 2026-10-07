package oracle

import (
	"context"
	"fmt"
	"log/slog"
)

type Handler struct {
	scv *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{scv: svc, log: log}
}

func (h *Handler) Type() string { return Type }

func (h *Handler) Handle(ctx context.Context, payload []byte) error {
	h.log.Info("Oracle task started")
	if err := h.scv.Execute(ctx); err != nil {
		return fmt.Errorf("Execute: %w", err)
	}
	h.log.Info("Oracle task done")
	return nil
}
