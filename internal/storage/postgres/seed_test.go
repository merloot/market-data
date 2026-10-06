package postgres

import (
	"context"
	"testing"
	"time"
)

type Seeder struct {
	repo *Repo
}

func NewSeeder(t *testing.T, repo *Repo) *Seeder {
	t.Helper()
	return &Seeder{repo: repo}
}

func (s *Seeder) Currency(t *testing.T, name, logoURL string) *Seeder {
	t.Helper()

	ctx := context.Background()
	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)

	if logoURL == "" {
		_, err := s.repo.pool.Exec(ctx,
			``,
			name,
		)
		if err != nil {
			t.Fatalf("Seed currency %s: %v", name, err)
		}
	}
	_, err := s.repo.pool.Exec(ctx, ``, name, now, yesterday)
	if err != nil {
		t.Fatalf("Seed history %s: %v", name, err)
	}
	return s
}

func (s *Seeder) Truncate(t *testing.T) *Seeder {
	t.Helper()

	_, err := s.repo.pool.Exec(
		context.Background(),
		`TRUNCATE currency, market_data_history RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("Truncate: %v", err)
	}
	return s
}
