package worker_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/merloot/market-data/internal/worker"
)

type fakeHandler struct {
	typ     string
	handled []byte
	err     error
}

func (f *fakeHandler) Type() string { return f.typ }
func (f *fakeHandler) Handle(_ context.Context, p []byte) error {
	f.handled = p
	return f.err
}

func TestRegistry_Handle(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := &fakeHandler{typ: "quote:fetch"}

	r, err := worker.NewRegistry(log, h)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	if err := r.Handle(context.Background(), "quote:fetch", []byte(`{}`)); err != nil {
		t.Fatalf("err = %v", err)
	}
	if string(h.handled) != `{}` {
		t.Errorf("handled = %q", h.handled)
	}
}

func TestRegistry_UnknownType(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r, _ := worker.NewRegistry(log)

	err := r.Handle(context.Background(), "unknown", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistry_DuplicateType(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err := worker.NewRegistry(log,
		&fakeHandler{typ: "x"},
		&fakeHandler{typ: "x"},
	)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}
