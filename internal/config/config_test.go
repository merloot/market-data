package config_test

import (
	"testing"

	"github.com/merloot/market-data/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DB_URL", "postgres://localhost/test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Errorf("HTTP.Addr = %q", cfg.HTTP.Addr)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("MaxConns = %d", cfg.Database.MaxConns)
	}
}

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DB_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}
