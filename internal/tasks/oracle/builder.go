package oracle

import (
	"context"
	"encoding/json"

	"github.com/merloot/market-data/internal/scheduling"
)

type Builder struct{}

func NewBuild() *Builder { return &Builder{} }

func (b *Builder) Name() string { return "oracle" }

func (b *Builder) Build(_ context.Context) ([]scheduling.Task, error) {
	payload, err := json.Marshal(Payload{})
	if err != nil {
		return nil, err
	}

	return []scheduling.Task{{
		Name:    "oracle:run",
		Spec:    Spec,
		Type:    Type,
		Payload: payload,
	}}, nil
}
