package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
)

var _ = errors.Is

func (d *Deps) handleDebugMarketData(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	symbol := r.URL.Query().Get("symbol")
	id := r.URL.Query().Get("id")

	if provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}

	c := market.Currency{
		CoinName: provider + ":" + firstNonEmpty(symbol, id),
		Provider: provider,
	}

	switch provider {
	case market.ProviderCoinGecko:
		if symbol == "" {
			writeError(w, http.StatusBadRequest, "symbol is required for coingecko")
			return
		}
		c.CoinGecko = symbol
	case market.ProviderCoinmarketcap:
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required for coinmarketcap")
			return
		}
		c.CoinMarketCap = id
	default:
		writeError(w, http.StatusBadRequest, "unknown provider: "+provider)
		return
	}

	got, err := d.MarketService.GetMarketDataList(r.Context(), []market.Currency{c})
	if err != nil {
		slog.Error("debug market-data", "err", err, "provider", provider)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"provider": provider,
		"currency": c,
		"data":     got,
	})
}

func (d *Deps) handleDebugChart(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	symbol := r.URL.Query().Get("symbol")
	id := r.URL.Query().Get("id")
	daysStr := r.URL.Query().Get("days")

	if provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}

	days := 1
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "days must be an integer")
			return
		}
	}

	c := market.Currency{
		CoinName: provider + ":" + firstNonEmpty(symbol, id),
		Provider: provider,
	}

	switch provider {
	case market.ProviderCoinGecko:
		if symbol == "" {
			writeError(w, http.StatusBadRequest, "symbol is required for coingecko")
			return
		}
		c.CoinGecko = symbol
	case market.ProviderCoinmarketcap:
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required for coinmarketcap")
			return
		}
		c.CoinMarketCap = id
	default:
		writeError(w, http.StatusBadRequest, "unknown provider: "+provider)
		return
	}

	start := time.Now()
	chart, err := d.MarketService.GetMarketDataChart(r.Context(), c, days)
	duration := time.Since(start)

	if err != nil {
		slog.Error("debug chart", "err", err, "provider", provider, "duration", duration, "error", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"provider":    provider,
		"currency":    c,
		"days":        days,
		"duration_ms": duration.Milliseconds(),
		"points":      len(chart.Prices),
		"chart":       chart,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
