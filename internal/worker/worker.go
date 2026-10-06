package worker

import "context"

type Handler interface {
	Type() string
	Handle(ctx context.Context, payload []byte) error
}

type Worker interface {
	Run(ctx context.Context) error
}
