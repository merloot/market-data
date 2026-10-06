package scheduling

import "context"

type Task struct {
	Name    string
	Spec    string
	Type    string
	Payload []byte
}

type Scheduler interface {
	Run(ctx context.Context) error
}