package oracle_test

import (
	"context"
	"testing"

	"github.com/merloot/market-data/internal/tasks/oracle"
)

func TestBuilder_Build(t *testing.T) {
	t.Parallel()

	b := oracle.NewBuilder()

	tasks, err := b.Builder(context.Background())
	if err != nil {
		t.Fatalf("Error %v=", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Got %d tasks, want 1", len(tasks))
	}

	if tasks[0].Name == "" {
		t.Error("Name is empty")
	}

	if tasks[0].Spec != oracle.Spec {
		t.Errorf("Spec = %q, want %q", tasks[0].Spec, oracle.Spec)
	}

	if tasks[0].Type != oracle.Type {
		t.Errorf("Type = %q, want %q", tasks[0].Type, oracle.Type)
	}
}