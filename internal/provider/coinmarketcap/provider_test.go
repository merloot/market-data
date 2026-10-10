package coinmarketcap_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/marketservice"
	"github.com/merloot/market-data/internal/provider/coinmarketcap"
)

var _ marketservice.Provider = (*coinmarketcap.Provider)(nil)

func TestProvider_GetMarketDataList(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/cryptocurrency/quotes/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "BTC,ETH" {
			t.Errorf("id = %q, want BTC,ETH", got)
		}
		if got := r.Header.Get("X-CMC_PRO_API_KEY"); got != "test-key" {
			t.Errorf("api key = %q, want test-key", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
            "data": {
                "BTC": {
                    "id": "1",
                    "circulating_supply": 19000000,
                    "total_supply": 21000000,
                    "last_updated": "2024-01-01T00:00:00Z",
                    "quote": {
                        "USD": {
                            "price": 50000,
                            "market_cap": 1000000000000,
                            "last_updated": "2024-01-01T00:00:00Z"
                        }
                    }
                }
            }
        }`))
	}))
	defer srv.Close()

	p := coinmarketcap.NewWithBaseURL("test-key", srv.URL)

	got, err := p.GetMarketDataList(context.Background(), []string{"BTC", "ETH"})
	if err != nil {
		t.Fatalf("Err = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Got %d items, want 1", len(got))
	}
	if got["BTC"].CurrentPrice != 50000 {
		t.Errorf("Price = %v, want ", got["BTC"].CurrentPrice)
	}

	if got["BTC"].CirculatingSupply != 19000000 {
		t.Errorf("Price = %v, want ", got["BTC"].CirculatingSupply)
	}
}

func TestProvider_GetMarketDataList_EmptyCurrencies(t *testing.T) {
	t.Parallel()

	p := coinmarketcap.NewWithBaseURL("test-key", "http://unused")
	got, err := p.GetMarketDataList(context.Background(), nil)
	if err != nil {
		t.Fatalf("Err = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Got %d items, want 0", len(got))
	}
}

func TestProvider_GetMarketDataList_RateLimited(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := coinmarketcap.NewWithBaseURL("test-key", srv.URL)
	_, err := p.GetMarketDataList(context.Background(), []string{"BTC"})
	if !errors.Is(err, market.ErrRateLimited) {
		t.Fatalf("Err = %v, want ErrRateLimit", err)
	}
}

func TestProvider_GetMarketDataList_HTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := coinmarketcap.NewWithBaseURL("test-key", srv.URL)
	_, err := p.GetMarketDataList(context.Background(), []string{"BTC"})
	if !errors.Is(err, market.ErrUpstream) {
		t.Fatalf("Err = %v, want ErrUpstream", err)
	}
}

func TestProvider_GetMarketDataChart(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/cryptocurrency/quotes/historical" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("interval"); got != "5m" {
			t.Errorf("Interval = %q, want 5m", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
            "data": {
                "quotes": [
                    {
                        "timestamp": "2024-01-01T00:00:00Z",
                        "quote": {
                            "USD": {
                                "price": 50000,
                                "market_cap": 1000000000000,
                                "total_supply": 21000000,
                                "circulating_supply": 19000000,
                                "timestamp": "2024-01-01T00:00:00Z"
                            }
                        }
                    }
                ]
            }
        }`))
	}))
	defer srv.Close()

	p := coinmarketcap.NewWithBaseURL("test-key", srv.URL)

	chart, err := p.GetMarketDataChart(context.Background(), market.Currency{CoinName: "BTC", Provider: "coinmarketcap", CoinMarketCap: "1"}, 1)
	if err != nil {
		t.Errorf("Err = %v", err)
	}
	if len(chart.Prices) != 1 {
		t.Fatalf("Got %d prices, want 1", len(chart.Prices))
	}

	if chart.Prices[0][1] != 50000 {
		t.Fatalf("Price = %v , want 50000", chart.Prices[0][1])
	}
}

func TestParseTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "RFC3339",
			input: "2026-01-01T00:00:00Z",
			want:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
		},
		{
			name:  "invalid",
			input: "Not a timestamp",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := coinmarketcap.ParseTimestampForTest(tt.input)
			if got != tt.want {
				t.Errorf("parseTimestamp(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
