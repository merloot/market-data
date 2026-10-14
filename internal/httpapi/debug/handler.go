package debug

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/merloot/market-data/internal/domain/market"
	"github.com/merloot/market-data/internal/httpapi/httputil"
)

func getMarketData(svc MarketService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		symbol := r.URL.Query().Get("symbol")
		id := r.URL.Query().Get("id")

		if provider == "" {
			httputil.WriteError(w, http.StatusBadRequest, "provider is required")
			return
		}

		c := market.Currency{
			CoinName: provider + ":" + firstNonEmpty(symbol, id),
			Provider: provider,
		}

		switch provider {
		case market.ProviderCoinGecko:
			if symbol == "" {
				httputil.WriteError(w, http.StatusBadRequest, "symbol is required for coingecko")
				return
			}
			c.CoinGecko = symbol
		case market.ProviderCoinmarketcap:
			if id == "" {
				httputil.WriteError(w, http.StatusBadRequest, "id is required for coinmarketcap")
				return
			}
			c.CoinMarketCap = id
		default:
			httputil.WriteError(w, http.StatusBadRequest, "unknown provider: "+provider)
			return
		}

		got, err := svc.GetMarketDataList(r.Context(), []market.Currency{c})
		if err != nil {
			slog.Error("debug market-data", "err", err, "provider", provider)
			httputil.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}

		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"provider": provider,
			"currency": c,
			"data":     got,
		})
	}
}

func getChart(svc MarketService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		symbol := r.URL.Query().Get("symbol")
		id := r.URL.Query().Get("id")
		daysStr := r.URL.Query().Get("days")

		if provider == "" {
			httputil.WriteError(w, http.StatusBadRequest, "provider is required")
			return
		}

		days := 1
		if daysStr != "" {
			var err error
			days, err = strconv.Atoi(daysStr)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "days must be an integer")
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
				httputil.WriteError(w, http.StatusBadRequest, "symbol is required for coingecko")
				return
			}
			c.CoinGecko = symbol
		case market.ProviderCoinmarketcap:
			if id == "" {
				httputil.WriteError(w, http.StatusBadRequest, "id is required for coinmarketcap")
				return
			}
			c.CoinMarketCap = id
		default:
			httputil.WriteError(w, http.StatusBadRequest, "unknown provider: "+provider)
			return
		}

		start := time.Now()
		chart, err := svc.GetMarketDataChart(r.Context(), c, days)
		duration := time.Since(start)

		if err != nil {
			slog.Error("debug chart", "err", err, "provider", provider, "duration", duration, "error", err)
			httputil.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}

		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"provider":    provider,
			"currency":    c,
			"days":        days,
			"duration_ms": duration.Milliseconds(),
			"points":      len(chart.Prices),
			"chart":       chart,
		})
	}
}

func firstNonEmpty(symbol, id string) string {
	if symbol != "" {
		return symbol
	}
	return id
}
