package market_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/merloot/market-data/internal/market"
)

func TestCoinGecko_GetMarketDataList(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/coins/markets" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("ids"); got != "btc,eth" {
			t.Errorf("ids = %q, want btc, eth", got)
		}
		if got := r.URL.Query().Get("vs_currency"); got != "usd" {
			t.Errorf("vs_currency = %q, want usd", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
		"id":"bitcoin",
		"symbol":"btc",
		"current_price": 50000, 
		"market_cap":1e12,
		"circulating_supply": 12,
		"total_supply": 13,
		"last_updated": "2026-01-01T00:00:00Z"
		},
		{
		"id":"ethereum",
		"symbol":"eth",
		"current_price": 500, 
		"market_cap":1e13,
		"circulating_supply": 13,
		"total_supply": 14,
		"last_updated": "2026-01-01T00:00:00Z"
		}
		]`))
	}))
	defer srv.Close()

	p := market.NewCoinGeckoWithBaseUrl("", srv.URL)

	got, err := p.GetMarketDataList(context.Background(), []string{"btc", "eth"})
	if err != nil {
		t.Fatalf("Error =%v", err)
	}

	if len(got) != 2 {
		t.Errorf("god %d, want 2 ", len(got))
	}
	if got["btc"].CurrentPrice != 50000 {
		t.Errorf("BTC price = %v, want 50000 ", got["btc"].CurrentPrice)
	}

	if got["eth"].CurrentPrice != 500 {
		t.Errorf("ETH price = %v, want 500 ", got["eth"].CurrentPrice)
	}
}

func TestCoinGecko_RateLimited(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := market.NewCoinGeckoWithBaseUrl("", srv.URL)
	_, err := p.GetMarketDataList(context.Background(), []string{"btc"})

	if !errors.Is(err, market.ErrRateLimited) {
		t.Fatalf("Error =%v want ErrRateLimited", err)
	}
}

func TestCoinGecko_HTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := market.NewCoinGeckoWithBaseUrl("", srv.URL)
	_, err := p.GetMarketDataList(context.Background(),[]string{"btc"})

	if !errors.Is(err, market.ErrUpstream) {
		t.Fatalf("Error = %v,want ErrUpstream", err)
	}
}
